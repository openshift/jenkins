package e2e

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"testing"
	"time"

	buildv1 "github.com/openshift/api/build/v1"
	imagev1 "github.com/openshift/api/image/v1"
	projectv1 "github.com/openshift/api/project/v1"
	templatev1 "github.com/openshift/api/template/v1"

	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilrand "k8s.io/apimachinery/pkg/util/rand"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
)

const (
	testNamespace   = "jenkins-sync-plugin-test-namespace-"
	finishedSuccess = "Finished: SUCCESS"
)

func instantiateTemplate(ta *testArgs) {
	template := ta.templateObj
	var err error
	if template == nil {
		template, err = templateClient.TemplateV1().Templates(ta.templateNs).Get(context.Background(),
			ta.template, metav1.GetOptions{})
		if err != nil {
			debugAndFailTest(ta, fmt.Sprintf("%#v", err))
		}
	} else {
		template, err = templateClient.TemplateV1().Templates(ta.templateNs).Create(context.Background(), template, metav1.CreateOptions{})
		if err != nil {
			debugAndFailTest(ta, fmt.Sprintf("%#v", err))
		}
	}

	// INSTANTIATE THE TEMPLATE.

	// To set Template parameters, create a Secret holding overridden parameters
	// and their values.
	var secret *corev1.Secret
	if ta.templateParams != nil {
		secret, err = kubeClient.CoreV1().Secrets(ta.ns).Create(context.Background(), &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: ta.template,
			},
			StringData: ta.templateParams,
		}, metav1.CreateOptions{})
		if err != nil {
			debugAndFailTest(ta, fmt.Sprintf("%#v", err))
		}
	}

	// Create a TemplateInstance object, linking the Template and a reference to
	// the Secret object created above.
	ti := &templatev1.TemplateInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name: ta.template,
		},
		Spec: templatev1.TemplateInstanceSpec{
			Template: *template,
		},
	}
	if secret != nil {
		ti.Spec.Secret = &corev1.LocalObjectReference{
			Name: secret.Name,
		}
	}
	ti, err = templateClient.TemplateV1().TemplateInstances(ta.ns).Create(context.Background(),
		ti, metav1.CreateOptions{})
	if err != nil {
		debugAndFailTest(ta, fmt.Sprintf("%#v", err))
	}

	// Watch the TemplateInstance object until it indicates the Ready or
	// InstantiateFailure status condition.
	watcher, err := templateClient.TemplateV1().TemplateInstances(ta.ns).Watch(context.Background(),
		metav1.SingleObject(ti.ObjectMeta),
	)
	if err != nil {
		debugAndFailTest(ta, fmt.Sprintf("%#v", err))
	}

	for event := range watcher.ResultChan() {
		switch event.Type {
		case watch.Modified:
			ti = event.Object.(*templatev1.TemplateInstance)

			for _, cond := range ti.Status.Conditions {
				// If the TemplateInstance contains a status condition
				// Ready == True, stop watching.
				if cond.Type == templatev1.TemplateInstanceReady &&
					cond.Status == corev1.ConditionTrue {
					ta.t.Logf("templateinstance %s/%s instantiation ready", ti.Namespace, ti.Name)
					watcher.Stop()
				}

				// If the TemplateInstance contains a status condition
				// InstantiateFailure == True, indicate failure.
				if cond.Type ==
					templatev1.TemplateInstanceInstantiateFailure &&
					cond.Status == corev1.ConditionTrue &&
					cond.Reason != "AlreadyExists" {
					debugAndFailTest(ta, fmt.Sprintf("templateinstance instantiation failed reason %s message %s", cond.Reason, cond.Message))
				}
			}

		default:
			ta.t.Logf("unexpected event type %s: %#v", string(event.Type), event.Object)
		}
	}

}

func setupThroughJenkinsLaunch(t *testing.T, ta *testArgs) *testArgs {
	if ta == nil {
		ta = &testArgs{
			t: t,
		}
	}
	setupClients(ta.t)

	if len(ta.ns) == 0 {
		ta.ns = generateName(testNamespace)
		_, err := projectClient.ProjectV1().ProjectRequests().Create(context.Background(), &projectv1.ProjectRequest{
			ObjectMeta: metav1.ObjectMeta{Name: ta.ns},
		}, metav1.CreateOptions{})

		if err != nil {
			debugAndFailTest(ta, fmt.Sprintf("%#v", err))
		}
	}

	if len(ta.template) == 0 {
		ta.template = "jenkins-ephemeral"
	}
	if len(ta.templateNs) == 0 {
		ta.templateNs = "openshift"
	}
	if ta.templateParams == nil {
		ta.templateParams = map[string]string{"MEMORY_LIST": "2048Mi"}
	}
	instantiateTemplate(ta)

	return ta
}

func waitForBuildSuccess(ta *testArgs, build *buildv1.Build) *buildv1.Build {
	watcher, err := buildClient.BuildV1().Builds(ta.ns).Watch(context.Background(),
		metav1.SingleObject(build.ObjectMeta))
	if err != nil {
		debugAndFailTest(ta, fmt.Sprintf("%#v", err))
	}

	for event := range watcher.ResultChan() {
		switch event.Type {
		case watch.Modified:
			build = event.Object.(*buildv1.Build)

			switch build.Status.Phase {
			case buildv1.BuildPhaseComplete:
				if ta.expectFail {
					debugAndFailTest(ta, fmt.Sprintf("build %s worked when not expected", build.Name))
				}
				ta.t.Logf("build %s mark completed", build.Name)
				watcher.Stop()
			case buildv1.BuildPhaseError:
				watcher.Stop()
				if ta.expectFail {
					return build
				}
				ta.t.Logf("build error: %#v", build)
				ta.t.Log("dump job log")
				_, err := NewRef(ta.t, kubeClient, ta.ns).JobLogs(ta.ns, ta.bc.Name)
				if err != nil {
					debugAndFailTest(ta, fmt.Sprintf("error getting job logs: %s", err.Error()))
				}
				ta.t.Log("dump namespace pod logs")
				dumpPods(ta)
				debugAndFailTest(ta, "")
			case buildv1.BuildPhaseFailed:
				watcher.Stop()
				if ta.expectFail {
					return build
				}
				ta.t.Logf("build failed: %#v", build)
				ta.t.Log("dump job log")
				_, err := NewRef(ta.t, kubeClient, ta.ns).JobLogs(ta.ns, ta.bc.Name)
				if err != nil {
					debugAndFailTest(ta, fmt.Sprintf("error getting job logs: %s", err.Error()))
				}
				ta.t.Log("dump namespace pod logs")
				dumpPods(ta)
				debugAndFailTest(ta, "")
			default:
				ta.t.Logf("build phase %s", build.Status.Phase)
			}

		}
	}
	return build
}

func instantiateBuild(ta *testArgs) *buildv1.Build {
	if !ta.skipBCCreate {
		_, err := buildClient.BuildV1().BuildConfigs(ta.ns).Create(context.Background(), ta.bc, metav1.CreateOptions{})
		if err != nil {
			debugAndFailTest(ta, fmt.Sprintf("%#v", err))
		}
	}
	buildReq := &buildv1.BuildRequest{
		ObjectMeta: metav1.ObjectMeta{Name: ta.bc.Name},
		Env:        ta.env,
	}
	build, err := buildClient.BuildV1().BuildConfigs(ta.ns).Instantiate(context.Background(),
		ta.bc.Name, buildReq, metav1.CreateOptions{})
	if err != nil {
		debugAndFailTest(ta, fmt.Sprintf("%#v", err))
	}
	if ta.returnBeforeBuildDone {
		return build
	}
	return waitForBuildSuccess(ta, build)
}

func podTemplateTest(podTemplateName, pipelineName string, ta *testArgs) {
	bc := &buildv1.BuildConfig{}
	bc.Name = strings.ReplaceAll(podTemplateName, ":", ".")
	pipelineDefinition := strings.ReplaceAll(pipelineName, "POD_TEMPLATE_NAME", podTemplateName)
	bc.Spec = buildv1.BuildConfigSpec{
		CommonSpec: buildv1.CommonSpec{
			Strategy: buildv1.BuildStrategy{
				Type: buildv1.JenkinsPipelineBuildStrategyType,
				JenkinsPipelineStrategy: &buildv1.JenkinsPipelineBuildStrategy{
					Jenkinsfile: pipelineDefinition,
				},
			},
		},
	}

	ta.bc = bc

	ta.jobLogSearch = finishedSuccess
	basicPipelineInvocationAndValidation(ta)

}

func javaBuilderPodTemplateTest(name string, ta *testArgs) {
	bc := &buildv1.BuildConfig{}
	bc.Name = name
	bc.Spec = buildv1.BuildConfigSpec{
		CommonSpec: buildv1.CommonSpec{
			Strategy: buildv1.BuildStrategy{
				Type: buildv1.JenkinsPipelineBuildStrategyType,
				JenkinsPipelineStrategy: &buildv1.JenkinsPipelineBuildStrategy{
					Jenkinsfile: javabuilder,
				},
			},
		},
	}

	ta.bc = bc

	ta.jobLogSearch = finishedSuccess
	basicPipelineInvocationAndValidation(ta)
}

func nodejsBuilderPodTemplateTest(name string, ta *testArgs) {
	bc := &buildv1.BuildConfig{}
	bc.Name = name
	bc.Spec = buildv1.BuildConfigSpec{
		CommonSpec: buildv1.CommonSpec{
			Strategy: buildv1.BuildStrategy{
				Type: buildv1.JenkinsPipelineBuildStrategyType,
				JenkinsPipelineStrategy: &buildv1.JenkinsPipelineBuildStrategy{
					Jenkinsfile: nodejsbuilder,
				},
			},
		},
	}

	ta.bc = bc

	ta.jobLogSearch = finishedSuccess
	basicPipelineInvocationAndValidation(ta)
}

func dumpPods(ta *testArgs) {
	podClient := kubeClient.CoreV1().Pods(ta.ns)
	podList, err := podClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		debugAndFailTest(ta, fmt.Sprintf("error list pods %v", err))
	}
	ta.t.Logf("dumpPods have %d items in list", len(podList.Items))
	for _, pod := range podList.Items {
		ta.t.Logf("dumpPods looking at pod %s in phase %s", pod.Name, pod.Status.Phase)

		for _, container := range pod.Spec.Containers {
			req := podClient.GetLogs(pod.Name, &corev1.PodLogOptions{Container: container.Name})
			readCloser, err := req.Stream(context.TODO())
			if err != nil {
				debugAndFailTest(ta, fmt.Sprintf("error getting pod logs for container %s: %s", container.Name, err.Error()))
			}
			b, err := io.ReadAll(readCloser)
			if err != nil {
				debugAndFailTest(ta, fmt.Sprintf("error reading pod stream %s", err.Error()))
			}
			podLog := string(b)
			ta.t.Logf("pod logs for container %s in pod %s:  %s", container.Name, pod.Name, podLog)

		}

	}
}

func debugAndFailTest(ta *testArgs, failMsg string) {
	dumpPods(ta)
	ta.t.Fatalf(failMsg)
}

func ensureBuildDeleted(buildName string, ta *testArgs) {
	err := wait.PollImmediate(5*time.Second, 30*time.Second, func() (done bool, err error) {
		_, err = buildClient.BuildV1().Builds(ta.ns).Get(context.Background(), buildName, metav1.GetOptions{})
		if err != nil && kerrors.IsNotFound(err) {
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		debugAndFailTest(ta, fmt.Sprintf("deleted build %s still present", buildName))
	}
}

func checkPodsForText(podName, searchItem string, ta *testArgs) bool {
	found := false
	podClient := kubeClient.CoreV1().Pods(ta.ns)
	pod, err := podClient.Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		debugAndFailTest(ta, fmt.Sprintf("error list pods %v", err))
	}
	ta.t.Logf("checkPodsForText looking at pod %s in phase %s", pod.Name, pod.Status.Phase)

	for _, container := range pod.Spec.Containers {
		// Retry getting the logs since the container might not be up yet
		err := wait.PollImmediate(30*time.Second, 5*time.Minute, func() (done bool, err error) {
			req := podClient.GetLogs(pod.Name, &corev1.PodLogOptions{Container: container.Name})
			readCloser, err := req.Stream(context.TODO())
			if err != nil {
				ta.t.Logf("error getting pod logs for container %q: %q", container.Name, err.Error())
				return false, nil
			}
			b, err := io.ReadAll(readCloser)
			if err != nil {
				ta.t.Logf("error reading pod stream %s", err.Error())
				return false, nil
			}
			podLog := string(b)
			ta.t.Logf("pod logs for container %s in pod %s:  %s", container.Name, pod.Name, podLog)
			if strings.Contains(podLog, searchItem) {
				found = true
			}

			return true, nil
		})
		if err != nil {
			debugAndFailTest(ta, fmt.Sprintf("unexpected results for %s", searchItem))
		}
	}

	return found
}

func rawURICheck(rawURI string, ta *testArgs, query ...string) {
	// made this 10 minutes to line up with sync plugin relist interval
	err := wait.PollImmediate(30*time.Second, 10*time.Minute, func() (done bool, err error) {
		j := NewRef(ta.t, kubeClient, ta.ns)
		defer j.DelRawPod()
		podName, err := j.RawURL(rawURI)
		if err != nil {
			ta.t.Logf("got error %s on output for %s", err.Error(), rawURI)
			return false, nil
		}
		found := true
		for _, q := range query {
			found = found && checkPodsForText(podName, q, ta)
		}
		if !found {
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		debugAndFailTest(ta, fmt.Sprintf("unexpected results for %s", rawURI))
	}
}

func uriPost(rawURI string, ta *testArgs) {
	err := wait.PollImmediate(5*time.Second, 1*time.Minute, func() (done bool, err error) {
		j := NewRef(ta.t, kubeClient, ta.ns)
		defer j.DelRawPostPod()
		podName, err := j.RawPost(rawURI)
		if err != nil {
			ta.t.Logf("raw post %s err: %s", rawURI, err.Error())
		}
		if checkPodsForText(podName, rawURI, ta) {
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		debugAndFailTest(ta, fmt.Sprintf("unexpected post results %s", rawURI))
	}
}

func jobLogCheck(name string, ta *testArgs, query ...string) {
	err := wait.PollImmediate(5*time.Second, 1*time.Minute, func() (done bool, err error) {
		j := NewRef(ta.t, kubeClient, ta.ns)
		defer j.DelJobPod()
		podName, err := j.JobLogs(ta.ns, name)
		if err != nil {
			ta.t.Logf("got error %s on job logs for bc %s", err.Error(), name)
			return false, nil
		}
		found := true
		for _, q := range query {
			found = found && checkPodsForText(podName, q, ta)
		}
		if !found {
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		debugAndFailTest(ta, fmt.Sprintf("jenkins job for bc %s still present", name))
	}
}

func credCheck(name string, ta *testArgs, query ...string) {
	err := wait.PollImmediate(5*time.Second, 1*time.Minute, func() (done bool, err error) {
		j := NewRef(ta.t, kubeClient, ta.ns)
		defer j.DelCredentialPod()
		podName, err := j.Credential(ta.ns, name)
		if err != nil {
			ta.t.Logf("got error %s on job logs for secret %s", err.Error(), name)
			return false, nil
		}
		found := true
		for _, q := range query {
			found = found && checkPodsForText(podName, q, ta)
		}
		if !found {
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		debugAndFailTest(ta, fmt.Sprintf("cred for secret %s not available", name))
	}
}

func buildConfigSelector(name string) labels.Selector {
	return labels.Set{buildv1.BuildConfigLabel: labelValue(name)}.AsSelector()
}

func labelValue(name string) string {
	if len(name) <= validation.DNS1123LabelMaxLength {
		return name
	}
	return name[:validation.DNS1123LabelMaxLength]
}

func isPruningDone(ta *testArgs) {
	var builds *buildv1.BuildList
	err := wait.PollImmediate(5*time.Second, 1*time.Minute, func() (done bool, err error) {
		builds, err = buildClient.BuildV1().Builds(ta.ns).List(context.Background(), metav1.ListOptions{LabelSelector: buildConfigSelector(ta.bc.Name).String()})
		if err != nil {
			ta.t.Logf("%s", err.Error())
			return false, nil
		}
		if ta.expectFail {
			if int32(len(builds.Items)) == *ta.bc.Spec.FailedBuildsHistoryLimit {
				return true, nil
			}
		} else {
			if int32(len(builds.Items)) == *ta.bc.Spec.SuccessfulBuildsHistoryLimit {
				return true, nil
			}
		}
		return false, nil
	})
	if err != nil {
		debugAndFailTest(ta, fmt.Sprintf("problem with build pruning for %s, num of builds: %d", ta.bc.Name, len(builds.Items)))
	}
}

func scaleJenkins(up bool, ta *testArgs) {
	replicaCount := 0
	if up {
		replicaCount = 1
	}
	dc, err := appClient.AppsV1().DeploymentConfigs(ta.ns).Get(context.Background(), "jenkins", metav1.GetOptions{})
	if err != nil {
		debugAndFailTest(ta, fmt.Sprintf("error getting dc jenkins scale: %s", err.Error()))
	}
	dc.Spec.Replicas = int32(replicaCount)
	_, err = appClient.AppsV1().DeploymentConfigs(ta.ns).Update(context.Background(), dc, metav1.UpdateOptions{})
	if err != nil {
		debugAndFailTest(ta, fmt.Sprintf("error updating dc jenkins scale to %d: %s", replicaCount, err.Error()))
	}
	err = wait.PollImmediate(5*time.Second, 1*time.Minute, func() (done bool, err error) {
		dc, err := appClient.AppsV1().DeploymentConfigs(ta.ns).Get(context.Background(), "jenkins", metav1.GetOptions{})
		if err != nil {
			ta.t.Logf("error getting dc jenkins: %s", err.Error())
		}
		if dc.Status.Replicas != int32(replicaCount) {
			ta.t.Logf("jenkins dc status still %d", dc.Status.Replicas)
			return false, nil
		}
		ta.t.Logf("jenkins dc status %d", replicaCount)
		return true, nil
	})
	if err != nil {
		debugAndFailTest(ta, "dc scale problems")
	}
}

const (
	maxNameLength          = 63
	randomLength           = 5
	maxGeneratedNameLength = maxNameLength - randomLength
)

func generateName(base string) string {
	if len(base) > maxGeneratedNameLength {
		base = base[:maxGeneratedNameLength]
	}
	return fmt.Sprintf("%s%s", base, utilrand.String(randomLength))

}

type testArgs struct {
	t                     *testing.T
	ns                    string
	template              string
	templateNs            string
	templateParams        map[string]string
	templateObj           *templatev1.Template
	bc                    *buildv1.BuildConfig
	skipBCCreate          bool
	returnBeforeBuildDone bool
	jobLogSearch          string
	env                   []corev1.EnvVar
	expectFail            bool
}

func basicPipelineInvocationAndValidation(ta *testArgs) *buildv1.Build {
	bld := instantiateBuild(ta)

	if len(ta.jobLogSearch) > 0 {
		jobLogCheck(ta.bc.Name, ta, ta.jobLogSearch)
	}
	return bld
}

func TestEnvVarOverride(t *testing.T) {
	bc := buildv1.BuildConfig{}
	bc.Name = "sync-plugin-bc-env-var-override"
	bc.Spec = buildv1.BuildConfigSpec{
		CommonSpec: buildv1.CommonSpec{
			Source: buildv1.BuildSource{
				Type: buildv1.BuildSourceGit,
				Git: &buildv1.GitBuildSource{
					//TODO need to find a home for this in an openshift related repo
					URI: "https://github.com/gabemontero/test-jenkins-bc-env-var-override",
				},
			},
			Strategy: buildv1.BuildStrategy{
				Type: buildv1.JenkinsPipelineBuildStrategyType,
				JenkinsPipelineStrategy: &buildv1.JenkinsPipelineBuildStrategy{
					Env: []corev1.EnvVar{
						{
							Name:  "FOO1",
							Value: "BAR1",
						},
					},
				},
			},
		},
	}
	ta := &testArgs{
		t:            t,
		bc:           &bc,
		jobLogSearch: "FOO1 is BAR1",
	}
	ta = setupThroughJenkinsLaunch(ta.t, ta)
	defer projectClient.ProjectV1().Projects().Delete(context.Background(), ta.ns, metav1.DeleteOptions{})

	basicPipelineInvocationAndValidation(ta)
	ta.env = []corev1.EnvVar{{Name: "FOO1", Value: "BAR2"}}
	ta.jobLogSearch = "FOO1 is BAR2"
	ta.skipBCCreate = true
	basicPipelineInvocationAndValidation(ta)
}

func TestCreateThenDeleteBC(t *testing.T) {
	bc := buildv1.BuildConfig{}
	bc.Name = "sync-plugin-create-then-delete-bc"
	bc.Spec = buildv1.BuildConfigSpec{
		CommonSpec: buildv1.CommonSpec{
			Strategy: buildv1.BuildStrategy{
				Type: buildv1.JenkinsPipelineBuildStrategyType,
				JenkinsPipelineStrategy: &buildv1.JenkinsPipelineBuildStrategy{
					Jenkinsfile: samplepipeline,
				},
			},
		},
	}

	ta := &testArgs{
		t:  t,
		bc: &bc,
	}
	setupThroughJenkinsLaunch(t, ta)
	defer projectClient.ProjectV1().Projects().Delete(context.Background(), ta.ns, metav1.DeleteOptions{})

	ta.jobLogSearch = finishedSuccess
	basicPipelineInvocationAndValidation(ta)

	err := buildClient.BuildV1().BuildConfigs(ta.ns).Delete(context.Background(), bc.Name, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("error on bc %s delete: %s", bc.Name, err.Error())
	}

	jobLogCheck(bc.Name, ta, "<body><h2>HTTP ERROR 404 Not Found</h2>")
}

func TestSecretCredentialSync(t *testing.T) {
	ta := setupThroughJenkinsLaunch(t, nil)
	defer projectClient.ProjectV1().Projects().Delete(context.Background(), ta.ns, metav1.DeleteOptions{})

	secret := &corev1.Secret{
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"password": []byte("c2VjcmV0Y3JlZHN5bmMK"),
			"username": []byte("c2VjcmV0Y3JlZHN5bmMK"),
		},
	}
	secret.Name = "secret-to-credential"
	secret.Labels = map[string]string{"credential.sync.jenkins.openshift.io": "true"}
	secret, err := kubeClient.CoreV1().Secrets(ta.ns).Create(context.Background(), secret, metav1.CreateOptions{})

	credCheck(secret.Name, ta, "secret-to-credential", "c2VjcmV0Y3JlZHN5bmMK")

	delete(secret.Labels, "credential.sync.jenkins.openshift.io")
	secret, err = kubeClient.CoreV1().Secrets(ta.ns).Update(context.Background(), secret, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error updating secret: %s", err.Error())
	}

	credCheck(secret.Name, ta, "<body><h2>HTTP ERROR 404 Not Found</h2>")

	secret.Labels = map[string]string{"credential.sync.jenkins.openshift.io": "true"}
	secret, err = kubeClient.CoreV1().Secrets(ta.ns).Update(context.Background(), secret, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error updating secret: %s", err.Error())
	}

	credCheck(secret.Name, ta, "secret-to-credential", "c2VjcmV0Y3JlZHN5bmMK")

	err = kubeClient.CoreV1().Secrets(ta.ns).Delete(context.Background(), secret.Name, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("error deleting secret %s: %s", secret.Name, err.Error())
	}

	credCheck(secret.Name, ta, "<body><h2>HTTP ERROR 404 Not Found</h2>")
}

func TestSecretCredentialSyncAfterStartup(t *testing.T) {
	ta := &testArgs{
		t: t,
	}

	setupClients(t)

	ta.ns = generateName(testNamespace)
	_, err := projectClient.ProjectV1().ProjectRequests().Create(context.Background(), &projectv1.ProjectRequest{
		ObjectMeta: metav1.ObjectMeta{Name: ta.ns},
	}, metav1.CreateOptions{})
	if err != nil {
		debugAndFailTest(ta, fmt.Sprintf("%#v", err))
	}

	secret := &corev1.Secret{
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"password": []byte("c2VjcmV0Y3JlZHN5bmMK"),
			"username": []byte("c2VjcmV0Y3JlZHN5bmMK"),
		},
	}
	secret.Name = "secret-to-credential"
	secret.Labels = map[string]string{"credential.sync.jenkins.openshift.io": "true"}
	secret, err = kubeClient.CoreV1().Secrets(ta.ns).Create(context.Background(), secret, metav1.CreateOptions{})

	setupThroughJenkinsLaunch(t, ta)
	defer projectClient.ProjectV1().Projects().Delete(context.Background(), ta.ns, metav1.DeleteOptions{})

	credCheck(secret.Name, ta, "secret-to-credential", "c2VjcmV0Y3JlZHN5bmMK")
}

func TestConfigMapPodTemplate(t *testing.T) {
	ta := setupThroughJenkinsLaunch(t, nil)
	defer projectClient.ProjectV1().Projects().Delete(context.Background(), ta.ns, metav1.DeleteOptions{})
	labels := map[string]string{"role": "jenkins-agent"}
	cmName := "config-map-with-podtemplate"
	podTemplateName := "jenkins-agent"
	cm := newPodTemplateConfigMap(cmName, podTemplateName, labels)
	cm, err := kubeClient.CoreV1().ConfigMaps(ta.ns).Create(context.Background(), cm, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error creating pod template cm: %s", err.Error())
	}
	podTemplateTest(podTemplateName, simplemaven2, ta)
}

func TestConfigMapLegacyPodTemplate(t *testing.T) {
	ta := setupThroughJenkinsLaunch(t, nil)
	defer projectClient.ProjectV1().Projects().Delete(context.Background(), ta.ns, metav1.DeleteOptions{})
	labels := map[string]string{"role": "jenkins-slave"}
	cmName := "config-map-with-legacy-podtemplate"
	podTemplateName := "jenkins-slave"
	cm := newPodTemplateConfigMap(cmName, podTemplateName, labels)
	cm, err := kubeClient.CoreV1().ConfigMaps(ta.ns).Create(context.Background(), cm, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error creating pod template cm: %s", err.Error())
	}
	podTemplateTest(podTemplateName, simplemaven2, ta)
}

func newPodTemplateConfigMap(configMapName string, podTemplateName string, templateLabels map[string]string) *corev1.ConfigMap {
	templateDefinition := `
        <org.csanchez.jenkins.plugins.kubernetes.PodTemplate>
          <inheritFrom></inheritFrom>
          <name>POD_TEMPLATE_NAME</name>
          <instanceCap>2147483647</instanceCap>
          <idleMinutes>0</idleMinutes>
          <label>POD_TEMPLATE_NAME</label>
          <serviceAccount>jenkins</serviceAccount>
          <nodeSelector></nodeSelector>
          <volumes/>
          <containers>
            <org.csanchez.jenkins.plugins.kubernetes.ContainerTemplate>
              <name>jnlp</name>
              <image>image-registry.openshift-image-registry.svc:5000/openshift/jenkins-agent-base:latest</image>
              <privileged>false</privileged>
              <alwaysPullImage>true</alwaysPullImage>
              <workingDir>/home/jenkins/agent</workingDir>
              <command></command>
              <args>\$(JENKINS_SECRET) \$(JENKINS_NAME)</args>
              <ttyEnabled>false</ttyEnabled>
              <resourceRequestCpu></resourceRequestCpu>
              <resourceRequestMemory></resourceRequestMemory>
              <resourceLimitCpu></resourceLimitCpu>
              <resourceLimitMemory></resourceLimitMemory>
              <envVars/>
            </org.csanchez.jenkins.plugins.kubernetes.ContainerTemplate>
            <org.csanchez.jenkins.plugins.kubernetes.ContainerTemplate>
              <name>java</name>
              <image>image-registry.openshift-image-registry.svc:5000/openshift/java:latest</image>
              <privileged>false</privileged>
              <alwaysPullImage>true</alwaysPullImage>
              <workingDir>/home/jenkins/agent</workingDir>
              <command>cat</command>
              <args></args>
              <ttyEnabled>true</ttyEnabled>
              <resourceRequestCpu></resourceRequestCpu>
              <resourceRequestMemory></resourceRequestMemory>
              <resourceLimitCpu></resourceLimitCpu>
              <resourceLimitMemory></resourceLimitMemory>
              <envVars/>
            </org.csanchez.jenkins.plugins.kubernetes.ContainerTemplate>
          </containers>
          <envVars/>
          <annotations/>
          <imagePullSecrets/>
          <nodeProperties/>
        </org.csanchez.jenkins.plugins.kubernetes.PodTemplate>
	`
	templateDefinition = strings.ReplaceAll(templateDefinition, "POD_TEMPLATE_NAME", podTemplateName)
	cm := &corev1.ConfigMap{Data: map[string]string{podTemplateName: templateDefinition}}
	cm.Labels = templateLabels
	cm.Name = configMapName
	return cm
}

func TestImageStreamPodTemplate(t *testing.T) {
	ta := setupThroughJenkinsLaunch(t, nil)
	defer projectClient.ProjectV1().Projects().Delete(context.Background(), ta.ns, metav1.DeleteOptions{})

	is := &imagev1.ImageStream{}
	podTemplateName := "sync-plugin-imagestream-pod-template"
	is.Name = podTemplateName
	is.Labels = map[string]string{"role": "jenkins-slave"}
	is.Spec.Tags = []imagev1.TagReference{
		{
			From: &corev1.ObjectReference{
				Kind: "DockerImage",
				Name: "registry.redhat.io/openshift4/ose-jenkins-agent-maven:v4.10",
			},
			Name: "base",
		},
		{
			From: &corev1.ObjectReference{
				Kind: "ImageStreamTag",
				Name: "base",
			},
			Name: "latest",
		},
	}

	is, err := imageClient.ImageV1().ImageStreams(ta.ns).Create(context.Background(), is, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error creating pod template stream: %s", err.Error())
	}

	podTemplateTest(podTemplateName, simplemaven1, ta)
}

func TestImageStreamTagPodTemplate(t *testing.T) {
	ta := setupThroughJenkinsLaunch(t, nil)
	defer projectClient.ProjectV1().Projects().Delete(context.Background(), ta.ns, metav1.DeleteOptions{})

	is := &imagev1.ImageStream{}
	podTemplateName := "sync-plugin-imagestreamtag-pod-template"
	podTemplateTag := "latest"
	is.Name = podTemplateName
	is.Labels = map[string]string{"role": "jenkins-slave"}
	is.Spec.Tags = []imagev1.TagReference{
		{
			From: &corev1.ObjectReference{
				Kind: "DockerImage",
				Name: "registry.redhat.io/openshift4/ose-jenkins-agent-maven:v4.10",
			},
			Name: "base",
		},
		{
			From: &corev1.ObjectReference{
				Kind: "ImageStreamTag",
				Name: "base",
			},
			Annotations: map[string]string{
				"role": "jenkins-slave",
			},
			Name: podTemplateTag,
		},
	}

	if _, err := imageClient.ImageV1().ImageStreams(ta.ns).Create(context.Background(), is, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error creating pod template stream: %s", err.Error())
	}

	podTemplateTest(podTemplateName+":"+podTemplateTag, simplemaven1, ta)
}

func TestJavaBuilderPodTemplate(t *testing.T) {
	ta := setupThroughJenkinsLaunch(t, nil)
	defer projectClient.ProjectV1().Projects().Delete(context.Background(), ta.ns, metav1.DeleteOptions{})
	javaBuilderPodTemplateTest("sync-plugin-java-builder-pod-template", ta)
}

func TestNodeJSBuilderPodTemplate(t *testing.T) {
	ta := setupThroughJenkinsLaunch(t, nil)
	defer projectClient.ProjectV1().Projects().Delete(context.Background(), ta.ns, metav1.DeleteOptions{})
	nodejsBuilderPodTemplateTest("sync-plugin-nodejs-builder-pod-template", ta)
}

func TestPruningSuccessfulPipeline(t *testing.T) {
	ta := setupThroughJenkinsLaunch(t, nil)
	defer projectClient.ProjectV1().Projects().Delete(context.Background(), ta.ns, metav1.DeleteOptions{})

	two := int32(2)
	bc := &buildv1.BuildConfig{}
	bc.Name = "sync-plugin-prune-successful-builds"
	bc.Spec = buildv1.BuildConfigSpec{
		SuccessfulBuildsHistoryLimit: &two,
		CommonSpec: buildv1.CommonSpec{
			Strategy: buildv1.BuildStrategy{
				Type: buildv1.JenkinsPipelineBuildStrategyType,
				JenkinsPipelineStrategy: &buildv1.JenkinsPipelineBuildStrategy{
					Jenkinsfile: simpleSuccessfulPipeline,
				},
			},
		},
	}

	ta.bc = bc
	ta.skipBCCreate = true
	var err error
	if ta.bc, err = buildClient.BuildV1().BuildConfigs(ta.ns).Create(context.Background(), bc, metav1.CreateOptions{}); err != nil {
		debugAndFailTest(ta, fmt.Sprintf("error creating bc: %s", err.Error()))
	}

	ta.jobLogSearch = finishedSuccess
	for i := 0; i < 4; i++ {
		basicPipelineInvocationAndValidation(ta)
	}

	isPruningDone(ta)
}

func TestPruningFailedPipeline(t *testing.T) {
	ta := setupThroughJenkinsLaunch(t, nil)
	defer projectClient.ProjectV1().Projects().Delete(context.Background(), ta.ns, metav1.DeleteOptions{})

	two := int32(2)
	bc := &buildv1.BuildConfig{}
	bc.Name = "sync-plugin-prune-failed-builds"
	bc.Spec = buildv1.BuildConfigSpec{
		FailedBuildsHistoryLimit: &two,
		CommonSpec: buildv1.CommonSpec{
			Strategy: buildv1.BuildStrategy{
				Type: buildv1.JenkinsPipelineBuildStrategyType,
				JenkinsPipelineStrategy: &buildv1.JenkinsPipelineBuildStrategy{
					Jenkinsfile: simpleFailedPipeline,
				},
			},
		},
	}

	ta.bc = bc
	ta.skipBCCreate = true
	ta.expectFail = true
	var err error
	if ta.bc, err = buildClient.BuildV1().BuildConfigs(ta.ns).Create(context.Background(), bc, metav1.CreateOptions{}); err != nil {
		debugAndFailTest(ta, fmt.Sprintf("error creating bc: %s", err.Error()))
	}

	for i := 0; i < 4; i++ {
		instantiateBuild(ta)
	}

	isPruningDone(ta)
}

func TestDeclarativePlusNodejs(t *testing.T) {
	ta := setupThroughJenkinsLaunch(t, nil)
	defer projectClient.ProjectV1().Projects().Delete(context.Background(), ta.ns, metav1.DeleteOptions{})

	bc := &buildv1.BuildConfig{}
	bc.Name = "sync-plugin-nodejs-declarative-builds"
	bc.Spec = buildv1.BuildConfigSpec{
		CommonSpec: buildv1.CommonSpec{
			Strategy: buildv1.BuildStrategy{
				Type: buildv1.JenkinsPipelineBuildStrategyType,
				JenkinsPipelineStrategy: &buildv1.JenkinsPipelineBuildStrategy{
					Jenkinsfile: fmt.Sprintf(nodejsDeclarative, ta.ns),
				},
			},
		},
	}

	ta.bc = bc
	ta.skipBCCreate = true
	var err error
	if ta.bc, err = buildClient.BuildV1().BuildConfigs(ta.ns).Create(context.Background(), bc, metav1.CreateOptions{}); err != nil {
		debugAndFailTest(ta, fmt.Sprintf("error creating bc: %s", err.Error()))
	}

	ta.jobLogSearch = finishedSuccess
	basicPipelineInvocationAndValidation(ta)

	err = wait.PollImmediate(5*time.Second, 1*time.Minute, func() (done bool, err error) {
		ep, err := kubeClient.CoreV1().Endpoints(ta.ns).Get(context.Background(), "nodejs-postgresql-example", metav1.GetOptions{})
		if err != nil {
			ta.t.Logf("%s", err.Error())
			return false, nil
		}
		if len(ep.Subsets) == 0 || len(ep.Subsets[0].Addresses) == 0 {
			ta.t.Logf("endpoint %s not ready", ep.Name)
			return false, nil
		}
		return true, nil
	})

}

func TestDeletedBuildDeletesRun(t *testing.T) {
	ta := setupThroughJenkinsLaunch(t, nil)
	defer projectClient.ProjectV1().Projects().Delete(context.Background(), ta.ns, metav1.DeleteOptions{})

	bc := &buildv1.BuildConfig{}
	bc.Name = "sync-plugin-pipeline-with-envs"
	bc.Spec = buildv1.BuildConfigSpec{
		CommonSpec: buildv1.CommonSpec{
			Strategy: buildv1.BuildStrategy{
				Type: buildv1.JenkinsPipelineBuildStrategyType,
				JenkinsPipelineStrategy: &buildv1.JenkinsPipelineBuildStrategy{
					Env: []corev1.EnvVar{
						{
							Name:  "FOO1",
							Value: "BAR1",
						},
					},
					Jenkinsfile: pipelineWithEnvs,
				},
			},
		},
	}

	ta.bc = bc
	ta.skipBCCreate = true
	var err error
	if ta.bc, err = buildClient.BuildV1().BuildConfigs(ta.ns).Create(context.Background(), bc, metav1.CreateOptions{}); err != nil {
		debugAndFailTest(ta, fmt.Sprintf("error creating bc: %s", err.Error()))
	}

	type buildInfo struct {
		number          int
		jenkinsBuildURI string
	}
	buildNameToBuildInfoMap := map[string]buildInfo{}

	ta.jobLogSearch = finishedSuccess
	for i := 1; i <= 5; i++ {
		bld := basicPipelineInvocationAndValidation(ta)

		jenkinsBuildURI, err := url.Parse(bld.Annotations[buildv1.BuildJenkinsBuildURIAnnotation])
		if err != nil {
			debugAndFailTest(ta, fmt.Sprintf("error with build uri annotation for build %s: %s", bld.Name, err.Error()))
		}
		buildNameToBuildInfoMap[bld.Name] = buildInfo{number: i, jenkinsBuildURI: jenkinsBuildURI.Path}
	}

	for buildName, buildInfo := range buildNameToBuildInfoMap {
		if buildInfo.number%2 == 0 {
			err = buildClient.BuildV1().Builds(ta.ns).Delete(context.Background(), buildName, metav1.DeleteOptions{})
			if err != nil {
				debugAndFailTest(ta, fmt.Sprintf("error deleting %s: %s", buildName, err.Error()))
			}
			ta.t.Logf("build %s deleted at time %s", buildName, time.Now().String())
			ensureBuildDeleted(buildName, ta)
		}
	}

	dbg := func(num int) {
		j := NewRef(ta.t, kubeClient, ta.ns)
		defer j.DelPastJobLogs()
		j.PastJobLogs(ta.ns, ta.bc.Name, num)
		ta.t.Logf("see job log for %s/%s/%d above ^^", ta.ns, ta.bc.Name, num)
	}
	for _, buildInfo := range buildNameToBuildInfoMap {
		dbg(buildInfo.number)
		if buildInfo.number%2 == 0 {
			rawURICheck(buildInfo.jenkinsBuildURI, ta, "<body><h2>HTTP ERROR 404 Not Found</h2>")
		} else {
			rawURICheck(buildInfo.jenkinsBuildURI, ta, fmt.Sprintf("Build #%d", buildInfo.number))
		}
	}

}

func TestBlueGreen(t *testing.T) {
	ta := setupThroughJenkinsLaunch(t, nil)
	defer projectClient.ProjectV1().Projects().Delete(context.Background(), ta.ns, metav1.DeleteOptions{})

	data := []byte(bluegreenTemplateYAML)
	annotationDecodingScheme := runtime.NewScheme()
	utilruntime.Must(templatev1.Install(annotationDecodingScheme))
	utilruntime.Must(templatev1.DeprecatedInstallWithoutGroup(annotationDecodingScheme))
	annotationDecoderCodecFactory := serializer.NewCodecFactory(annotationDecodingScheme)
	decoder := annotationDecoderCodecFactory.UniversalDecoder(templatev1.GroupVersion)
	bgt := &templatev1.Template{}
	err := runtime.DecodeInto(decoder, data, bgt)
	if err != nil {
		t.Fatalf("err creating template from yaml: %s", err.Error())
	}

	ta.template = bgt.Name
	ta.templateNs = ta.ns
	ta.templateObj = bgt
	ta.templateParams = map[string]string{"VERBOSE": "true", "APPLICATION_DOMAIN": fmt.Sprintf("nodejs-%s.ocp.io", ta.ns)}
	instantiateTemplate(ta)

	ta.skipBCCreate = true
	ta.returnBeforeBuildDone = true
	ta.bc = &buildv1.BuildConfig{}
	ta.bc.Name = "bluegreen-pipeline"
	buildAndSwitch := func(newColor string) {
		b := instantiateBuild(ta)

		jenkinsBuildURI := b.Annotations[buildv1.BuildJenkinsBuildURIAnnotation]
		if len(jenkinsBuildURI) == 0 {
			err = wait.PollImmediate(5*time.Second, 1*time.Minute, func() (done bool, err error) {
				b, err = buildClient.BuildV1().Builds(ta.ns).Get(context.Background(), b.Name, metav1.GetOptions{})
				if err != nil {
					t.Logf("build get error: %s", err.Error())
					return false, nil
				}
				jenkinsBuildURI = b.Annotations[buildv1.BuildJenkinsBuildURIAnnotation]
				if len(jenkinsBuildURI) > 0 {
					return true, nil
				}
				return false, nil
			})
		}

		u, err := url.Parse(jenkinsBuildURI)
		if err != nil {
			t.Fatalf("bad build uri %s: %s", jenkinsBuildURI, err.Error())
		}
		jenkinsBuildURI = strings.Trim(u.Path, "/") // trim leading https://host/ and trailing /

		rawURICheck(jenkinsBuildURI+"/consoleText", ta, "Approve?")

		uriPost(jenkinsBuildURI+"/input/Approval/proceedEmpty", ta)

		t.Logf("approval post for %s done", newColor)

		waitForBuildSuccess(ta, b)
		jobLogCheck("bluegreen-pipeline", ta, finishedSuccess)

		// verify route color
		r, err := routeClient.RouteV1().Routes(ta.ns).Get(context.Background(), "nodejs-postgresql-example", metav1.GetOptions{})
		if err != nil {
			t.Fatalf("error on route get: %s", err.Error())
		}
		activeRoute := strings.TrimSpace(r.Spec.To.Name)
		if activeRoute != fmt.Sprintf("nodejs-postgresql-example-%s", newColor) {
			t.Fatalf("unexpected route value for %s: %s", newColor, activeRoute)
		}
	}

	buildAndSwitch("green")
	buildAndSwitch("blue")
}

func TestPersistentVolumes(t *testing.T) {
	ta := &testArgs{t: t}
	setupClients(ta.t)
	randomTestNamespaceName := generateName(testNamespace)
	ta.ns = randomTestNamespaceName
	_, err := projectClient.ProjectV1().ProjectRequests().Create(context.Background(), &projectv1.ProjectRequest{
		ObjectMeta: metav1.ObjectMeta{Name: randomTestNamespaceName},
	}, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("%#v", err)
	}
	defer projectClient.ProjectV1().Projects().Delete(context.Background(), ta.ns, metav1.DeleteOptions{})

	tester := NewTester(kubeClient, ta.ns, t)
	pvServerPod, pvs := tester.SetupK8SNFSServerAndVolume()
	pvCleanup := func() {
		for _, pv := range pvs {
			kubeClient.CoreV1().PersistentVolumes().Delete(context.Background(), pv.Name, metav1.DeleteOptions{})
			kubeClient.CoreV1().Pods(ta.ns).Delete(context.Background(), pvServerPod.Name, metav1.DeleteOptions{})
			kubeClient.RbacV1().ClusterRoleBindings().Delete(context.Background(), "priv-scc-binding-nfs-pvs", metav1.DeleteOptions{})
		}
	}
	defer pvCleanup()

	ta.template = "jenkins-persistent"
	ta.templateNs = "openshift"
	ta.templateParams = map[string]string{"MEMORY_LIST": "2048Mi"}
	instantiateTemplate(ta)

	bc := &buildv1.BuildConfig{}
	bc.Name = "sync-plugin-pipeline-with-envs"
	bc.Spec = buildv1.BuildConfigSpec{
		CommonSpec: buildv1.CommonSpec{
			Strategy: buildv1.BuildStrategy{
				Type: buildv1.JenkinsPipelineBuildStrategyType,
				JenkinsPipelineStrategy: &buildv1.JenkinsPipelineBuildStrategy{
					Env: []corev1.EnvVar{
						{
							Name:  "FOO1",
							Value: "BAR1",
						},
					},
					Jenkinsfile: pipelineWithEnvs,
				},
			},
		},
	}

	ta.bc = bc
	ta.skipBCCreate = true
	if ta.bc, err = buildClient.BuildV1().BuildConfigs(ta.ns).Create(context.Background(), bc, metav1.CreateOptions{}); err != nil {
		debugAndFailTest(ta, fmt.Sprintf("error creating bc: %s", err.Error()))
	}

	type buildInfo struct {
		number          int
		jenkinsBuildURI string
	}
	buildNameToBuildInfoMap := map[string]buildInfo{}

	ta.jobLogSearch = finishedSuccess
	for i := 1; i <= 5; i++ {
		bld := basicPipelineInvocationAndValidation(ta)

		jenkinsBuildURI, err := url.Parse(bld.Annotations[buildv1.BuildJenkinsBuildURIAnnotation])
		if err != nil {
			t.Fatalf("error with build uri annotation for build %s: %s", bld.Name, err.Error())
		}
		buildNameToBuildInfoMap[bld.Name] = buildInfo{number: i, jenkinsBuildURI: jenkinsBuildURI.Path}
	}

	scaleJenkins(false, ta)

	// make sure jenkins is down
	ta.t.Log("making sure jenkins is down via http get to jenkins console")
	rawURICheck("", ta, "Failed connect to")
	ta.t.Log("http get to jenkins console came back OK")

	for buildName, buildInfo := range buildNameToBuildInfoMap {
		if buildInfo.number%2 == 0 {
			err = buildClient.BuildV1().Builds(ta.ns).Delete(context.Background(), buildName, metav1.DeleteOptions{})
			if err != nil {
				t.Fatalf("error deleting %s: %s", buildName, err.Error())
			}
			ta.t.Logf("build %s deleted at time %s", buildName, time.Now().String())
			ensureBuildDeleted(buildName, ta)
		}
	}

	scaleJenkins(true, ta)

	// make sure jenkins is up
	ta.t.Log("making sure jenkins is up via http get to jenkins console")
	rawURICheck("", ta, "<title>")
	ta.t.Log("http get to jenkins console came back OK")

	dbg := func(num int) {
		j := NewRef(ta.t, kubeClient, ta.ns)
		defer j.DelPastJobLogs()
		j.PastJobLogs(ta.ns, ta.bc.Name, num)
		ta.t.Logf("see job log for %s/%s/%d above ^^", ta.ns, ta.bc.Name, num)
	}
	for _, buildInfo := range buildNameToBuildInfoMap {
		dbg(buildInfo.number)
		if buildInfo.number%2 == 0 {
			rawURICheck(buildInfo.jenkinsBuildURI, ta, "<body><h2>HTTP ERROR 404 Not Found</h2>")
		} else {
			rawURICheck(buildInfo.jenkinsBuildURI, ta, fmt.Sprintf("Build #%d", buildInfo.number))
		}
	}

}
