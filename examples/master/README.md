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

The last step is to instantiate the `sample-app` template. The [sample
app](sample-app) is a simple Ruby application
that runs Sinatra and has one unit test defined to exercise the CI flow.

You have to instantiate the template in both `ci` and `stage` projects.

## Workflow

You can [watch the youtube](https://www.youtube.com/watch?v=HsdmSaz1zhs)
video that shows the full workflow. What happens in the video is:

1. When the `sample-app-test` job is started it fetches the [sample-app](sample-app) sources,
   installs all required rubygems using bundler and then executes the sample unit tests.
   In the job definition, we restricted this job to run only on slaves that have
   the *ruby-20-centos7* label. This will match the Kubernetes Pod Template you see
   in the Kubernetes plugin configuration. Once this job is started and queued,
   the plugin connects to OpenShift and starts the slave Pod using the converted
   S2I image. The job then runs entirely on that slave.
   When this job finishes, the Pod is automatically destroyed by the Kubernetes
   plugin.

2. If the unit tests passed, the `sample-app-build` is started automatically via
   the Jenkins [promoted builds](https://wiki.jenkins-ci.org/display/JENKINS/Promoted+Builds+Plugin)
   plugin. This job will leverage the OpenShift Jenkins plugin and start
   build of the Docker image which will contain 'sample-app'.

3. Once the new Docker image for the `sample-app` is built, the
   `sample-app-stage` project will automatically deploy it into `stage` project
   and notify the QA team about availability for testing.

3. If the `sample-app` image passes the `stage` testing, you have to **manually
   promote** the `sample-app-build` build to be deployed to OpenShift. Since
   re-deploying the application replaces the existing running application, human
   intervention is needed to confirm this step.

4. Once the build is promoted, the `sample-app-deploy` job is started. This job
   will scale down the existing application deployment and redeploy it using the
   new Docker image.
