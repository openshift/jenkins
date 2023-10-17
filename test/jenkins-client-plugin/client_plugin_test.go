package e2e

import (
	"context"
	"fmt"
	"io"
	"testing"

	buildv1 "github.com/openshift/api/build/v1"
	projectv1 "github.com/openshift/api/project/v1"
	templatev1 "github.com/openshift/api/template/v1"
	buildset "github.com/openshift/client-go/build/clientset/versioned"
	projectset "github.com/openshift/client-go/project/clientset/versioned"
	templateset "github.com/openshift/client-go/template/clientset/versioned"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	utilrand "k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/apimachinery/pkg/watch"
	kubeset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	kubeConfig     *rest.Config
	kubeClient     *kubeset.Clientset
	buildClient    *buildset.Clientset
	projectClient  *projectset.Clientset
	templateClient *templateset.Clientset
)

const (
	testNamespace                           = "jenkins-client-plugin-test-namespace-"
	pipeline_verify_normal_headless_service = `
          try {
              timeout(time: 20, unit: 'MINUTES') {
                  // Select the default cluster
                  openshift.withCluster() {
                      // Select the default project
                      openshift.withProject() {
						// launch redis template
						openshift.newApp("redis-ephemeral", "--name", "redis", "-p", "MEMORY_LIMIT=128Mi")

						// make sure dc rolled out
						def dcs = openshift.selector("dc", "redis")
						dcs.related('pods').untilEach(1) {
							if (it.object().status.phase != 'Pending') {
								return true;
							}
							return false;
						}

                        // Verify Normal Services
                        def connectedNormalService = openshift.verifyService('redis')
                        // Verify Headless Services with Selectors
                        def connectedHeadlessService = openshift.verifyService('redis-headless')
                      }
                  }
              }
          } catch (err) {
             echo "in catch block"
             echo "Caught: ${err}"
             currentBuild.result = 'FAILURE'
             throw err
          }
`
	pipeline_parallel_with_lock = `
  pipeline {
    agent none
    stages {
        stage('test init') {
    	    steps {
    	        script {
	                openshift.setLockName('openshift-dls-test')
    	        }
    	    }
        }
        stage('Run tests') {
            parallel {
                stage('Test run 1') {
                    steps {
                        script {
                            actualTest()
                        }
                    }
                }

                stage('Test run 2') {
                    steps {
                        script {
                            actualTest()
                        }
                    }
                }
            }
        }
    }
  }
  void actualTest() {
      try {
          timeout(time: 20, unit: 'MINUTES') {
              // Select the default cluster
              openshift.withCluster() {
                  // Select the default project
                  openshift.withProject() {
                      echo "Using project ${openshift.project()} in cluster with url ${openshift.cluster()}"
				      sleep 10
                  }
              }
          }
      } catch (err) {
          echo "in catch block"
          echo "Caught: ${err}"
          currentBuild.result = 'FAILURE'
          throw err
      }

  }
`
	pipeline_run_cmd = `
	try {
        timeout(time: 20, unit: 'MINUTES') {
			openshift.withCluster() {
				openshift.withProject() {
					// exercise oc run path, including verification of proper handling of groovy cps
					// var binding (converting List to array)
					// using the quay origin-jenkins:4.3+ as it deploys without error on 4.x
					def runargs1 = []
					runargs1 << "jenkins-second-deployment"
					runargs1 << "--image=quay.io/openshift/origin-jenkins:4.3"
					runargs1 << "--dry-run"
					runargs1 << "-o yaml"
					def r1 = openshift.run(runargs1)
					echo "run 1 status: ${r1.status}"
					echo "run 1 err: ${r1.err}"
					echo "run 1 actions size: ${r1.actions.size()}"
					echo "run 1 stdout: ${r1.getOut()}"
					echo "run 1 stderr: ${r1.getErr()}"
					if (r1.status != 0) {
						error("r1 error")
					}
					if (r1.getOut().length() == 0) {
						error("r1 expected output")
					}
					if (!r1.getOut().contains("kind: Pod")) {
						error("r1 no pod yaml")
					}

					// FYI - pipeline cps groovy compile does not allow String[] runargs2 =  {"jenkins-second-deployment", "--image=docker.io/openshift/jenkins-2-centos7:latest", "--dry-run"}
					String[] runargs2 = new String[4]
					runargs2[0] = "jenkins-second-deployment"
					runargs2[1] = "--image=quay.io/openshift/origin-jenkins:4.3"
					runargs2[2] = "--dry-run"
					runargs2[3] = "-o yaml"
					def r2 = openshift.run(runargs2)
					echo "run 2 status: ${r2.status}"
					echo "run 2 err: ${r2.err}"
					echo "run 2 actions size: ${r2.actions.size()}"
					echo "run 2 stdout: ${r2.getOut()}"
					echo "run 2 stderr: ${r2.getErr()}"
					if (r2.status != 0) {
						error("r2 error")
					}
					if (r2.getOut().length() == 0) {
						error("r2 expected output")
					}
					if (!r2.getOut().contains("kind: Pod")) {
						error("r2 no pod yaml")
					}
				}
			}
		}
	} catch (err) {
        echo "in catch block"
        echo "Caught: ${err}"
        currentBuild.result = 'FAILURE'
        throw err
    }
`
	pipeline_selector_patch = `
    try {
        timeout(time: 20, unit: 'MINUTES') {
            // Select the default cluster
            openshift.withCluster() {
                // Select the default project
                openshift.withProject() {

                    // Output the url of the currently selected cluster
                    echo "Using project ${openshift.project()} in cluster with url ${openshift.cluster()}"

                    def currentProject = openshift.project()
                    def templateSelector = openshift.selector("template", "postgresql-ephemeral")
                    def exist = templateSelector.exists()
                    if (!exist) {
                       openshift.create('https://raw.githubusercontent.com/openshift/cluster-samples-operator/release-4.11/assets/operator/ocp-x86_64/postgresql/templates/postgresql-ephemeral.json')
                    } else {
                       openshift.selector( 'svc', [ app:'postgresql-ephemeral' ] ).delete()
                       openshift.selector( 'dc', [ app:'postgresql-ephemeral' ] ).delete()
                    }
                    openshift.newApp("--template=${currentProject}/postgresql-ephemeral")
                    openshift.patch("dc/postgresql", '\'{"spec":{"strategy":{"type":"Recreate"}}}\'')


                }
            }
        }
    } catch (err) {
        echo "in catch block"
        echo "Caught: ${err}"
        currentBuild.result = 'FAILURE'
        throw err
    }
`
	pipeline_unquoted_param_spaces = `
    try {
        timeout(time: 20, unit: 'MINUTES') {
            // Select the default cluster
            openshift.withCluster() {
                // Select the default project
                openshift.withProject() {

                    // Output the url of the currently selected cluster
                    echo "Using project ${openshift.project()} in cluster with url ${openshift.cluster()}"

                    // verify we can handle unquoted param values with spaces
                    def templateSelector = openshift.selector( "template", "mariadb-ephemeral")
                    def templateExists = templateSelector.exists()
                    def template
                    if (!templateExists) {
                        template = openshift.create('https://raw.githubusercontent.com/openshift/cluster-samples-operator/release-4.11/assets/operator/ocp-x86_64/mariadb/templates/mariadb-ephemeral.json').object()
                    } else {
                        template = templateSelector.object()
                    }
                    def muser = "All Users"
                    openshift.process( template, '-p', "MYSQL_USER=${muser}")
                    def exist2 = openshift.selector("template", "grape-spring-boot").exists()
                    if (!exist2) {
                        openshift.create("https://raw.githubusercontent.com/openshift/jenkins-client-plugin/master/examples/issue184-template.yml")
                    }
                    def exist3 = openshift.selector("template", "postgresql-ephemeral").exists()
                    if (!exist3) {
                       openshift.create('https://raw.githubusercontent.com/openshift/cluster-samples-operator/release-4.11/assets/operator/ocp-x86_64/postgresql/templates/postgresql-ephemeral.json')
                    }
                    openshift.process("postgresql-ephemeral", "-p=MEMORY_LIMIT=120 -p=NAMESPACE=80 -p=DATABASE_SERVICE_NAME=\"-Xmx768m -Dmy.sys.param=aete\" -p=POSTGRESQL_USER=verify -p=POSTGRESQL_PASSWORD=aete -p=POSTGRESQL_DATABASE=400 -p=POSTGRESQL_VERSION=grape-regtest-tools-aete")
                    openshift.process("grape-spring-boot", "-p=LIVENESS_INITIAL_DELAY_SECONDS=120 -p=READYNESS_INITIAL_DELAY_SECONDS=80 -p=JVMARGS=\"-Xmx768m -Dmy.sys.param=aete\"-p=APPNAME=verify -p=DEPLOYMENTTAG=aete -p=ROLLING_TIMEOUT_SECONDS=400 -p=NAMESPACE=grape-regtest-tools-aete")
                    openshift.process("grape-spring-boot", "-p LIVENESS_INITIAL_DELAY_SECONDS=120 -p READYNESS_INITIAL_DELAY_SECONDS=80 -p JVMARGS=\"-Xmx768m -Dmy.sys.param=aete\"-p APPNAME=verify -p DEPLOYMENTTAG=aete -p ROLLING_TIMEOUT_SECONDS=400 -p NAMESPACE=grape-regtest-tools-aete")
                    openshift.process("grape-spring-boot", "-p=LIVENESS_INITIAL_DELAY_SECONDS=120", "-p=READYNESS_INITIAL_DELAY_SECONDS=80", "-p=JVMARGS=\"-Xmx768m -Dmy.sys.param=aete\"", "-p=APPNAME=verify", "-p=DEPLOYMENTTAG=aete", "-p=ROLLING_TIMEOUT_SECONDS=400", "-p=NAMESPACE=grape-regtest-tools-aete")


                }
            }
        }
    } catch (err) {
        echo "in catch block"
        echo "Caught: ${err}"
        currentBuild.result = 'FAILURE'
        throw err
    }
`
	pipeline = `
    /**
     * This script does nothing in particular,
     * but is meant to show actual usage of most of the API.
     */

    try {
        timeout(time: 20, unit: 'MINUTES') {
            // Select the default cluster
            openshift.withCluster() {
                // Select the default project
                openshift.withProject() {

                    // Output the url of the currently selected cluster
                    echo "Using project ${openshift.project()} in cluster with url ${openshift.cluster()}"

                    // Test selector.annotate
                    def railsTemplate = openshift.create("https://raw.githubusercontent.com/openshift/rails-ex/master/openshift/templates/rails-postgresql.json")
                    railsTemplate.annotate([key1:"value1", key2:"value2"])
                    railsTemplate.delete()

                    def saSelector1 = openshift.selector( "serviceaccount" )
                    saSelector1.describe()

                    def templateSelector = openshift.selector( "template", "mariadb-ephemeral")

                    def templateExists = templateSelector.exists()

                    def templateGeneratedSelector = openshift.selector(["dc/mariadb", "service/mariadb", "secret/mariadb"])

                    def objectsGeneratedFromTemplate = templateGeneratedSelector.exists()

                    // create single object in array
                    def bc = [[
                        "kind":"BuildConfig",
                        "apiVersion":"build.openshift.io/v1",
                        "metadata":[
                            "name":"test",
                            "labels":[
                                "name":"test"
                            ]
                        ],
                        "spec":[
                            "triggers":[],
                            "source":[
                                "type":"Binary"
                            ],
                            "strategy":[
                                "type":"Source",
                                "sourceStrategy":[
                                    "from":[
                                        "kind":"DockerImage",
                                        "name":"centos/ruby-25-centos7"
                                    ]
                                ]
                            ],
                            "output":[
                                "to":[
                                    "kind":"ImageStreamTag",
                                    "name":"test:latest"
                                ]
                            ]
                        ]
                      ]
                    ]
                    def objs = openshift.create( bc )
                    objs.describe()
                    openshift.delete("bc", "test")
                    // switch to delete below when v1.0.10 is available in the image
                    //openshift.delete(bc)
                    openshift.create("configmap", "foo")
                    openshift.create("configmap", "bar")
                    openshift.delete("configmap/foo", "configmap/bar")
                    openshift.create("configmap", "foo")
                    openshift.delete("configmap/foo")

                    def template
                    if (!templateExists) {
                        template = openshift.create('https://raw.githubusercontent.com/openshift/cluster-samples-operator/release-4.11/assets/operator/ocp-x86_64/mariadb/templates/mariadb-ephemeral.json').object()
                    } else {
                        template = templateSelector.object()
                    }

                    // Explore the Groovy object which models the OpenShift template as a Map
                    echo "Template contains ${template.parameters.size()} parameters"

                    // For fun, modify the template easily while modeled in Groovy
                    template.labels["mylabel"] = "myvalue"

                    // Process the modeled template. We could also pass JSON/YAML, a template name, or a url instead.
                    // note: -p option for oc process not in the oc version that we currently ship with openshift jenkins images
                    def objectModels = openshift.process( template )//, "-p", "MEMORY_LIMIT=600Mi")

                    // objectModels is a list of objects the template defined, modeled as Groovy objects
                    echo "The template references ${objectModels.size()} objects"

                    // For fun, modify the objects that have been defined by processing the template
                    for ( o in objectModels ) {
                        o.metadata.labels[ "anotherlabel" ] = "anothervalue"
                    }

                    def objects
                    def verb
                    if (!objectsGeneratedFromTemplate) {
                        verb = "Created"
                        // Serialize the objects and pass them to the create API.
                        // We could also pass JSON/YAML directly; openshift.create(readFile('some.json'))
                        objects = openshift.create(objectModels)
                    } else {
                        verb = "Found"
                        objects = templateGeneratedSelector
                    }

                    // Create returns a selector which will always select the objects created
                    objects.withEach {
                        // Each loop binds the variable 'it' to a selector which selects a single object
                        echo "${verb} ${it.name()} from template with labels ${it.object().metadata.labels}"
                    }

                    // Filter created objects and create a selector which selects only the new DeploymentConfigs
                    def dcs = objects.narrow("dc")
                    echo "Database will run in deployment config: ${dcs.name()}"
                    // Find a least one pod related to the DeploymentConfig and wait it satisfies a condition
                    dcs.related('pods').untilEach(1) {
                        // untilEach only terminates when each selected item causes the body to return true
                        if (it.object().status.phase != 'Pending') {
                        // some example debug of the pod in question
                            shortname = it.object().metadata.name
                            echo openshift.rsh("${shortname}", "ps ax").out
                            return true;
                        }
                        return false;
                    }

                    // Print out all pods created by the DC
                    echo "Template created pods: ${dcs.related('pods').names()}"

                    // Show how we can use labels to select as well
                    echo "Finding dc using labels instead: ${openshift.selector('dc',[mylabel:'myvalue']).names()}"

                    echo "DeploymentConfig description"
                    dcs.describe()
                    echo "DeploymentConfig history"
                    dcs.rollout().history()
					dcs.rollout().status("-w")

                    openshift.verifyService('mariadb')

                    def rubySelector = openshift.selector("bc", "ruby")
                    def builds
                    try {
                        rubySelector.object()
                        builds = rubySelector.related( "builds" )
                    } catch (Throwable t) {
                        // The selector returned from newBuild will select all objects created by the operation
                        nb = openshift.newBuild( "https://github.com/openshift/ruby-hello-world", "--name=ruby" )

                        // Print out information about the objects created by newBuild
                        echo "newBuild created: ${nb.count()} objects : ${nb.names()}"

                        // Filter non-BuildConfig objects and create selector which will find builds related to the BuildConfig
                        builds = nb.narrow("bc").related( "builds" )

                    }

                    //make sure we handle empty selectors correctly
                    def nopods = openshift.selector("pod", [ app: "asdf" ])
                    nopods.withEach {
                      echo "should not see this echo"
                    }

                    // Raw watch which only terminates when the closure body returns true
                    builds.watch {
                        // 'it' is bound to the builds selector.
                        // Continue to watch until at least one build is detected
                        if ( it.count() == 0 ) {
                            return false
                        }
                        // Print out the build's name and terminate the watch
                        echo "Detected new builds created by buildconfig: ${it.names()}"
                        return true
                    }

                    echo "Waiting for builds to complete..."

                    // Like a watch, but only terminate when at least one selected object meets condition
                    builds.untilEach {
                        return it.object().status.phase == "Complete"
                    }

                    // Print a list of the builds which have been created
                    echo "Build logs for ${builds.names()}:"

                    // Find the bc again, and ask for its logs
                    def result = rubySelector.logs()

                    // Each high-level operation exposes stout/stderr/status of oc actions that composed
                    echo "Result of logs operation:"
                    echo "  status: ${result.status}"
                    echo "  stderr: ${result.err}"
                    echo "  number of actions to fulfill: ${result.actions.size()}"
                    echo "  first action executed: ${result.actions[0].cmd}"

                    // The following steps below are geared toward testing of bugs or features that have been introduced
                    // into the openshift client plugin since its initial release


                    // Empty static / selectors are powerful tools to check the state of the system.
                    // Intentionally create one using a narrow and exercise it.
                    emptySelector = openshift.selector("pods").narrow("bc")
                    openshift.failUnless(!emptySelector.exists()) // Empty selections never exist
                    openshift.failUnless(emptySelector.count() == 0)
                    openshift.failUnless(emptySelector.names().size() == 0)
                    emptySelector.delete() // Should have no impact
                    emptySelector.label(["x":"y"]) // Should have no impact

                    // sanity check for latest and cancel
                    def dc3Selector = openshift.selector("dc", "mariadb")
                    dc3Selector.rollout().latest()
                    sleep 3
                    dc3Selector.rollout().cancel()

                    // validate some watch/selector error handling
                    try {
                        timeout(time: 10, unit: 'SECONDS') {
                            builds.untilEach {
                                  sleep(20)
                            }
                        }
                        error( "exception did not escape the watch as expected" )
                    } catch ( e ) {
                        // test successful
                    }
                    try {
                        builds.watch {
                            error( "this should be thrown" )
                        }
                        error( "exception did not escape the watch as expected" )
                    } catch ( e ) {
                        // test successful
                    }

                }
            }
        }
    } catch (err) {
        echo "in catch block"
        echo "Caught: ${err}"
        currentBuild.result = 'FAILURE'
        throw err
    }

//}

`
)

func setupClients(t *testing.T) {
	var err error
	if kubeConfig == nil {
		kubeConfig, err = GetConfig()
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

}

func instantiateTemplate(ta *testArgs) {
	template, err := templateClient.TemplateV1().Templates(ta.templateNs).Get(context.Background(),
		ta.template, metav1.GetOptions{})
	if err != nil {
		ta.t.Fatalf("%#v", err)
	}

	// INSTANTIATE THE TEMPLATE.

	// To set Template parameters, create a Secret holding overridden parameters
	// and their values.
	secret, err := kubeClient.CoreV1().Secrets(ta.ns).Create(context.Background(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: ta.template,
		},
		StringData: ta.templateParams,
	}, metav1.CreateOptions{})
	if err != nil {
		ta.t.Fatalf("%#v", err)
	}

	// Create a TemplateInstance object, linking the Template and a reference to
	// the Secret object created above.
	ti, err := templateClient.TemplateV1().TemplateInstances(ta.ns).Create(context.Background(),
		&templatev1.TemplateInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name: ta.template,
			},
			Spec: templatev1.TemplateInstanceSpec{
				Template: *template,
				Secret: &corev1.LocalObjectReference{
					Name: secret.Name,
				},
			},
		}, metav1.CreateOptions{})
	if err != nil {
		ta.t.Fatalf("%#v", err)
	}

	// Watch the TemplateInstance object until it indicates the Ready or
	// InstantiateFailure status condition.
	watcher, err := templateClient.TemplateV1().TemplateInstances(ta.ns).Watch(context.Background(),
		metav1.SingleObject(ti.ObjectMeta),
	)
	if err != nil {
		ta.t.Fatalf("%#v", err)
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
					dumpPods(ta)
					ta.t.Fatalf("templateinstance instantiation failed reason %s message %s", cond.Reason, cond.Message)
				}
			}

		default:
			ta.t.Logf("unexpected event type %s: %#v", string(event.Type), event.Object)
		}
	}

}

func instantiateBuild(ta *testArgs) {
	if !ta.skipBCCreate {
		_, err := buildClient.BuildV1().BuildConfigs(ta.ns).Create(context.Background(), ta.bc, metav1.CreateOptions{})
		if err != nil {
			ta.t.Fatalf("%#v", err)
		}
	}
	build, err := buildClient.BuildV1().BuildConfigs(ta.ns).Instantiate(context.Background(),
		ta.bc.Name,
		&buildv1.BuildRequest{
			ObjectMeta: metav1.ObjectMeta{Name: ta.bc.Name},
		}, metav1.CreateOptions{})
	if err != nil {
		ta.t.Fatalf("%#v", err)
	}
	watcher, err := buildClient.BuildV1().Builds(ta.ns).Watch(context.Background(),
		metav1.SingleObject(build.ObjectMeta))
	if err != nil {
		ta.t.Fatalf("%#v", err)
	}

	for event := range watcher.ResultChan() {
		switch event.Type {
		case watch.Modified:
			build = event.Object.(*buildv1.Build)

			switch build.Status.Phase {
			case buildv1.BuildPhaseComplete:
				watcher.Stop()
				return
			case buildv1.BuildPhaseError:
				ta.t.Logf("build error: %#v", build)
				ta.t.Log("dump job log")
				NewRef(ta.t, kubeClient, ta.ns).JobLogs(ta.ns, ta.bc.Name)
				ta.t.Log("dump namespace pod logs")
				dumpPods(ta)
				watcher.Stop()
				ta.t.Fatal()
			case buildv1.BuildPhaseFailed:
				ta.t.Logf("build failed: %#v", build)
				ta.t.Log("dump job log")
				NewRef(ta.t, kubeClient, ta.ns).JobLogs(ta.ns, ta.bc.Name)
				ta.t.Log("dump namespace pod logs")
				dumpPods(ta)
				watcher.Stop()
				ta.t.Fatal()
			default:
				ta.t.Logf("build phase %s", build.Status.Phase)
			}

		}
	}
}

func dumpPods(ta *testArgs) {
	podClient := kubeClient.CoreV1().Pods(ta.ns)
	podList, err := podClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		ta.t.Fatalf("error list pods %v", err)
	}
	ta.t.Logf("dumpPods have %d items in list", len(podList.Items))
	for _, pod := range podList.Items {
		ta.t.Logf("dumpPods looking at pod %s in phase %s", pod.Name, pod.Status.Phase)

		for _, container := range pod.Spec.Containers {
			req := podClient.GetLogs(pod.Name, &corev1.PodLogOptions{Container: container.Name})
			readCloser, err := req.Stream(context.TODO())
			if err != nil {
				ta.t.Fatalf("error getting pod logs for container %s: %s", container.Name, err.Error())
			}
			b, err := io.ReadAll(readCloser)
			if err != nil {
				ta.t.Fatalf("error reading pod stream %s", err.Error())
			}
			podLog := string(b)
			ta.t.Logf("pod logs for container %s in pod %s:  %s", container.Name, pod.Name, podLog)

		}

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
	t              *testing.T
	ns             string
	template       string
	templateNs     string
	templateParams map[string]string
	bc             *buildv1.BuildConfig
	skipBCCreate   bool
}

func basicPipelineInvocationAndValidation(ta *testArgs) {
	setupClients(ta.t)

	randomTestNamespaceName := generateName(testNamespace)
	ta.ns = randomTestNamespaceName
	defer projectClient.ProjectV1().Projects().Delete(context.Background(), randomTestNamespaceName, metav1.DeleteOptions{})
	_, err := projectClient.ProjectV1().ProjectRequests().Create(context.Background(), &projectv1.ProjectRequest{
		ObjectMeta: metav1.ObjectMeta{Name: randomTestNamespaceName},
	}, metav1.CreateOptions{})

	if err != nil {
		ta.t.Fatalf("%#v", err)
	}

	ta.template = "jenkins-ephemeral"
	ta.templateNs = "openshift"
	ta.templateParams = map[string]string{"MEMORY_LIST": "2048Mi"}
	instantiateTemplate(ta)

	instantiateBuild(ta)

}

func TestPlugin(t *testing.T) {
	bc := buildv1.BuildConfig{}
	bc.Name = "client-plugin-sample"
	bc.Spec = buildv1.BuildConfigSpec{
		CommonSpec: buildv1.CommonSpec{
			Strategy: buildv1.BuildStrategy{
				JenkinsPipelineStrategy: &buildv1.JenkinsPipelineBuildStrategy{
					Jenkinsfile: pipeline,
				},
			},
		},
	}
	ta := &testArgs{
		t:  t,
		bc: &bc,
	}
	basicPipelineInvocationAndValidation(ta)
}

func TestUnquotedParamsWithSpaces(t *testing.T) {
	bc := buildv1.BuildConfig{}
	bc.Name = "client-plugin-unquoted-params-spaces"
	bc.Spec = buildv1.BuildConfigSpec{
		CommonSpec: buildv1.CommonSpec{
			Strategy: buildv1.BuildStrategy{
				JenkinsPipelineStrategy: &buildv1.JenkinsPipelineBuildStrategy{
					Jenkinsfile: pipeline_unquoted_param_spaces,
				},
			},
		},
	}
	ta := &testArgs{
		t:  t,
		bc: &bc,
	}
	basicPipelineInvocationAndValidation(ta)
}

func TestSelectorWithPath(t *testing.T) {
	bc := buildv1.BuildConfig{}
	bc.Name = "client-plugin-selector-with-patch"
	bc.Spec = buildv1.BuildConfigSpec{
		CommonSpec: buildv1.CommonSpec{
			Strategy: buildv1.BuildStrategy{
				JenkinsPipelineStrategy: &buildv1.JenkinsPipelineBuildStrategy{
					Jenkinsfile: pipeline_selector_patch,
				},
			},
		},
	}
	ta := &testArgs{
		t:  t,
		bc: &bc,
	}
	basicPipelineInvocationAndValidation(ta)
}

func TestRun(t *testing.T) {
	bc := buildv1.BuildConfig{}
	bc.Name = "client-plugin-run-cmd"
	bc.Spec = buildv1.BuildConfigSpec{
		CommonSpec: buildv1.CommonSpec{
			Strategy: buildv1.BuildStrategy{
				JenkinsPipelineStrategy: &buildv1.JenkinsPipelineBuildStrategy{
					Jenkinsfile: pipeline_run_cmd,
				},
			},
		},
	}
	ta := &testArgs{
		t:  t,
		bc: &bc,
	}
	basicPipelineInvocationAndValidation(ta)
}

func TestParallelWithLock(t *testing.T) {
	bc := buildv1.BuildConfig{}
	bc.Name = "client-plugin-parallel-with-lock"
	bc.Spec = buildv1.BuildConfigSpec{
		CommonSpec: buildv1.CommonSpec{
			Strategy: buildv1.BuildStrategy{
				JenkinsPipelineStrategy: &buildv1.JenkinsPipelineBuildStrategy{
					Jenkinsfile: pipeline_parallel_with_lock,
				},
			},
		},
	}
	ta := &testArgs{
		t:  t,
		bc: &bc,
	}
	basicPipelineInvocationAndValidation(ta)
}

func TestVerifyHeadlessService(t *testing.T) {
	ta := &testArgs{t: t}
	setupClients(ta.t)

	randomTestNamespaceName := generateName(testNamespace)
	ta.ns = randomTestNamespaceName
	defer projectClient.ProjectV1().Projects().Delete(context.Background(), randomTestNamespaceName, metav1.DeleteOptions{})
	_, err := projectClient.ProjectV1().ProjectRequests().Create(context.Background(), &projectv1.ProjectRequest{
		ObjectMeta: metav1.ObjectMeta{Name: randomTestNamespaceName},
	}, metav1.CreateOptions{})

	if err != nil {
		ta.t.Fatalf("%#v", err)
	}

	ta.template = "jenkins-ephemeral"
	ta.templateNs = "openshift"
	ta.templateParams = map[string]string{"MEMORY_LIST": "2048Mi"}
	instantiateTemplate(ta)

	// create headless servie to go along with normal service defined in template
	// used in pipeline
	headless := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "redis-headless",
			Labels: map[string]string{"app": "redis"},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port: 6379,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 6379,
						StrVal: "",
					},
				},
			},
			Selector:  map[string]string{"name": "redis"},
			ClusterIP: "None",
		},
	}
	_, err = kubeClient.CoreV1().Services(randomTestNamespaceName).Create(context.Background(), headless, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("%s", err.Error())
	}

	bc := &buildv1.BuildConfig{}
	bc.Name = "client-plugin-verify-headless"
	bc.Spec = buildv1.BuildConfigSpec{
		CommonSpec: buildv1.CommonSpec{
			Strategy: buildv1.BuildStrategy{
				JenkinsPipelineStrategy: &buildv1.JenkinsPipelineBuildStrategy{
					Jenkinsfile: pipeline_verify_normal_headless_service,
				},
			},
		},
	}
	ta.bc = bc
	instantiateBuild(ta)
}

func TestMultiNamespaceTemplates(t *testing.T) {
	setupClients(t)

	randomTestNamespaceName1 := generateName(testNamespace)
	randomTestNamespaceName2 := generateName(testNamespace)
	randomTestNamespaceName3 := generateName(testNamespace)
	defer projectClient.ProjectV1().Projects().Delete(context.Background(), randomTestNamespaceName1, metav1.DeleteOptions{})
	defer projectClient.ProjectV1().Projects().Delete(context.Background(), randomTestNamespaceName2, metav1.DeleteOptions{})
	defer projectClient.ProjectV1().Projects().Delete(context.Background(), randomTestNamespaceName3, metav1.DeleteOptions{})

	testNamespaces := []string{randomTestNamespaceName1, randomTestNamespaceName2, randomTestNamespaceName3}

	for _, randomTestNamespaceName := range testNamespaces {
		_, err := projectClient.ProjectV1().ProjectRequests().Create(context.Background(), &projectv1.ProjectRequest{
			ObjectMeta: metav1.ObjectMeta{Name: randomTestNamespaceName},
		}, metav1.CreateOptions{})

		if err != nil {
			t.Fatalf("%#v", err)
		}
	}

	ta1 := &testArgs{
		t:              t,
		ns:             randomTestNamespaceName1,
		template:       "jenkins-ephemeral",
		templateNs:     "openshift",
		templateParams: map[string]string{"MEMORY_LIST": "2048Mi"},
	}
	instantiateTemplate(ta1)

	_, err := kubeClient.RbacV1().RoleBindings(randomTestNamespaceName2).Create(context.Background(), &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "binding2",
		},
		Subjects: []rbacv1.Subject{
			{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "User",
				Name:     "system:serviceaccount:" + randomTestNamespaceName1 + ":jenkins",
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "edit",
		},
	}, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("%#v", err)
	}
	_, err = kubeClient.RbacV1().RoleBindings(randomTestNamespaceName3).Create(context.Background(), &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "binding3",
		},
		Subjects: []rbacv1.Subject{
			{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "User",
				Name:     "system:serviceaccount:" + randomTestNamespaceName1 + ":jenkins",
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "edit",
		},
	}, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("%#v", err)
	}

	testTemplateTemplate := &templatev1.Template{
		ObjectMeta: metav1.ObjectMeta{
			Name: "multi-namespace-template",
		},
		Objects: []runtime.RawExtension{
			{
				Object: &corev1.Secret{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Secret",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "mariadb",
						Namespace: "${NAMESPACE1}",
					},
					StringData: map[string]string{
						"database-name": "foo",
					},
				},
			},
			{
				Object: &corev1.Secret{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Secret",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "mariadb",
						Namespace: "${NAMESPACE2}",
					},
					StringData: map[string]string{
						"database-name": "foo",
					},
				},
			},
			{
				Object: &corev1.Service{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Service",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "mariadb",
						Namespace: "${NAMESPACE3}",
					},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{
							{
								Name: "mariadb",
								Port: 3306,
							},
						},
						Selector: map[string]string{
							"name": "mariadb",
						},
					},
				},
			},
		},
		Parameters: []templatev1.Parameter{
			{
				Name:  "NAMESPACE1",
				Value: randomTestNamespaceName1,
			},
			{
				Name:  "NAMESPACE2",
				Value: randomTestNamespaceName2,
			},
			{
				Name:  "NAMESPACE3",
				Value: randomTestNamespaceName3,
			},
		},
		ObjectLabels: map[string]string{
			"template": "multi-namespace-template",
		},
	}
	_, err = templateClient.TemplateV1().Templates(randomTestNamespaceName1).Create(context.Background(), testTemplateTemplate, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("%s", err.Error())
	}

	testPipelineTemplate := &templatev1.Template{
		ObjectMeta: metav1.ObjectMeta{
			Name: "multi-namespace-pipeline",
		},
		Objects: []runtime.RawExtension{
			{
				Object: &buildv1.BuildConfig{
					TypeMeta: metav1.TypeMeta{
						Kind:       "BuildConfig",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "multi-namespace-pipeline",
					},
					Spec: buildv1.BuildConfigSpec{
						CommonSpec: buildv1.CommonSpec{
							Strategy: buildv1.BuildStrategy{
								Type: buildv1.JenkinsPipelineBuildStrategyType,
								JenkinsPipelineStrategy: &buildv1.JenkinsPipelineBuildStrategy{
									Env: []corev1.EnvVar{
										{
											Name:  "NAMESPACE1",
											Value: "${NAMESPACE1}",
										},
										{
											Name:  "NAMESPACE2",
											Value: "${NAMESPACE2}",
										},
										{
											Name:  "NAMESPACE3",
											Value: "${NAMESPACE3}",
										},
									},
									Jenkinsfile: `
          try {
              timeout(time: 20, unit: 'MINUTES') {
                  // Select the default cluster
                  openshift.withCluster() {
                      // Select the default project
                      openshift.withProject() {

                          // Output the url of the currently selected cluster
                          echo "Using project ${openshift.project()} in cluster with url ${openshift.cluster()}"

                          def templateSelector = openshift.selector( "template", "multi-namespace-template")
                          template = templateSelector.object()

                          // Explore the Groovy object which models the OpenShift template as a Map
                          echo "Template contains ${template.parameters.size()} parameters"

                          // Process the modeled template. We could also pass JSON/YAML, a template name, or a url instead.
                          def objectModels = openshift.process( template, "-p", "NAMESPACE1=${env.NAMESPACE1}", "-p", "NAMESPACE2=${env.NAMESPACE2}", "-p", "NAMESPACE3=${env.NAMESPACE3}" )

                          // objectModels is a list of objects the template defined, modeled as Groovy objects
                          echo "The template references ${objectModels.size()} objects"

                          def objects = openshift.create(objectModels)

                          // Create returns a selector which will always select the objects created
                          objects.withEach {
                              // Each loop binds the variable 'it' to a selector which selects a single object
                              echo "Created ${it.name()} from template with labels ${it.object().metadata.labels}"
                          }


                      }
                  }
              }
          } catch (err) {
             echo "in catch block"
             echo "Caught: ${err}"
             currentBuild.result = 'FAILURE'
             throw err
          }
`,
								},
							},
						},
					},
				},
			},
		},
		Parameters: []templatev1.Parameter{
			{
				Name:  "NAMESPACE1",
				Value: randomTestNamespaceName1,
			},
			{
				Name:  "NAMESPACE2",
				Value: randomTestNamespaceName2,
			},
			{
				Name:  "NAMESPACE3",
				Value: randomTestNamespaceName3,
			},
		},
		ObjectLabels: map[string]string{
			"template": "multi-namespace-pipeline",
		},
	}

	_, err = templateClient.TemplateV1().Templates(randomTestNamespaceName1).Create(context.Background(), testPipelineTemplate, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("%s", err.Error())
	}

	ta2 := &testArgs{
		t:              t,
		ns:             randomTestNamespaceName1,
		template:       "multi-namespace-pipeline",
		templateNs:     randomTestNamespaceName1,
		templateParams: map[string]string{"NAMESPACE": randomTestNamespaceName1, "NAMESPACE2": randomTestNamespaceName2, "NAMESPACE3": randomTestNamespaceName3},
	}
	instantiateTemplate(ta2)

	bc, err := buildClient.BuildV1().BuildConfigs(randomTestNamespaceName1).Get(context.Background(), "multi-namespace-pipeline", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	ta3 := &testArgs{
		t:            t,
		ns:           randomTestNamespaceName1,
		bc:           bc,
		skipBCCreate: true,
	}
	instantiateBuild(ta3)

	_, err = kubeClient.CoreV1().Secrets(randomTestNamespaceName1).Get(context.Background(), "mariadb", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	_, err = kubeClient.CoreV1().Secrets(randomTestNamespaceName2).Get(context.Background(), "mariadb", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	_, err = kubeClient.CoreV1().Services(randomTestNamespaceName3).Get(context.Background(), "mariadb", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("%s", err.Error())
	}

}
