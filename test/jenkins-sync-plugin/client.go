package e2e

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"testing"

	appset "github.com/openshift/client-go/apps/clientset/versioned"
	buildset "github.com/openshift/client-go/build/clientset/versioned"
	imageset "github.com/openshift/client-go/image/clientset/versioned"
	projectset "github.com/openshift/client-go/project/clientset/versioned"
	routeset "github.com/openshift/client-go/route/clientset/versioned"
	templateset "github.com/openshift/client-go/template/clientset/versioned"

	kubeset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	kubeConfig     *rest.Config
	kubeClient     *kubeset.Clientset
	buildClient    *buildset.Clientset
	appClient      *appset.Clientset
	projectClient  *projectset.Clientset
	templateClient *templateset.Clientset
	imageClient    *imageset.Clientset
	routeClient    *routeset.Clientset
)

func getConfig() (*rest.Config, error) {
	// If an env variable is specified with the config locaiton, use that
	if len(os.Getenv("KUBECONFIG")) > 0 {
		return clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	}
	// If no explicit location, try the in-cluster config
	if c, err := rest.InClusterConfig(); err == nil {
		return c, nil
	}
	// If no in-cluster config, try the default location in the user's home directory
	if usr, err := user.Current(); err == nil {
		if c, err := clientcmd.BuildConfigFromFlags(
			"", filepath.Join(usr.HomeDir, ".kube", "config")); err == nil {
			return c, nil
		}
	}

	return nil, fmt.Errorf("could not locate a kubeconfig")
}

func setupClients(t *testing.T) {
	var err error
	if kubeConfig == nil {
		kubeConfig, err = getConfig()
		if err != nil {
			t.Fatalf("%#v", err)
		}
	}

	if kubeClient == nil {
		kubeClient, err = kubeset.NewForConfig(kubeConfig)
		if err != nil {
			t.Fatalf("%#v", err)
		}
	}

	if buildClient == nil {
		buildClient, err = buildset.NewForConfig(kubeConfig)
		if err != nil {
			t.Fatalf("%#v", err)
		}
	}

	if projectClient == nil {
		projectClient, err = projectset.NewForConfig(kubeConfig)
		if err != nil {
			t.Fatalf("%#v", err)
		}
	}

	if templateClient == nil {
		templateClient, err = templateset.NewForConfig(kubeConfig)
		if err != nil {
			t.Fatalf("%#v", err)
		}
	}

	if imageClient == nil {
		imageClient, err = imageset.NewForConfig(kubeConfig)
		if err != nil {
			t.Fatalf("%#v", err)
		}
	}

	if routeClient == nil {
		routeClient, err = routeset.NewForConfig(kubeConfig)
		if err != nil {
			t.Fatalf("%#v", err)
		}
	}

	if appClient == nil {
		appClient, err = appset.NewForConfig(kubeConfig)
		if err != nil {
			t.Fatalf("%#v", err)
		}
	}

}
