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
		t:          t,
	}
	return j
}

func (j *JenkinsRef) JobLogs(namespace, bcName string) {
	fullURL := fmt.Sprintf("http://%s:%v/job/%s/job/%s-%s/lastBuild/consoleText", j.host, j.port, namespace, namespace, bcName)
	j.t.Logf("full URL %s\n", fullURL)
	cmd := fmt.Sprintf("TOKEN=`cat /var/run/secrets/kubernetes.io/serviceaccount/token` && curl -X GET -H \"Authorization: Bearer $TOKEN\" %s", fullURL)
	j.uri_tester.CreateExecPod("curl-job-pod", cmd)
}
