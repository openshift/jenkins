package e2e

import (
	"context"
	"fmt"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeset "k8s.io/client-go/kubernetes"
)

// JenkinsRef represents a Jenkins instance running on an OpenShift server
type JenkinsRef struct {
	host string
	port string
	// The namespace in which the Jenkins server is running
	namespace  string
	uri_tester *Tester
	kubeClient *kubeset.Clientset
	t          *testing.T
}

// NewRef creates a jenkins reference from an OC client
func NewRef(t *testing.T, kubeClient *kubeset.Clientset, testNamespace string) *JenkinsRef {
	svc, err := kubeClient.CoreV1().Services(testNamespace).Get(context.Background(), "jenkins", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("%#v", err)
	}
	serviceIP := svc.Spec.ClusterIP
	port := svc.Spec.Ports[0].Port

	j := &JenkinsRef{
		host:       serviceIP,
		port:       fmt.Sprintf("%d", port),
		namespace:  testNamespace,
		uri_tester: NewTester(kubeClient, testNamespace, t),
		kubeClient: kubeClient,
		t:          t,
	}
	return j
}

func (j *JenkinsRef) RawURL(uri string) (string, error) {
	podName := "curl-raw-pod"
	fullURL := fmt.Sprintf("http://%s:%v/%s", j.host, j.port, uri)
	cmd := fmt.Sprintf("TOKEN=`cat /var/run/secrets/kubernetes.io/serviceaccount/token` && curl -X GET -H \"Authorization: Bearer $TOKEN\" %s", fullURL)
	return podName, j.uri_tester.CreateExecPod(podName, cmd)
}

func (j *JenkinsRef) DelRawPod() error {
	return j.kubeClient.CoreV1().Pods(j.namespace).Delete(context.Background(), "curl-raw-pod", metav1.DeleteOptions{})
}

func (j *JenkinsRef) RawPost(uri string) (string, error) {
	podName := "curl-raw-post-pod"
	fullURL := fmt.Sprintf("http://%s:%v/%s", j.host, j.port, uri)
	cmd := fmt.Sprintf("TOKEN=`cat /var/run/secrets/kubernetes.io/serviceaccount/token` && curl -X POST -H \"Authorization: Bearer $TOKEN\" -H \"Content-Type: ''\" -d '' %s", fullURL)
	return podName, j.uri_tester.CreateExecPod(podName, cmd)
}

func (j *JenkinsRef) DelRawPostPod() error {
	return j.kubeClient.CoreV1().Pods(j.namespace).Delete(context.Background(), "curl-raw-post-pod", metav1.DeleteOptions{})
}

func (j *JenkinsRef) JobLogs(namespace, bcName string) (string, error) {
	podName := "curl-job-pod"
	fullURL := fmt.Sprintf("http://%s:%v/job/%s/job/%s-%s/lastBuild/consoleText", j.host, j.port, namespace, namespace, bcName)
	j.t.Logf("getting latest job log for %s/%s: full URL %s\n", namespace, bcName, fullURL)
	cmd := fmt.Sprintf("TOKEN=`cat /var/run/secrets/kubernetes.io/serviceaccount/token` && curl -X GET -H \"Authorization: Bearer $TOKEN\" %s", fullURL)
	return podName, j.uri_tester.CreateExecPod(podName, cmd)
}

func (j *JenkinsRef) DelJobPod() error {
	return j.kubeClient.CoreV1().Pods(j.namespace).Delete(context.Background(), "curl-job-pod", metav1.DeleteOptions{})
}

func (j *JenkinsRef) PastJobLogs(namespace, bcName string, num int) (string, error) {
	podName := "curl-past-job-pod"
	fullURL := fmt.Sprintf("http://%s:%v/job/%s/job/%s-%s/%d/consoleText", j.host, j.port, namespace, namespace, bcName, num)
	j.t.Logf("getting job num %d job log for %s/%s: full URL %s\n", num, namespace, bcName, fullURL)
	cmd := fmt.Sprintf("TOKEN=`cat /var/run/secrets/kubernetes.io/serviceaccount/token` && curl -X GET -H \"Authorization: Bearer $TOKEN\" %s", fullURL)
	return podName, j.uri_tester.CreateExecPod(podName, cmd)
}

func (j *JenkinsRef) DelPastJobLogs() error {
	return j.kubeClient.CoreV1().Pods(j.namespace).Delete(context.Background(), "curl-past-job-pod", metav1.DeleteOptions{})
}

func (j *JenkinsRef) Credential(namespace, secretName string) (string, error) {
	podName := "curl-cred-pod"
	fullURL := fmt.Sprintf("http://%s:%v/credentials/store/system/domain/_/credential/%s-%s/", j.host, j.port, namespace, secretName)
	j.t.Logf("full URL %s\n", fullURL)
	cmd := fmt.Sprintf("TOKEN=`cat /var/run/secrets/kubernetes.io/serviceaccount/token` && curl -X GET -H \"Authorization: Bearer $TOKEN\" %s", fullURL)
	return podName, j.uri_tester.CreateExecPod(podName, cmd)
}

func (j *JenkinsRef) DelCredentialPod() error {
	return j.kubeClient.CoreV1().Pods(j.namespace).Delete(context.Background(), "curl-cred-pod", metav1.DeleteOptions{})
}
