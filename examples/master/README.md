# Jenkins Master example

This directory contains an example of a Jenkins setup that is configured to
demonstrate the CI/pipeline workflow for the sample-app application using the
Jenkins Master/Slave setup and automation done on OpenShift v3.

<p align="center">
<img width="420" src="https://raw.githubusercontent.com/mfojtik/jenkins-ci/master/jenkins-flow.png"/>
</p>

## Installation

To start, you have to manually enter following commands in OpenShift:


```console
# Create project and allow Jenkins to talk to OpenShift API
$ oc new-project ci
$ oc policy add-role-to-user edit system:serviceaccount:ci:default

# Create the 'staging' project where we deploy the sample-app for testing
$ oc new-project stage
$ oc policy add-role-to-user edit system:serviceaccount:stage:default
$ oc policy add-role-to-user edit system:serviceaccount:ci:default

# Now create the templates
$ oc create -n ci -f https://raw.githubusercontent.com/openshift/jenkins/master/examples/master/jenkins-with-k8s-plugin.json
$ oc create -n ci -f https://raw.githubusercontent.com/openshift/jenkins/master/examples/slave/s2i-slave-template.json
```

## Instantiating templates from OpenShift web console

Navigate to the OpenShift UI and choose the `ci` project we created in the previous
step. Now click *Add to Project* button and then click *Show All
Template*. You should see *jenkins-master* template and *s2i-to-jenkins-slave*.

Note: To continue, you have to first create your Jenkins Slave image. Please see
the [instructions](../slave/README.md) here.

### Jenkins Master

This template defines a
[BuildConfig](https://docs.openshift.org/latest/dev_guide/builds.html#defining-a-buildconfig)
that use the [Docker
Strategy](https://docs.openshift.org/latest/dev_guide/builds.html#docker-strategy-options)
to rebuild the official [OpenShift Jenkins Image](https://github.com/openshift/jenkins).
This template also defines a [Deployment Configuration](https://docs.openshift.org/latest/dev_guide/deployments.html#creating-a-deployment-configuration) that will start just one instance
of the Jenkins server.

When you choose the `jenkins-master` template, you have to specify these parameters:

* **JENKINS_SERVICE_NAME** - The name of the Jenkins service (default: *jenkins*)
* **JENKINS_IMAGE** - The name of the original Jenkins image to use
* **JENKINS_PASSWORD** - The Jenkins 'admin' user password

Once you instantiate this template, you should see a new service *jenkins* in
the overview page and a route *https://jenkins-ci.router.default.svc.cluster.local/*.

Optionally, you can configure your own repository with Jenkins plugins,
configuration, etc by changing these parameters:

* **JENKINS_S2I_REPO_URL** - The Git repository with your Jenkins configuration
* **JENKINS_S2I_REPO_CONTEXTDIR** - The directory inside the repository
* **JENKINS_S2I_REPO_REF** - Specify branch or Git ref

You have to wait until OpenShift rebuilds the original image to include all
plugins and configuration needed.

### Sample Application

The last step is to instantiate the `sample-app` templates.  Look for the `sample-app-masterslave-ci.json` and
`sample-app-masterslave-stage.json` files under the [jenkins example folder in OpenShift Origin](https://github.com/openshift/origin/tree/master/examples/jenkins).
These two templates revolve around a [simple Ruby application that runs Sinatra](https://github.com/mfojtik/sample-app) and has one unit test defined to exercise the CI flow.

Instantiate each template in their corresponding projects (`sample-app-masterslave-ci.json` in `ci` and `sample-app-masterslave-stage.json` in `stage`).

The `stage` project spins up deployments of the application for general testing when they are successfully built and unit tested
in the `ci` project.  Deployments in the `ci` project are not spun up without human intervention, presumably following successful testing in
the `stage` project.

See the following workflow for precise details in how these projects are leveraged.

## Workflow

You can [watch the youtube](https://www.youtube.com/watch?v=HsdmSaz1zhs)
video that shows the full workflow. The video includes one addition which we could not
preconfigure in this sample (valid email accounts you have).

But that detail aside, The breakdown of this sample workflow that is illustrated in the the video is:

1. When the `sample-app-test` job is started it fetches the [sample-app](https://github.com/mfojtik/sample-app) sources,
   installs all required rubygems using bundler and then executes the sample unit tests.
   In the job definition, we restricted this job to run only on slaves that have
   the *ruby-20-centos7* label. This will match the Kubernetes Pod Template you see
   in the Kubernetes plugin configuration. Once this job is started and queued,
   the plugin connects to OpenShift and starts the slave Pod using the converted
   S2I image. The job then runs entirely on that slave.
   When this job finishes, the Pod is automatically destroyed by the Kubernetes
   plugin.

2. If the unit tests in `sample-app-test` passed and the job completes successfully, the `sample-app-build` job is started automatically via
   the Jenkins [promoted builds](https://wiki.jenkins-ci.org/display/JENKINS/Promoted+Builds+Plugin)
   plugin. This job will leverage the [OpenShift Jenkins Pipeline plugin](https://github.com/openshift/jenkins-plugin) and start an OpenShift
   build that creates a Docker image that contains 'sample-app'.  The build results will be validated, as well as whether a deployment
   and replication controller based on the resulting image are available (with a replica count of 0).

3. Once the new Docker image for the `sample-app` is built, the
   `sample-app-stage` job will validate that the new Docker image is successfully deployed with a running replica into the `stage` project.  The
   deployment validation is started automatically because the `sample-app-stage` job continually monitors the ImageStream where the image resides
   in the `ci` project and will detect any change in the ImageStream produced by the `sample-app-build` job.

   A sample polling interval has been provided with the sample, but alter the interval to best suit your needs.  Also, no email notification is
   preconfigured (a running SMTP server and a valid email are needed for the post build action, and by extendsion the entire job, to succeed), but you can set up SMTP in your Jenkins environment and
   add a Jenkins email post build action to notify the QA team about your new image's availability for testing.
   
   Also note, with the `sample-app-stage` job's use of the OpenShift Pipeline Plugin's Jenkins SCM extension point (where it polls OpenShift ImageStream's instead
   of your classic source control management systems), the Jenkins SCM extension point infrastructure will start polling the `ci` project as soon as the Jenkins master comes up.
   Additionally, Jenkins' SCM extension point  infrastructure will trigger the job at least once because there are no existing job results, without even contacting the OpenShift Pipeline Plugin's
   SCM extension point to see if a job invocation is needed.  So you'll see
   an initial run of `sample-app-stage` attempted and failed because you have not run `sample-app-test` yet, which would result in a `sample-app-build` run
   that produces the deployment the `sample-app-stage` job is trying to validate.

4. If the `sample-app` image passes whatever `stage` testing you deem appropriate, this sample workflow here introduces the human element of signing off
   on the new Docker image.  You perform this sign off when you **manually promote** the `sample-app-build` build that served as input into the `stage` testing.
   This provides to go ahead for the image to be spun up in the `ci` project and replace any existing, earlier versions of `sample-app` application.

5. Once the build is promoted, the `sample-app-deploy` job is started. This job
   will scale up the new application deployment (remember, the deployment in the `ci` project triggered by the new build is configured to 0 replicas to allow the testing
   to occur first), and then verify the new deployment is available with a running replica.
