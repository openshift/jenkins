package e2e

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"

	buildv1 "github.com/openshift/api/build/v1"
	projectv1 "github.com/openshift/api/project/v1"
	templatev1 "github.com/openshift/api/template/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilrand "k8s.io/apimachinery/pkg/util/rand"
)

var (
	activeNamespaces   = make(map[string]bool)
	activeNamespacesMu sync.Mutex
)

func registerNamespace(ns string) {
	activeNamespacesMu.Lock()
	defer activeNamespacesMu.Unlock()
	activeNamespaces[ns] = true
}

func unregisterNamespace(ns string) {
	activeNamespacesMu.Lock()
	defer activeNamespacesMu.Unlock()
	delete(activeNamespaces, ns)
}

func cleanupNamespaces() {
	activeNamespacesMu.Lock()
	defer activeNamespacesMu.Unlock()
	for ns := range activeNamespaces {
		fmt.Printf("cleaning up namespace %s\n", ns)
		_ = projectClient.ProjectV1().Projects().Delete(context.Background(), ns, metav1.DeleteOptions{})
	}
}

func registerCleanup(ta *testArgs) {
	registerNamespace(ta.ns)
	ta.t.Cleanup(func() {
		ta.t.Logf("cleaning up namespace %s", ta.ns)
		projectClient.ProjectV1().Projects().Delete(context.Background(), ta.ns, metav1.DeleteOptions{})
		unregisterNamespace(ta.ns)
	})
}

const (
	testNamespace          = "jenkins-e2e-test-namespace-"
	maxNameLength          = 63
	randomLength           = 5
	maxGeneratedNameLength = maxNameLength - randomLength
)

type testArgs struct {
	t                     *testing.T
	ns                    string
	template              string
	templateNs            string
	templateParams        map[string]string
	templateObj           *templatev1.Template
	upgradeImageStreamTag string
}

// JenkinsAuth represents the authentication credentials for Jenkins.
// It can be either basic authentication or bearer token authentication.
type jenkinsAuth struct {
	username string
	password string
	token    string
}

func basicAuth(username, password string) jenkinsAuth {
	return jenkinsAuth{username: username, password: password}
}

func bearerAuth(token string) jenkinsAuth {
	return jenkinsAuth{token: token}
}

// Apply applies the authentication credentials to the HTTP request.
func (a jenkinsAuth) apply(req *http.Request) {
	if a.token != "" {
		req.Header.Set("Authorization", "Bearer "+a.token)
	} else {
		req.SetBasicAuth(a.username, a.password)
	}
}

// GetOAuthToken gets the OAuth token from the kubeconfig.
func getOAuthToken(ta *testArgs) string {
	setupClients(ta.t)
	token := kubeConfig.BearerToken
	if token == "" {
		debugAndFailTest(ta, "no bearer token found in kubeconfig; OAuth tests require a token-based login (e.g. oc login --token=...)")
	}
	return token
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
	ta.t.Logf("test running in namespace: %s", ta.ns)

	if len(ta.template) == 0 {
		ta.template = "jenkins-ephemeral"
	}
	if len(ta.templateNs) == 0 {
		ta.templateNs = "openshift"
	}
	if ta.templateParams == nil {
		ta.templateParams = map[string]string{"MEMORY_LIMIT": "2048Mi"}
	}
	instantiateTemplate(ta)

	return ta
}

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

	watcher, err := templateClient.TemplateV1().TemplateInstances(ta.ns).Watch(context.Background(),
		metav1.SingleObject(ti.ObjectMeta),
	)
	if err != nil {
		debugAndFailTest(ta, fmt.Sprintf("%#v", err))
	}

	for event := range watcher.ResultChan() {
		switch event.Type {
		case watch.Added, watch.Modified:
			ti = event.Object.(*templatev1.TemplateInstance)

			for _, cond := range ti.Status.Conditions {
				if cond.Type == templatev1.TemplateInstanceReady &&
					cond.Status == corev1.ConditionTrue {
					ta.t.Logf("templateinstance %s/%s instantiation ready", ti.Namespace, ti.Name)
					watcher.Stop()
				}

				if cond.Type ==
					templatev1.TemplateInstanceInstantiateFailure &&
					cond.Status == corev1.ConditionTrue &&
					cond.Reason != "AlreadyExists" {
					debugAndFailTest(ta, fmt.Sprintf("templateinstance instantiation failed reason %s message %s", cond.Reason, cond.Message))
				}
			}

		case watch.Deleted:
			debugAndFailTest(ta, "templateinstance was deleted while waiting for it to be ready")

		case watch.Error:
			ta.t.Logf("watch error: %#v", spew.Sdump(event.Object))

		default:
			ta.t.Logf("unexpected event type %s", string(event.Type))
		}
	}

}

func generateName(base string) string {
	if len(base) > maxGeneratedNameLength {
		base = base[:maxGeneratedNameLength]
	}
	return fmt.Sprintf("%s%s", base, utilrand.String(randomLength))
}

func newHTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
}

func getJenkinsURL(ta *testArgs) string {
	route, err := routeClient.RouteV1().Routes(ta.ns).Get(context.Background(), "jenkins", metav1.GetOptions{})
	if err != nil {
		debugAndFailTest(ta, fmt.Sprintf("error getting jenkins route: %v", err))
	}
	return fmt.Sprintf("https://%s", route.Spec.Host)
}

func logJenkinsVersion(ta *testArgs, jenkinsURL string, auth jenkinsAuth) {
	httpClient := newHTTPClient()
	req, err := http.NewRequest("GET", jenkinsURL+"/api/json", nil)
	if err != nil {
		ta.t.Logf("failed to create jenkins version request: %v", err)
		return
	}
	auth.apply(req)

	resp, err := httpClient.Do(req)
	if err != nil {
		ta.t.Logf("failed to fetch jenkins version: %v", err)
		return
	}
	defer resp.Body.Close()

	if version := resp.Header.Get("X-Jenkins"); version != "" {
		ta.t.Logf("Jenkins version: %s", version)
	} else if version := resp.Header.Get("X-Hudson"); version != "" {
		ta.t.Logf("Jenkins version (Hudson): %s", version)
	} else {
		ta.t.Logf("Jenkins version: unknown (response status: %d, headers: %v)", resp.StatusCode, resp.Header)
	}
}

// waitForJenkinsAPI polls apiPath until Jenkins responds without a 5xx error.
func waitForJenkinsAPI(ta *testArgs, jenkinsURL, apiPath string) {
	endpoint := jenkinsURL + apiPath
	ta.t.Logf("waiting for Jenkins API at %s", endpoint)
	httpClient := newHTTPClient()
	httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	var lastErr error
	err := wait.PollUntilContextTimeout(context.Background(), 10*time.Second, 10*time.Minute, true, func(ctx context.Context) (bool, error) {
		req, reqErr := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
		if reqErr != nil {
			return false, reqErr
		}
		resp, respErr := httpClient.Do(req)
		if respErr != nil {
			lastErr = respErr
			ta.t.Logf("Jenkins not reachable yet: %v", respErr)
			return false, nil
		}
		defer resp.Body.Close()

		if resp.StatusCode < 500 {
			ta.t.Logf("Jenkins API responding at %s (status: %d)", apiPath, resp.StatusCode)
			return true, nil
		}
		lastErr = fmt.Errorf("status %d", resp.StatusCode)
		ta.t.Logf("Jenkins not ready yet at %s: %v", apiPath, lastErr)
		return false, nil
	})
	if err != nil {
		debugAndFailTest(ta, fmt.Sprintf("Jenkins API did not become available at %s: %v (last: %v)", apiPath, err, lastErr))
	}
}

func checkJenkinsLogin(ta *testArgs, jenkinsURL string, auth jenkinsAuth) {
	httpClient := newHTTPClient()
	req, err := http.NewRequest("GET", jenkinsURL+"/api/json", nil)
	if err != nil {
		debugAndFailTest(ta, fmt.Sprintf("failed to create login request: %v", err))
	}
	auth.apply(req)

	resp, err := httpClient.Do(req)
	if err != nil {
		debugAndFailTest(ta, fmt.Sprintf("login request failed: %v", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		debugAndFailTest(ta, fmt.Sprintf("login failed: status %d, body: %s", resp.StatusCode, string(body)))
	}
	ta.t.Logf("successfully authenticated to Jenkins")
}

func changeJenkinsPassword(ta *testArgs, jenkinsURL, targetUser, newPass string, auth jenkinsAuth) {
	groovyScript := fmt.Sprintf(`
import hudson.model.User
import hudson.security.HudsonPrivateSecurityRealm

def user = User.getById("%s", false)
if (user == null) {
    throw new RuntimeException("user %s not found")
}
user.addProperty(HudsonPrivateSecurityRealm.Details.fromPlainPassword("%s"))
user.save()
println("password changed successfully")
`, targetUser, targetUser, newPass)

	httpClient := newHTTPClient()

	crumbURL := jenkinsURL + "/crumbIssuer/api/json"
	crumbReq, err := http.NewRequest("GET", crumbURL, nil)
	if err != nil {
		debugAndFailTest(ta, fmt.Sprintf("error creating crumb request: %v", err))
	}
	auth.apply(crumbReq)
	crumbResp, err := httpClient.Do(crumbReq)
	if err != nil {
		debugAndFailTest(ta, fmt.Sprintf("error fetching crumb: %v", err))
	}
	defer crumbResp.Body.Close()

	crumbBody, err := io.ReadAll(crumbResp.Body)
	if err != nil {
		debugAndFailTest(ta, fmt.Sprintf("error reading crumb response: %v", err))
	}

	var crumbHeader, crumbValue string
	if crumbResp.StatusCode == http.StatusOK {
		body := string(crumbBody)
		if idx := strings.Index(body, `"crumbRequestField":"`); idx >= 0 {
			start := idx + len(`"crumbRequestField":"`)
			end := strings.Index(body[start:], `"`)
			crumbHeader = body[start : start+end]
		}
		if idx := strings.Index(body, `"crumb":"`); idx >= 0 {
			start := idx + len(`"crumb":"`)
			end := strings.Index(body[start:], `"`)
			crumbValue = body[start : start+end]
		}
	}

	form := url.Values{"script": {groovyScript}}
	scriptReq, err := http.NewRequest("POST", jenkinsURL+"/scriptText", strings.NewReader(form.Encode()))
	if err != nil {
		debugAndFailTest(ta, fmt.Sprintf("error creating script request: %v", err))
	}
	auth.apply(scriptReq)
	scriptReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if crumbHeader != "" && crumbValue != "" {
		scriptReq.Header.Set(crumbHeader, crumbValue)
	}

	scriptResp, err := httpClient.Do(scriptReq)
	if err != nil {
		debugAndFailTest(ta, fmt.Sprintf("error executing password change script: %v", err))
	}
	defer scriptResp.Body.Close()

	respBody, _ := io.ReadAll(scriptResp.Body)
	if scriptResp.StatusCode != http.StatusOK {
		debugAndFailTest(ta, fmt.Sprintf("password change script failed (status %d): %s", scriptResp.StatusCode, string(respBody)))
	}
	ta.t.Logf("password change response: %s", strings.TrimSpace(string(respBody)))
}

func upgradeJenkinsImage(ta *testArgs) {
	if ta.upgradeImageStreamTag == "" {
		debugAndFailTest(ta, "upgradeImageStreamTag is not set on testArgs")
	}
	ta.t.Logf("upgrading Jenkins DeploymentConfig to image stream tag: %s", ta.upgradeImageStreamTag)

	parts := strings.SplitN(ta.upgradeImageStreamTag, ":", 2)
	if len(parts) != 2 {
		debugAndFailTest(ta, fmt.Sprintf("invalid image stream tag format %q, expected name:tag", ta.upgradeImageStreamTag))
	}
	isName, isTag := parts[0], parts[1]

	dc, err := appClient.AppsV1().DeploymentConfigs(ta.ns).Get(context.Background(), "jenkins", metav1.GetOptions{})
	if err != nil {
		debugAndFailTest(ta, fmt.Sprintf("error getting jenkins DeploymentConfig: %v", err))
	}

	previousGeneration := dc.Status.ObservedGeneration

	triggerUpdated := false
	for i := range dc.Spec.Triggers {
		if dc.Spec.Triggers[i].Type == "ImageChange" && dc.Spec.Triggers[i].ImageChangeParams != nil {
			dc.Spec.Triggers[i].ImageChangeParams.From.Name = isName + ":" + isTag
			dc.Spec.Triggers[i].ImageChangeParams.From.Namespace = "openshift"
			triggerUpdated = true
			ta.t.Logf("updated ImageChange trigger to %s:%s", isName, isTag)
		}
	}
	if !triggerUpdated {
		debugAndFailTest(ta, "no ImageChange trigger found on jenkins DeploymentConfig")
	}

	_, err = appClient.AppsV1().DeploymentConfigs(ta.ns).Update(context.Background(), dc, metav1.UpdateOptions{})
	if err != nil {
		debugAndFailTest(ta, fmt.Sprintf("error updating jenkins DeploymentConfig: %v", err))
	}

	ta.t.Logf("waiting for DeploymentConfig generation to advance")
	err = wait.PollUntilContextTimeout(context.Background(), 5*time.Second, 5*time.Minute, false, func(ctx context.Context) (bool, error) {
		updatedDC, getErr := appClient.AppsV1().DeploymentConfigs(ta.ns).Get(ctx, "jenkins", metav1.GetOptions{})
		if getErr != nil {
			return false, getErr
		}
		if updatedDC.Status.ObservedGeneration > previousGeneration {
			ta.t.Logf("DeploymentConfig generation advanced: %d -> %d (latestVersion: %d)",
				previousGeneration, updatedDC.Status.ObservedGeneration, updatedDC.Status.LatestVersion)
			return true, nil
		}
		ta.t.Logf("waiting for generation to advance beyond %d (current: %d)",
			previousGeneration, updatedDC.Status.ObservedGeneration)
		return false, nil
	})
	if err != nil {
		debugAndFailTest(ta, fmt.Sprintf("DeploymentConfig generation did not advance: %v", err))
	}
}

func dumpPods(ta *testArgs) {
	podClient := kubeClient.CoreV1().Pods(ta.ns)
	podList, err := podClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		debugAndFailTest(ta, fmt.Sprintf("error list pods %v", err))
	}
	ta.t.Logf("dumpPods have %d items in list", len(podList.Items))
	ta.t.Logf("dumpPods items: %#v", podList.Items)
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
			ta.t.Logf("pod logs for container %s in pod %s:  %s", container.Name, pod.Name, string(b))
		}
	}
}

func debugAndFailTest(ta *testArgs, failMsg string) {
	dumpPods(ta)
	ta.t.Fatalf("%s", failMsg)
}

func createJenkinsPipelineBuildConfig(ta *testArgs, name, jenkinsfile string) {
	ta.t.Logf("creating BuildConfig %q with jenkinsPipelineStrategy", name)
	bc := &buildv1.BuildConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ta.ns,
		},
		Spec: buildv1.BuildConfigSpec{
			CommonSpec: buildv1.CommonSpec{
				Strategy: buildv1.BuildStrategy{
					Type: buildv1.JenkinsPipelineBuildStrategyType,
					JenkinsPipelineStrategy: &buildv1.JenkinsPipelineBuildStrategy{
						Jenkinsfile: jenkinsfile,
					},
				},
			},
		},
	}

	_, err := buildClient.BuildV1().BuildConfigs(ta.ns).Create(context.Background(), bc, metav1.CreateOptions{})
	if err != nil {
		debugAndFailTest(ta, fmt.Sprintf("error creating BuildConfig %s: %v", name, err))
	}
	ta.t.Logf("BuildConfig %q created", name)
}

func updateBuildConfigJenkinsfile(ta *testArgs, name, jenkinsfile string) {
	ta.t.Logf("updating BuildConfig %q jenkinsfile", name)
	bc, err := buildClient.BuildV1().BuildConfigs(ta.ns).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		debugAndFailTest(ta, fmt.Sprintf("error getting BuildConfig %s: %v", name, err))
	}
	bc.Spec.Strategy.JenkinsPipelineStrategy.Jenkinsfile = jenkinsfile
	_, err = buildClient.BuildV1().BuildConfigs(ta.ns).Update(context.Background(), bc, metav1.UpdateOptions{})
	if err != nil {
		debugAndFailTest(ta, fmt.Sprintf("error updating BuildConfig %s: %v", name, err))
	}
	ta.t.Logf("BuildConfig %q updated", name)
}

func startBuild(ta *testArgs, bcName string) string {
	ta.t.Logf("starting build from BuildConfig %q", bcName)
	buildReq := &buildv1.BuildRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name: bcName,
		},
	}
	build, err := buildClient.BuildV1().BuildConfigs(ta.ns).Instantiate(context.Background(), bcName, buildReq, metav1.CreateOptions{})
	if err != nil {
		debugAndFailTest(ta, fmt.Sprintf("error instantiating build for %s: %v", bcName, err))
	}
	ta.t.Logf("build %q started", build.Name)
	return build.Name
}

func waitForBuildComplete(ta *testArgs, buildName string) {
	ta.t.Logf("waiting for build %q to complete", buildName)

	err := wait.PollUntilContextTimeout(context.Background(), 10*time.Second, 10*time.Minute, true, func(ctx context.Context) (bool, error) {
		build, getErr := buildClient.BuildV1().Builds(ta.ns).Get(ctx, buildName, metav1.GetOptions{})
		if getErr != nil {
			ta.t.Logf("error fetching build %s: %v", buildName, getErr)
			return false, nil
		}

		phase := build.Status.Phase
		ta.t.Logf("build %s phase: %s", buildName, phase)

		switch phase {
		case buildv1.BuildPhaseComplete:
			return true, nil
		case buildv1.BuildPhaseFailed, buildv1.BuildPhaseError, buildv1.BuildPhaseCancelled:
			return false, fmt.Errorf("build %s finished with phase %s", buildName, phase)
		default:
			return false, nil
		}
	})
	if err != nil {
		debugAndFailTest(ta, fmt.Sprintf("build %s did not complete successfully: %v", buildName, err))
	}
}

type jenkinsBuildAnnotations struct {
	LogURL        string
	ConsoleLogURL string
	BlueOceanURL  string
	BuildURI      string
	StatusJSON    string
}

func getJenkinsBuildAnnotations(ta *testArgs, buildName string) jenkinsBuildAnnotations {
	ta.t.Logf("fetching Jenkins annotations from build %q", buildName)

	var annotations jenkinsBuildAnnotations
	err := wait.PollUntilContextTimeout(context.Background(), 10*time.Second, 3*time.Minute, true, func(ctx context.Context) (bool, error) {
		build, getErr := buildClient.BuildV1().Builds(ta.ns).Get(ctx, buildName, metav1.GetOptions{})
		if getErr != nil {
			ta.t.Logf("error fetching build %s: %v", buildName, getErr)
			return false, nil
		}
		logURL := build.Annotations[buildv1.BuildJenkinsLogURLAnnotation]
		if logURL == "" {
			ta.t.Logf("jenkins-log-url annotation not yet set on build %s", buildName)
			return false, nil
		}
		annotations = jenkinsBuildAnnotations{
			LogURL:        logURL,
			ConsoleLogURL: build.Annotations[buildv1.BuildJenkinsConsoleLogURLAnnotation],
			BlueOceanURL:  build.Annotations[buildv1.BuildJenkinsBlueOceanLogURLAnnotation],
			BuildURI:      build.Annotations[buildv1.BuildJenkinsBuildURIAnnotation],
			StatusJSON:    build.Annotations[buildv1.BuildJenkinsStatusJSONAnnotation],
		}
		return true, nil
	})
	if err != nil {
		debugAndFailTest(ta, fmt.Sprintf("jenkins annotations never appeared on build %s: %v", buildName, err))
	}
	ta.t.Logf("Jenkins annotations for %s: logURL=%s consoleLogURL=%s blueOceanURL=%s buildURI=%s",
		buildName, annotations.LogURL, annotations.ConsoleLogURL, annotations.BlueOceanURL, annotations.BuildURI)
	return annotations
}

func getBuildLog(ta *testArgs, buildName string, auth jenkinsAuth) string {
	annotations := getJenkinsBuildAnnotations(ta, buildName)
	logURL := annotations.LogURL
	ta.t.Logf("fetching Jenkins console log from %s", logURL)

	httpClient := newHTTPClient()
	var logText string
	var lastErr error
	err := wait.PollUntilContextTimeout(context.Background(), 10*time.Second, 3*time.Minute, true, func(ctx context.Context) (bool, error) {
		req, reqErr := http.NewRequestWithContext(ctx, "GET", logURL, nil)
		if reqErr != nil {
			return false, reqErr
		}
		auth.apply(req)

		resp, respErr := httpClient.Do(req)
		if respErr != nil {
			lastErr = respErr
			ta.t.Logf("Jenkins log not reachable yet for %s: %v", buildName, respErr)
			return false, nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("status %d", resp.StatusCode)
			ta.t.Logf("Jenkins log returned %d for %s, retrying", resp.StatusCode, buildName)
			return false, nil
		}

		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			lastErr = readErr
			return false, nil
		}
		logText = string(body)
		return true, nil
	})
	if err != nil {
		debugAndFailTest(ta, fmt.Sprintf("failed to fetch Jenkins log for %s: %v (last: %v)", buildName, err, lastErr))
	}
	return logText
}

func assertBuildLog(ta *testArgs, buildName string, auth jenkinsAuth, expectedText string) {
	ta.t.Logf("checking build log of %q for text %q", buildName, expectedText)
	logText := getBuildLog(ta, buildName, auth)
	if !strings.Contains(logText, expectedText) {
		debugAndFailTest(ta, fmt.Sprintf("expected text %q not found in build log for %s.\nFull log:\n%s", expectedText, buildName, logText))
	}
	ta.t.Logf("found expected text %q in build log for %q", expectedText, buildName)
}

func getInstalledPlugins(ta *testArgs, jenkinsURL string, auth jenkinsAuth) []string {
	httpClient := newHTTPClient()
	endpoint := jenkinsURL + "/pluginManager/api/json?depth=1"
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		debugAndFailTest(ta, fmt.Sprintf("error creating plugin list request: %v", err))
	}
	auth.apply(req)

	resp, err := httpClient.Do(req)
	if err != nil {
		debugAndFailTest(ta, fmt.Sprintf("error fetching plugin list: %v", err))
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		debugAndFailTest(ta, fmt.Sprintf("error reading plugin list response: %v", err))
	}
	if resp.StatusCode != http.StatusOK {
		debugAndFailTest(ta, fmt.Sprintf("plugin list request returned status %d: %s", resp.StatusCode, string(body)))
	}

	var result struct {
		Plugins []struct {
			ShortName string `json:"shortName"`
		} `json:"plugins"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		debugAndFailTest(ta, fmt.Sprintf("error parsing plugin list JSON: %v", err))
	}

	plugins := make([]string, len(result.Plugins))
	for i, p := range result.Plugins {
		plugins[i] = p.ShortName
	}
	return plugins
}

func checkPluginsInstalled(ta *testArgs, jenkinsURL string, auth jenkinsAuth, expected []string) {
	installed := getInstalledPlugins(ta, jenkinsURL, auth)
	ta.t.Logf("found %d installed plugins", len(installed))

	installedSet := make(map[string]bool, len(installed))
	for _, p := range installed {
		installedSet[p] = true
	}
	for _, exp := range expected {
		if !installedSet[exp] {
			debugAndFailTest(ta, fmt.Sprintf("expected plugin %q not found in installed plugins: %v", exp, installed))
		}
	}
	ta.t.Logf("all expected plugins are installed")
}

func createJenkinsJob(ta *testArgs, jenkinsURL, jobName string, auth jenkinsAuth, jobXML string) int {
	httpClient := newHTTPClient()
	endpoint := jenkinsURL + "/createItem?name=" + url.QueryEscape(jobName)
	req, err := http.NewRequest("POST", endpoint, strings.NewReader(jobXML))
	if err != nil {
		debugAndFailTest(ta, fmt.Sprintf("error creating job creation request: %v", err))
	}
	auth.apply(req)
	req.Header.Set("Content-Type", "application/xml")

	resp, err := httpClient.Do(req)
	if err != nil {
		debugAndFailTest(ta, fmt.Sprintf("error sending job creation request: %v", err))
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	ta.t.Logf("createItem %q returned status %d", jobName, resp.StatusCode)
	return resp.StatusCode
}

func getJenkinsJob(ta *testArgs, jenkinsURL, jobName string, auth jenkinsAuth) int {
	httpClient := newHTTPClient()
	endpoint := jenkinsURL + "/job/" + url.PathEscape(jobName) + "/api/json"
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		debugAndFailTest(ta, fmt.Sprintf("error creating job get request: %v", err))
	}
	auth.apply(req)

	resp, err := httpClient.Do(req)
	if err != nil {
		debugAndFailTest(ta, fmt.Sprintf("error fetching job %q: %v", jobName, err))
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	ta.t.Logf("GET job %q returned status %d", jobName, resp.StatusCode)
	return resp.StatusCode
}
