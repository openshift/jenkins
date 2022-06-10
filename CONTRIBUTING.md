# Development and release process for the OpenShift Jenkins plugins

The OpenShift organization currently maintains several plugins related to integration between V3 OpenShift and V2 Jenkins.

|     |      |      |   |
| -------------- | --------------------------   | ------------------------   | -------------- |
| `OpenShift Client Plugin` | [OpenShift Repo](https://github.com/openshift/jenkins-client-plugin) | [Jenkins Repo](https://github.com/jenkinsci/openshift-client-plugin) | [Wiki](https://wiki.jenkins.io/display/JENKINS/OpenShift+Client+Plugin) |
| `OpenShift Sync Plugin` | [OpenShift Repo](https://github.com/openshift/jenkins-sync-plugin) | [Jenkins Repo](https://github.com/jenkinsci/openshift-sync-plugin) | [Wiki](https://wiki.jenkins.io/display/JENKINS/OpenShift+Sync+Plugin) |
| `OpenShift Login Plugin` | [OpenShift Repo](https://github.com/openshift/jenkins-openshift-login-plugin) | [Jenkins Repo](https://github.com/jenkinsci/openshift-login-plugin) | [Wiki](https://wiki.jenkins.io/display/JENKINS/OpenShift+Login+Plugin) |
| `OpenShift Pipeline Plugin` | [OpenShift Repo](https://github.com/openshift/jenkins-plugin) | [Jenkins Repo](https://github.com/jenkinsci/openshift-pipeline-plugin) | [Wiki](https://wiki.jenkins.io/display/JENKINS/OpenShift+Pipeline+Plugin) |

The *OpenShift Pipeline Plugin* is only in maintenance for the 3.x stream and is not supported in 4.x.  The message to use the client plugin instead seems to have been received, and it has been quite a while since we have seen any github or bugzilla 
activity with it.

As the development process for each of these plugins are similar in many ways, we are documenting the process here under 
the repository for the OpenShift Jenkins images, where each of the plugins will reference this document in their respective 
READMEs.

## Constructing your development sandbox

You will need an environment with the following tools installed:

* Maven (the `mvn` command)
* Git
* Java (need v8)
* an IDE or editor of your choosing

The Jenkins open source project already has a bunch of useful links on setting up your development environment, including 
launching Jenkins from the IDE (though there are pros and cons for developing this way vs. developing against a Jenkins 
server running in an OpenShift pod).

Here are a few of those links for reference:

* [https://wiki.jenkins.io/display/JENKINS/Extend+Jenkins](https://wiki.jenkins.io/display/JENKINS/Extend+Jenkins) (note, there are many useful child pages under this wiki page)  
* [https://wiki.jenkins.io/display/JENKINS/Plugin+tutorial](https://wiki.jenkins.io/display/JENKINS/Plugin+tutorial) (one of the key child pages under Extend Jenkins)
* [https://wiki.jenkins.io/display/JENKINS/Setting+up+Eclipse+to+build+Jenkins](https://wiki.jenkins.io/display/JENKINS/Setting+up+Eclipse+to+build+Jenkins) (has concepts common to other IDEs undoubtedly)

### Basic code development flow

Our plugins are constructed such that if they are running in an OpenShift pod, they can determine how to connect to the 
associated OpenShift master automatically, and no configuration of the plugin from the Jenkins console is needed.

If you choose to run an external Jenkins server and you would like to test interaction with an OpenShift master, you will need to manually configure the plugin.  See each plugin's README or the OpenShift documentation for the specifics.

An example flow when running in an OpenShift pod:

1. Clone this git repository:
    ```
    git clone https://github.com/openshift/<plugin in question>-plugin.git
    ```
1. Enable the provided git hooks: see https://github.com/openshift/jenkins/blob/master/README.md#plugin-installation-for-centos7-v4.10+
   ```
   git config core.hooksPath .githooks/
   ```
1. In the root of the local repository, run maven 
    ```
    cd <plugin in question>-plugin
    mvn
    ```
1. Maven will build target/<plugin name>.hpi  (the Jenkins plugin binary)
1. Open Jenkins in your browser, log in as an administrator, and navigate as follows:
1. Manage Jenkins > Manage Plugins.
1. Select the "Advanced" Tab.
1. Find the "Upload Plugin" HTML form and click "Browse".
1. Find the <plugin name>.hpi built in the previous steps.
1. Submit the file.
1. Check that Jenkins should be restarted.



### Additional options for updating the plugin in Jenkins (when running in OpenShift)

Aside from updating the plugin via the Jenkins plugin manager, OpenShift's capabilities provide various
means for building a new image containing your plugin and pointing your jenkins deployment to that image.
For example, consider templates such as [this one](example-plugin-development-template.yaml).

Or you can ...

#### ... mimic the PR testing done in OpenShift CI/CD implementation

What follows in ["Actual PR Testing"](#actual-pr-testing) is a complete description of the build and test flows for PRs against the [OpenShift/Jenkins Images repo](https://github.com/openshift/jenkins) and 
the three plugins we provide to facilitate OpenShift/Jenkins integration.  In the end of that chapter will detail the tests run in the PR and how you can run them against your local clusters.

## Actual PR Testing

Each plugin repository, and the images repository, under [https://github.com/openshift](https://github.com/openshift) is under the umbrella of the Prow based OpenShift CI/CD infrastructure.  As such, there is 

1. an `openshift-ci-robot` bot that accepts [these commands](https://go.k8s.io/bot-commands?repo=openshift%2Forigin) from within a PR
1. a [code review process](https://github.com/kubernetes/community/blob/master/contributors/guide/owners.md#the-code-review-process) that leverages the aforementioned commands
1. OpenShift CI/CD's [specific Prow configuration](https://github.com/openshift/release) for the various repositories of the project, detailed below

### Prow configuration details

If you look at the [CI workflow configuration summary](https://github.com/openshift/release#ci-workflow-configuration), the `ci-operator/config`, `ci-operator/jobs`, and `ci-operator/templates` bullets there pertain in some fashion to the Jenkins related repositories.  Specifics within that context:

* For the plugins, we've managed to only need a `master` branch to date, and only `ci-operator/config` and `ci-operator/jobs` artifacts have had to be created for the three plugins
* For master and 4.x branches of the jenkins images out of [the OpenShift/Jenkins repo](https://github.com/openshift/jenkins), we now only need `ci-operator/config` and `ci-operator/jobs` artifacts.  
* For 3.11, we had to also compose a special template under `ci-operator/templates`
* For branches prior to 3.11, the only tests performed are some basic image startup and `s2i` invocations that were the equivalent of running `make test` from the local copy of the repository on your workstation.  They are still Jenkins based and not Prow based.  No tests were executed in an actual OpenShift environment.  At this point, only changes to address support cases are going into branches older than 3.11.

#### Jenkins images

Quick preamble ... some cool facts about OpenShift CI:
* it runs on OpenShift
* it leverages development features like Builds and ImageStreams extensively

##### 3.11 specifics

The ci-operator based configuration, the `ci-operator/config`, for PR testing in 3.11 is at [https://github.com/openshift/release/blob/master/ci-operator/config/openshift/jenkins/openshift-jenkins-openshift-3.11.yaml](https://github.com/openshift/release/blob/master/ci-operator/config/openshift/jenkins/openshift-jenkins-openshift-3.11.yaml).

The parts at the beginning and end of that file are more generic setup needs for [ci-operator](https://github.com/openshift/ci-operator).

What makes 3.11 a bit more complicated than 4.0 is
* During the 3.11 timeframe, `ci-operator` was pretty brand new.  In particular, updating image streams on the test cluster was not baked into the system... enhancements came during 4.0 that helped there
* 3.x did not create image streams for the slave/agent example images we ship

With that background, let's examine the relevant parts of [https://github.com/openshift/release/blob/master/ci-operator/config/openshift/jenkins/openshift-jenkins-openshift-3.11.yaml](https://github.com/openshift/release/blob/master/ci-operator/config/openshift/jenkins/openshift-jenkins-openshift-3.11.yaml).

First, we make a copy of the test container images and store it as ImagStreamTag `cluster-tests`:

```
  cluster-tests:
    cluster: https://api.ci.openshift.org
    name: origin-v3.11
    namespace: openshift
    tag: tests
```

Then, we are going to build a new test container image with: 

```
- from: cluster-tests
  inputs:
    src:
      paths:
      - destination_dir: .
        source_path: /go/src/github.com/openshift/jenkins/test-e2e/.
  to: tests

```

which employs an OpenShift Docker Strategy Build, using this [Dockerfile](https://github.com/openshift/jenkins/blob/master/test-e2e/Dockerfile) for the docker build.  And the buid is adding a [script](https://github.com/openshift/jenkins/blob/master/test-e2e/tag-in-image.sh) that will tag the Jenkins images created during the PRs build into the test cluster's jenkins imagestream in the openshift namespace.  The `to: tests` line updates the existing test image.  It is literally storing the output image from the OpenShift Docker Strategy Build into the `tests` ImageStreamTag in the CI systems's internal ImageStreams.  

Next, this stanza is for the main jenkins image:

```
- dockerfile_path: Dockerfile
  from: base
  inputs:
    src:
      paths:
      - destination_dir: .
        source_path: /go/src/github.com/openshift/jenkins/2/.
  to: 2-centos

```

Let's dive into this stanza:
* As with updating the test image, this yaml is short hand for defining an OpenShift Docker Build Strategy build, and it does a Docker build with:
    * `from: base` is an imagestreamtag to the latest 3.11.x OpenShift CLI image ... that gets substituted into the Dockerfile's FROM clause for the Docker build
    * `Dockerfile` means we will literally using the `Dockerfile` at the `source_path`, where the `/go/src/github.com/openshift/jenkins/2/.` corresponds to the `git checkout` for the git branch we are testing.
    * the `to: 2-centos` tells `ci-operator` to set the output of the Docker Build Strategy build to an imagestreamtag named `2-centos` in the test artifacts.  Then, `ci-operator` takes all such imagestremtag names, converts them to upper case, and then sets `IMAGE_<TAG NAME>` environment variables into the test system
* There are similar stanzas for the `slave-base`, `agent-maven-3.5`, and `agent-nodejs-8` images.  

But again, since we do not have imagestreams defined for those slave/agent images in 3.x, we have to do some more manipulation of the Prow setup. With that, let's move onto the `ci-operator/jobs` definitions for 3.11.

The key two files are the [presubmit definition](https://github.com/openshift/release/blob/master/ci-operator/jobs/openshift/jenkins/openshift-jenkins-openshift-3.11-presubmits.yaml) and the [postsubmit definition](https://github.com/openshift/release/blob/master/ci-operator/jobs/openshift/jenkins/openshift-jenkins-openshift-3.11-postsubmits.yaml).

The most relevant pieces of the presubmit (aside from a bunch of Prow and ci-operator gorp you can dismiss):
* Whether a PR related job is Prow based or Jenkins based is indicated by the `agent` setting.  In the presubmits, they are all Prow based:

```
  - agent: kubernetes
```

* There are unique Prow definitions for each branch a given repo.  So this stanza signifies that this handles PRs for the `openshift-3.11` branch of the jenkins repo:

```
    branches:
    - openshift-3.11
```

* If the building of the image fails for some reason within a PR (`yum mirror` flake during RPM install, Jenkins update center flake during plugin download, a bug in your PR), you can re-run the OpenShift Docker Strategy build defined in the config via a `/test image` comment made to the PR, which maps to the ci-operator definition with the stanza:

```
    rerun_command: /test images
```

* To kick off a another run of the PR tests, you can type in `/test e2e-gcp` as a PR comment because of:

```
    rerun_command: /test e2e-gcp
```

* To trigger the tagging of the PRs newly build Jenkins image into the test cluster, this stanza sets up an environment variable that allows code down the line to call that script we mounted into the test container earlier:

```
        - name: PREPARE_COMMAND
          value: tag-in-image.sh
```


* As an optimization, we only run the Jenkins extended tests defined in https://github.com/openshift/origin.  This is achieved via:

```
        - name: TEST_COMMAND
          value: TEST_FOCUS='openshift pipeline' TEST_PARALLELISM=3 run-tests
```

OpenShift extended tests are based on the Golang Ginkgo test framework.  We leverage the focus feature of Ginkgo to limit the tests executed to the ones pertaining the our OpenShift/Jenkins integrations.  The `run-tests` executable actually launches the Ginkgo tests against the test cluster, and will leverage `TEST_FOCUS`.  The `TEST_PARALLELISM=3` is also worth noting, as it is unique to Jenkins e2e.  We discovered that the amount of memory Jenkins needed to run Pipelines was such that we could overload our 3.11 based GCP clusters used for CI.  Setting this environment variable restricted the amount of Jenkins based tests that would be run concurrently. 

The most relevant pieces of the postsubmit:
* There is 1 Prow based and 1 Jenkins based task in the postsubmits.

* The Prow based one gathers artifacts from the various image and test jobs for analysis.  Aside from debugging failures, the artifacts will prove helpful when we are updating the versions of any of the plugins the image depends on (more on that below).

* The Jenkins based job is a legacy from our pre-Prow, pre-4.x days, that still runs on the OpenShift CI Jenkins server at https://ci.openshift.redhat.com/jenkins.  It pushes new versions of the 3.11 images to docker.io.  With 4.0, we are only pushing the community versions of the images to quay.io. 


OK, now with `ci-operator/config` and `ci-operator/jobs` covered for 3.11, `ci-operator/templates` is the next item to cover.  Reminder, the need for jenkins to define a template is unique to 3.11.

The template is at [https://github.com/openshift/release/blob/master/ci-operator/templates/openshift/openshift-ansible/cluster-launch-e2e-openshift-jenkins.yaml](https://github.com/openshift/release/blob/master/ci-operator/templates/openshift/openshift-ansible/cluster-launch-e2e-openshift-jenkins.yaml).  And by template, yes, we mean an OpenShift template used to create API objects, where parameters are defined to vary the specific settings of those API objects.  The template to a large degree is a clone of https://github.com/openshift/release/blob/master/ci-operator/templates/openshift/openshift-ansible/cluster-launch-e2e.yaml, but with additions to:
* Tag in the jenkins image from the PR's build into the test cluster's jenkins imagestream
* Change the image pull spec of our example maven and nodejs agent images used to create our default k8s plugin pod templates configuration 

To tag in the Jenkins image, we added a `prepare` container into the e2e pod.  It leverages the `PREPARE_COMMAND` and `IMAGE_2_CENTOS` environment variables noted above to call that script from the Jenkins repo that was added to the base test container from the CI system.

To update the agent images used, the `IMAGE_MAVEN_AGENT` and `IMAGE_NODEJS_AGENT` environment variables, which the `ci-operator/config` noted above sets to the pull spec of the newly built images from the PR, are read in by the extended tests in https://github.com/openshift/origin.  Those values in turn are used to set the [environment variables](https://github.com/openshift/jenkins#environment-variables) that the Jenkins image recognizes for overriding the default images used for the k8s plugin configuration.

##### 4.x specifics

So, by comparison, the `ci-operator/config` for the master branch is [here](https://github.com/openshift/release/blob/master/ci-operator/config/openshift/jenkins/openshift-jenkins-master.yaml).  As versions of 4.x accumulate, you'll see more `openshift-jenkins-<branch name>.yaml` files in that directory.  As of this writing, the initial GA of 4.x has not occurred.

Key differences from 3.11:
* The Prow based infra evolved in many ways....like leveraging AWS vs. GCP for example.  But it got to the point also, where the mapping to imagestreams running in the test cluster was more direct.  So in a stanza like:

```
- dockerfile_path: Dockerfile.rhel7
  from: base
  inputs:
    src:
      paths:
      - destination_dir: .
        source_path: /go/src/github.com/openshift/jenkins/2/.
to: jenkins
```

that still executed an OpenShift Docker Strategy build against the `git checkout` from the PR, but now the `to: jenkins` line meant the system would tag our resulting image directly into the test cluster's `jenkins` imagestream.

So no more need for the adding in [the tagging script](https://github.com/openshift/jenkins/blob/master/test-e2e/tag-in-image.sh) into the test container and doing special test cluster setup in the template.

Also, with the agent images now having imagestreams in a 4.x cluster as part of being part of the installed payload, we can leverage the same pattern for those as well, simply tagging in the PR's versions of those images into the test cluster.  Hence, no need for the `IMAGE_MAVEN_AGENT` and `IMAGE_NODEJS_AGENT` environment variables.

With both those removed, we no longer needed a special template under `ci-operator/templates` for jenkins.  We now use the default one in 4.x.

You will also notice the use of `Dockerfile.rhel7` vs. `Dockerfile` for the `dockerfile_path`.  This stems from the [move in 4.x to the UBI](https://github.com/openshift/jenkins#installation-openshift-v4) and the end of CentOS based content for OpenShift.

The extended tests defined in [OpenShift origin](https://github.com/openshift/origin) were also reworked in 4.x.  The use of Ginkgo focuses were moved to only within the test executables, and can no longer be specified from the command line.  "Test suites" were defined for the most prominent focuses, including one for Jenkins called `openshift/jenkins-e2e`.  This simplification, along with other rework, makes it easier to run the extended tests against existing clusters (including ones you stand up for your development ... more on that later).  
The tests defined for jenkins  are as of 4.10 a "meets min" validation that JenkinsPipelineStrategy builds work.  And they are only invoked if you run `e2e-aws-jenkins` in openshift/origin PRs or openshift/cluster-samples-operator PRs.  The more precise jenkins e2e's centering on the plugins are now defined in the sync and client plugins respectively (and are "controller like" vs. ginkgo).
The generic OpenShift conformance regression bucket is also still included:

```
tests:
- as: e2e-aws-jenkins-sync-plugin
  skip_if_only_changed: ^docs/|\.md$|^(?:.*/)?(?:\.gitignore|OWNERS|PROJECT|LICENSE)$
  steps:
    cluster_profile: aws
    test:
    - ref: jenkins-sync-plugin-e2e
    workflow: ipi-aws
- as: e2e-aws-jenkins-client-plugin
  skip_if_only_changed: ^docs/|\.md$|^(?:.*/)?(?:\.gitignore|OWNERS|PROJECT|LICENSE)$
  steps:
    cluster_profile: aws
    test:
    - ref: jenkins-client-plugin-tests
    workflow: ipi-aws
```

Those jenkins plugin e2e's are also used in their respective repos' PRs.  The are defined in the [CI Step Registry](https://github.com/openshift/release/tree/master/ci-operator/step-registry/jenkins)

Moving on to the `ci-operator/jobs` data, not as different from 3.11 as the `ci-operator/config`, [the presubmits](https://github.com/openshift/release/blob/master/ci-operator/jobs/openshift/jenkins/openshift-jenkins-master-presubmits.yaml) are the ci-operator and Prow related definitions for the `e2e-aws` and `e2e-aws-jenkins` test jobs noted above, as well as the job to build the images.  Each can be re-run via `/test e2e-aws`, `/test e2e-aws-jenkins`, or `/test images`.

Likewise, the [postsubmit](https://github.com/openshift/release/blob/master/ci-operator/jobs/openshift/jenkins/openshift-jenkins-master-postsubmits.yaml) has the Prow job defined to collect the test artifacts as with 3.11.  The Jenkins based job to push the resulting images has been removed as part of the move off of docker.io and onto quay.io.  The CI system in general will push the updates to quay.io for all relevant images, and we no longer need special jobs for our Jenkins images on this front. 

#### Jenkins plugins

For the plugins, we've managed to only need a `master` branch to date in providing support against 3.x and 4.x.  The OpenShift features needed to support the various OpenShift/Jenkins integrations landed in 3.4, and we are only actively supporting OpenShift/Jenkins images back to 3.6.  An experimental branch `v3.6-fixes` was created in https://github.com/jenkinsci/openshift-sync-plugin a while back to confirm we *could* backport specific fixes to older versions if need be and craft versions like `1.0.24.1` if need be.  But for simplicity that will be a last resort sort of thing. 

Also note, we have not been updating older versions of the OpenShift/Jenkins image with newer versions of our 3 plugins to pull in fixes unless a support case comes in through bugzilla that dictates so.

With that background as context:
* During the 4.0 time frame PR testing for the supported plugins has migrated fully to Prow (and away from the OpenShift CI Jenkins server)
* Only `ci-operator/config` and `ci-operator/jobs` artifacts are needed for the three plugins, similar to what exists for the [jenkins repo](https://github.com/openshift/jenkins).
* The [deprecated plugin](https://github.com/openshift/jenkins-plugin) remains on jobs defined at [https://github.com/openshift/aos-cd-jobs](https://github.com/openshift/aos-cd-jobs) that run on the OpenShift CI Jenkins server ... note since its EOL notice on Aug 3 2018 there have been no further changes needed to that plugin from a support perspective.
  
The `ci-operator/config` files for each plugin are very similar.  For reference, here are the locations for the [sync plugin](https://github.com/openshift/release/blob/master/ci-operator/config/openshift/jenkins-sync-plugin/openshift-jenkins-sync-plugin-master.yaml), the [client plugin](https://github.com/openshift/release/blob/master/ci-operator/config/openshift/jenkins-client-plugin/openshift-jenkins-client-plugin-master.yaml), and [login plugin](https://github.com/openshift/release/blob/master/ci-operator/config/openshift/jenkins-openshift-login-plugin/openshift-jenkins-openshift-login-plugin-master.yaml).

The 3 key elements are:
* We define a "copy" of the latest jenkins imagestream to serve as the basis of the new image build that contains the new plugin binary stemming from the PR's changes:

```
base_images:
  original_jenkins:
    name: '4.0'
    namespace: ocp
    tag: jenkins

```

* With that as the base image, like before, we define an OpenShift Docker Strategy build against the given plugin's checked out repo, leveraging the new 4.x specific `Dockerfile` files present for each.  The `ci-operator/config` for that build is:

```
images:
- dockerfile_path: Dockerfile
  from: original_jenkins
  inputs:
    src:
      paths:
      - destination_dir: .
        source_path: /go/src/github.com/openshift/jenkins-openshift-login-plugin/.
to: jenkins
```

Where `source_path` and `dockerfile_path` correspond to the `git checkout` of the plugin repo for the PRs branch, and `to: jenkins` says take the resulting image and tag it into the jenkins imagestream.

The `Dockerfile` files are like:

```
FROM quay.io/openshift/origin-jenkins-agent-maven:v4.0 AS builder
WORKDIR /java/src/github.com/openshift/jenkins-login-plugin
COPY . .
USER 0
RUN export PATH=/opt/rh/rh-maven35/root/usr/bin:$PATH && mvn clean package

FROM quay.io/openshift/origin-jenkins:v4.0
RUN rm /opt/openshift/plugins/openshift-login.jpi
COPY --from=builder /java/src/github.com/openshift/jenkins-login-plugin/target/openshift-login.hpi /opt/openshift/plugins
RUN mv /opt/openshift/plugins/openshift-login.hpi /opt/openshift/plugins/openshift-login.jpi
```

Where the second `FROM ..` is replaced with the image pull spec for `from: original_jenkins`, and we use `AS builder` and our maven image to compile the plugin using the PR's branch, and then copy the resulting hpi file into the Jenkins image's plugin directory, so it is picked up when the images is started up.

* And lastly, we leverage just the jenkins-e2e suite for plugin testing with:

```
tests:
- as: e2e-aws-jenkins
  commands: TEST_SUITE=openshift/jenkins-e2e run-tests
  openshift_installer:
cluster_profile: aws
```

The use of the generic conformance regression bucket is omitted for plugin testing, as it will be covered when we attempt to update the openshift/jenkins image with any new version of the plugin.

(A subtle reminder that what we *officially support* is the openshift/jenkins images.  New releases of the plugin not yet incorporated into the image are considered "pre-release function").

The `ci-operator/jobs` files are very similar to the ones for the jenkins image itself for 4.x, in the Prow jobs they define, etc.  For reference here are the locations of the [client plugin](https://github.com/openshift/release/tree/master/ci-operator/jobs/openshift/jenkins-client-plugin), [login plugin](https://github.com/openshift/release/tree/master/ci-operator/jobs/openshift/jenkins-openshift-login-plugin), and [sync plugin](https://github.com/openshift/release/tree/master/ci-operator/jobs/openshift/jenkins-sync-plugin).

### Extended tests

##### 4.11 and later

- [meets min jenkins pipeline strategy verification in openshift/origin, only for openshift/origin and openshift/cluster-samples-operator jobs](https://github.com/openshift/origin/blob/master/test/extended/builds/pipeline_origin_bld.go)
- [client plugin tests](https://github.com/openshift/jenkins-client-plugin/tree/master/test/e2e)
- [sync plugin tests](https://github.com/openshift/jenkins-sync-plugin/tree/master/test/e2e)
- login plugin uses the sync plugin tests, as it always fetches logs using the login plugin's command line with oauth token flow

##### 4.19 and earlier

First, the extended tests in [OpenShift Origin](https://github.com/openshift/origin) that we've made some references to ... where are the Jenkins ones specifically?  There are two golang files:
* The [reduced jenkins e2e suite](https://github.com/openshift/origin/blob/master/test/extended/builds/pipeline_origin_bld.go) run in the OpenShift Build's regression bucket
* In addition to the minimal suite, [these tests](https://github.com/openshift/origin/blob/master/test/extended/builds/pipeline_jenkins_e2e.go) get run in the jenkins e2e for the image and each plugin

The structure of those golang files are a bit unique from a Ginkgo perspective, and as compared to the other tests in OpenShift origin.  Given the memory demands of Jenkins, as well as intermittent issues with Jenkins trying to contact the Update Center, we've gone through pains:
* To minimize the number of Jenkins instance running concurrently during the e2e.
* To minimize the number of times we have to bring Jenkins up

So you'll see less use of `g.Context(..)` and `g.It(...)`, as well as cleaning up of resources between logical tests.  Currently the divisions in the tests that result in concurrent test runs are between:
* The client plugin and sync plugin
* The ephemeral storage template and persistent storage template

The `openshift-tests` binary in the 4.x branches of [OpenShift Origin](https://github.com/openshift/origin) includes those tests (and as it turns out, can be run against both 3.11 and 4.x clusters).  Once you have a cluster up, and the `openshift-tests` binary built (run `hack/build-go.sh cmd/openshift-tests` from you clone of origin), you can:
* set and export KUBECONFIG to the location of the admin.kubeconfig for the cluster
* run `openshift-tests run openshift/jenkins-e2e --include-success` against the cluster ... the `openshift/jenkins-e2e` is considered a "suite" in `openshift-tests` and under the covers it leverages Ginkgo focuses to pick up the tests from those two golang files.

To run extended tests against one of your clusters using a set of local changes of a plugin, from one of the plugin repo's top dirs, you can:
* run `docker build -f ./Dockerfile -t <a publicly accessible docker registry spec, like docker.io/gmontero/plugin-tests:latest>`
* run `docker push <a publicly accessible docker registry spec, like docker.io/gmontero/plugin-tests:latest>`
* run `oc tag --source=docker <a publicly accessible docker registry spec, like docker.io/gmontero/plugin-tests:latest> openshift/jenkins:2`
* run `openshift-tests run openshift/jenkins-e2e --include-success`  ... the imagestream controller in OpenShift will pull the publicly accessible docker registry spec, like docker.io/gmontero/plugin-tests:latest, when the standard jenkins template is provisioned. 

### the "PR-Testing" folders in the plugins

The Dockerfiles and scripts in these folders were used in the pre-Prow days, where the Jenkins jobs would build Docker images with the updated plugin local to the test nodes, and then via environment variables, the extended tests would provision Jenkins either using the standards templates, or specialized ones that would leverage the local test image.

We can most likely remove these, but are holding on in case we have to create non-master branches for back porting fixes to older versions of the plugins in corresponding older openshift/jenkins images.

## Process for cutting a release of a plugin

Once we've merged changes into one of the OpenShift org GitHub repositories for a given plugin, we need to transfer the associated commit to the corresponding JenkinsCI org GitHub repository and follow [the upstream Jenkins project release process](https://wiki.jenkins.io/display/JENKINS/Hosting+Plugins) when we have deemed changes suitable for inclusion into the non-subscription OpenShift Jenkins image (the CentOS7 based one hosted on docker.io for 3.x, the UBI based one hosted on quay.io for 4.x).  

### Accounts and configuration (both local and upstream in various remote, Jenkins related resources) 

Some key specifics from [the upstream Jenkins project release process](https://wiki.jenkins.io/display/JENKINS/Hosting+Plugins):

* You need a login/account via [https://accounts.jenkins.io/](https://accounts.jenkins.io/) .... by extension it should also give you access to [https://issues.jenkins-ci.org](https://issues.jenkins-ci.org).  See [https://wiki.jenkins-ci.org/display/JENKINS/User+Account+on+Jenkins](https://wiki.jenkins-ci.org/display/JENKINS/User+Account+on+Jenkins).
* You should add this account to your `~/.m2/settings.xml`.  The release process noted above has details on how to do that, as well as workarounds for potential hiccups.  Read thoroughly.
* Someone on the OpenShift Developer Experience team who already has access will need to give you the necessary permissions in the files for each plugin at [https://github.com/jenkins-infra/repository-permissions-updater/tree/master/permissions](https://github.com/jenkins-infra/repository-permissions-updater/tree/master/permissions).  Existing users will need to construct a PR that adds your ID to the list of folks.
* Similarly for the [https://github.com/jenkinsci](https://github.com/jenkinsci) repositories for each plugin, we'll need to update each repositories administrator lists, etc.
with your github ID. We currently have these github teams defined:
one has to be a member of the github teams:
    * [https://github.com/orgs/jenkinsci/teams/openshift-sync-plugin-developers](https://github.com/orgs/jenkinsci/teams/openshift-sync-plugin-developers)
    * [https://github.com/orgs/jenkinsci/teams/openshift-login-plugin-developers](https://github.com/orgs/jenkinsci/teams/openshift-login-plugin-developers)
    * [https://github.com/orgs/jenkinsci/teams/openshift-client-plugin-developers](https://github.com/orgs/jenkinsci/teams/openshift-client-plugin-developers)

### Where the officially cut versions of each plugin are hosted

For our Jenkins image repository to include particular versions of our plugins in the image, the plugin versions in question
need to be available at these locations, depending on the particular plugin of course.  These are the official landing spots 
for a newly minted version of a particular plugin.

* [https://updates.jenkins.io/download/plugins/openshift-client](https://updates.jenkins.io/download/plugins/openshift-client)
* [https://updates.jenkins.io/download/plugins/openshift-sync](https://updates.jenkins.io/download/plugins/openshift-sync)
* [https://updates.jenkins.io/download/plugins/openshift-login](https://updates.jenkins.io/download/plugins/openshift-login)

We as of yet have not had to pay attention to them, but the CI jobs over on CloudBee's Jenkins server for our 4 plugins are:

* [https://jenkins.ci.cloudbees.com/job/plugins/job/openshift-client-plugin/](https://jenkins.ci.cloudbees.com/job/plugins/job/openshift-client-plugin/)
* [https://jenkins.ci.cloudbees.com/job/plugins/job/openshift-sync-plugin/](https://jenkins.ci.cloudbees.com/job/plugins/job/openshift-sync-plugin/)
* [https://jenkins.ci.cloudbees.com/job/plugins/job/openshift-login-plugin/](https://jenkins.ci.cloudbees.com/job/plugins/job/openshift-login-plugin/)

These kick in when we cut the official version at the Jenkins Update Center for a given plugin.


### Set up local repository to cut release:

To cut a new release of any of our plugins, you will set up a local clone of the [https://github.com/jenkinsci](https://github.com/jenkinsci) repository for the plugin in question, like [https://github.com/jenkinsci/openshift-client-plugin](https://github.com/jenkinsci/openshift-client-plugin),
and then transfer the necessary commits from the corresponding [https://github.com/openshift](https://github.com/openshift) repository, like [https://github.com/openshift/jenkins-client-plugin](https://github.com/openshift/jenkins-client-plugin).

### Transfer commits from the [https://github.com/openshift](https://github.com/openshift) repositories to the [https://github.com/jenkinsci](https://github.com/jenkinsci) repositories ... prior to generating a new plugin release

In your clone of `https://github.com/jenkinsci/<plugin dir>`, set up your git remotes so origin is the `https://github.com/jenkinsci/<plugin dir>` repository, and upstream is the `https://github.com/openshift/<plugin dir>/` repository.  Using `openshift-client` as an example (and substitute the other plugin names if working with those plugins):

1. From the parent directory you've chosen for you local repository, clone it ... for example, run `git clone git@github.com:jenkinsci/openshift-client-plugin.git` for the client plugin
1. Change into the resulting directory, again for example `openshift-client-plugin`, and add an git upstream link for the corresponding repo under openshift ...for example, run `git remote add upstream git://github.com/openshift/jenkins-client-plugin` for the client plugin
1. Then pull and rebase the latest changes from [https://github.com/openshift/jenkins-client-plugin](https://github.com/openshift/jenkins-client-plugin) with the following:

```
	$ git checkout master	
	$ git fetch upstream	
	$ git fetch upstream --tags	
	$ git cherry-pick <commit id>  # for each commit that needs to be migrated	
	$ git push origin master	
	$ git push origin --tags
```

### Submit the new release to the Jenkins organization

After pushing the desired commits to the [https://github.com/jenkinsci](https://github.com/jenkinsci) repository for the plugin in question, you can now actually initiate the process to create a new version of the plugin in the Jenkins Update Center.

Prerequisite: your Git ID should have push access to the [https://github.com/jenkinsci](https://github.com/jenkinsci) repositories for this plugin; your Jenkins ID (again see [https://wiki.jenkins-ci.org/display/JENKINS/User+Account+on+Jenkins](https://wiki.jenkins-ci.org/display/JENKINS/User+Account+on+Jenkins)) is listed in the permission file for the plugin, like [https://github.com/jenkins-infra/repository-permissions-updater/blob/master/permissions/plugin-openshift-pipeline.yml](https://github.com/jenkins-infra/repository-permissions-updater/blob/master/permissions/plugin-openshift-pipeline.yml).  Given these assumptions:

1. Then run `mvn release:prepare release:perform`
1. You'll minimally be prompted for the `release version`, `release tag`, and the `new development version`.  Default choices will be provided for each, and the defaults are acceptable, so you can just hit the enter key for all three prompts.  As an example, if we are currently at v1.0.36, it will provide 1.0.37 for the new `release version` and `release tag`.  For the `new development version` it will provide 1.0.38-SNAPSHOT, which is again acceptable.  
1. The `mvn release:prepare release:perform` command will take a few minutes to build the plugin and go through various verifications, followed by a push of the built artifacts up to the Jenkins Artifactory server.  This typically works without further involvement but has failed for various reasons in the past.  If so, to retry with the same release version, you will need to call `git reset --hard HEAD~2` to back off the two commits created as part of publishing the release (the "release commits", where the pom.xml is updated to reflect the new version and the next snapshot version), as well as use `git tag` to delete both the local and remote version of the corresponding tag.  After deleting the commits and tags, use `git push -f` to update the commits at the Jenkinsci GitHub Org repo in question. Address whatever issues you have (you might have to solicit help on the Jenkins developer group: [https://groups.google.com/forum/#!forum/jenkinsci-dev](https://groups.google.com/forum/#!forum/jenkinsci-dev)) or on the #jenkins channel on freenode (Daniel Beck from CloudBees has been helpful and returned message on #jenkins), then try again.
1. If `mvn release:prepare release:perform` completes successfully, those "release commits" will look something like this if you ran `git log -2`:

```
commit 1c6dabc66c24c4627941cfb9fc2a53ccb0de59b0
Author: gabemontero <gmontero@redhat.com>
Date:   Thu Oct 26 14:18:52 2017 -0400

    [maven-release-plugin] prepare for next development iteration

commit e040110d466249dd8c6f559e343a1c6b4b5f19a8
Author: gabemontero <gmontero@redhat.com>
Date:   Thu Oct 26 14:18:48 2017 -0400

    [maven-release-plugin] prepare release openshift-login-1.0.0
```

### Transfer commits from the [https://github.com/jenkinsci](https://github.com/jenkinsci) repositories to the [https://github.com/openshift](https://github.com/openshift) repositories ... after generating a new plugin release

Keeping the commit lists between the openshift and jenkinsci repositories as close as possible helps with general sanity, as we do our development work on the openshift side, but have to cut the plugin releases on the jenkinsci side.  So we want to pick the 2 version commits back into our corresponding openshift repositories.

First, from you clone on the openshift side, say for example the clone for the sync plugin, run `git fetch git@github.com:jenkinsci/openshift-sync-plugin.git`

Second, create a new branch.

Third, pick the two version commits via `git cherry-pick`

Lastly, push the branch and create a PR against the https://github.com/openshift repo in question.  Quick FYI, the plugin repos are now set up with the right github rebase configuration options such that we do not get that extra, empty "merge commit".

### Testing:

A described above, we can employ our CI server and extended test framework for respositories in [https://github.com/openshift](https://github.com/openshift).  We cannot for [https://github.com/jenkinsci](https://github.com/jenkinsci).  For this reason we maintain the separate repositories, where we run OpenShift CI against incoming changes before publishing them to the "official" [https://github.com/jenkinsci](https://github.com/jenkinsci) repositories.

### Release cycle (when we get to the end of a release):

We can continue development of new features for the next release in the [https://github.com/openshift](https://github.com/openshift) repositories.  But if a new bug for the current release
pops up in the interim, we can make the change in [https://github.com/jenkinsci](https://github.com/jenkinsci) first, then cherry-pick into [https://github.com/openshift](https://github.com/openshift) (employing
our regression testing), and then when ready cut the new version of the plugin in [https://github.com/jenkinsci](https://github.com/jenkinsci).

### What version is a change really in:

Through various scenarios and human error, those "release commits" (where the version in the pom.xml is updated) sometimes have landed in [https://github.com/openshift](https://github.com/openshift) repositories after the fact, or out of order with
how they landed in [https://github.com/jenkinsci](https://github.com/jenkinsci) repositories.  The [https://github.com/jenkinsci](https://github.com/jenkinsci) repositories are the *master* in this case.  The plugin version a change is in is expressed by the commit orders in the pom.xml release changes in the various [https://github.com/jenkinsci](https://github.com/jenkinsci) repositories.


## FINALLY .... updating our OpenShift Jenkins images with the new plugin versions

### Step 1: updating via "api.ci", the OpenShift Prow CI system

As referred to previously, the new plugin version will land at `https://updates.jenkins.io/download/plugins/<plugin-name>`.  Monitor that page for the existence of the new version of the plugin.  Warning: the link for the new version can show up, but does not mean the file is available yet.  Click the link to confirm you can download the new version of the plugin.  When you can download the new version file, the new release is available.

At this point, we are back to the steps articulated in [the base plugin installation section of this repository's README](https://github.com/openshift/jenkins/blob/master/README.md#base-set-of-plugins).  You'll 
modify the [text file](https://github.com/openshift/jenkins/blob/master/2/contrib/openshift/base-plugins.txt) with the new version for whatever OpenShift plugin you have cut a new version for, and create a new PR.  The image build and e2e extended test runs we detailed in ["Actual PR Testing"](#actual-pr-testing) will commence.

If the PR passes all tests and merges, the api.ci system will promote the jenkins images to quay.io for 4.x, and we have the separate jenkins job in the openshift ci jenkins server to push the 3.11 images to docker.io.

The image build from the PR in particular is of interest when it comes to plugin versions within our openshift/jenkins image, and what we have to do in creating the RPM based images hosted on registy.redhat.io/registry.access.redhat.com that we provide subscription level support for.  The PR will have a link to the `ci/prow/images` job.  If you click that link, then the `artifacts` link, then the next `artifacts` link, then `build-logs`, you'll see gzipped output from each of the image builds.  Click the one for the main jenkins image.  If you search for the string `Installed plugins:` you'll find the complete list of every plugin that was installed.  Copy that output to clipboard and paste it into the PR that just merged.  See [https://github.com/openshift/jenkins/pull/829#issuecomment-477637521](https://github.com/openshift/jenkins/pull/829#issuecomment-477637521) as an example.

### Step 2 (4.11 and later)

The jenkins images got removed from the OCP install payload in 4.11, and we now use the CPaaS Images Akram created for us.  A net of all this is that the OCP samples operator no longer modifies the jenkins related ImageStreams with 
the image ref from the install payload.  Jenkins is now treated like any other sample.  With that though:
- we no longer have to submit ART requests after our openshift/jenkins PRs merge
- we do need to update our (ImageStreams)[https://github.com/openshift/jenkins/tree/master/openshift/imagestreams] with new image refs

Moving out of the payload, we also were able to create more than one ImageStreamTag, so we can now support some of the upgrade options customers have asked for in the past:
- still upgrade when OCP upgrades (we use a precise SHA image ref for this)
- use the OCP ImageStream scheduled import feature, so the ImageStream controllers periodically access registry.redhat.io if updates have been made
- use a constant Image ref for the ImageStreamTag, so that unless we change that ref in someway, the ImageStream update made by samples operator on an upgrade should no change the ImageStreamTag's image ref, and hence the Image Change Controller should 
not rollout a new Jenkins Deployment.  The customer than can control when Jenkins is redeployed by doing an 'oc import-image' for the jenkins ImageStream

Once the Jenkins ImageStreams in the repo are updated, wait for the "Automatic importer job update" commit in [OpenShift Library](https://github.com/openshift/library/commits/master) that pulls in our changes.

After that, open a [samples operator](https://github.com/openshift/cluster-samples-operator) PR that follows [this process](https://github.com/openshift/cluster-samples-operator#update-the-content-in-this-repository-from-httpsgithubcomopenshiftlibrary) for
updating the samples content for jenkins.  NOTE: that script will pull in all the samples updates.  If you only want to pull in jenkins, edit the git commits to undo those changes.  NOTE: be sure the run the optional `e2e-aws-jenkins` test job as a sanity check.

### Step 2 (4.10 and earlier): updating OSBS/Brew for generating the officially supported images available with Red Hat subscriptions

First, some background:  for all Red Hat officially supported content, to ensure protection from outside hackers, all content is built in a quarantined system with no access to the external Internet.  As such, we have to inject all content into OpenShift's Brew server (see links like [https://pagure.io/koji/](https://pagure.io/koji/)  and [https://osbs.readthedocs.io/en/latest/](https://osbs.readthedocs.io/en/latest/) if you are interested in the details/histories of this infrastructure), which is then scrubbed before official builds are run with it.  The injection is specifically the creation of an RPM which contains all the plugin binaries.

The team responsible for all this build infrastructure for OpenShift, the Automated Response Team or ART, not surprisingly has there own, separate, Jenkins instance that helps manage some of their dev flows.

They have provided us a Jenkins pipeline (fitting, I know) that facilitates the building of the plugin RPM and its injection into the Brew pipeline that ultimate results in an image getting built.  They have also provided a pipeline for updating the version of Jenkins core we have installed.  The pipeline for updating the version of the Jenkins core in our image is at [https://github.com/openshift/aos-cd-jobs/tree/master/jobs/devex/jenkins-bump-version](https://github.com/openshift/aos-cd-jobs/tree/master/jobs/devex/jenkins-bump-version) and the pipeline for updating the set of plugins installed is at [https://github.com/openshift/aos-cd-jobs/tree/master/jobs/devex/jenkins-plugins](https://github.com/openshift/aos-cd-jobs/tree/master/jobs/devex/jenkins-plugins).

Now, we *used* to be able to log onto their Jenkins server and initiate runs of those 2 pipelines, but towards the end of the 4.1 cycle, corporate processes and guidelines changed such that only members of the ART are allowed to access it.

So now, you need to open a Jira bug on the ART board at [https://jira.coreos.com/secure/RapidBoard.jspa?rapidView=85](https://jira.coreos.com/secure/RapidBoard.jspa?rapidView=85) to inform them of what is needed.  Supply these parameters:
* The jenkins core version to base off of (just supply what we are shipping with 3.11 or 4.x)
* The "OCP_RELEASE" is the OpenShift release (OCP == OpenShift Container Platform) ... so either 3.11, 4.0, 4.1, etc.
* The plugin list is the list you saved from the image build in the PR.  Remove the "Installed plugins" header, but include the `<plugin name>:<plugin version>` lines

An example of such a request is at [https://jira.coreos.com/browse/ART-673](https://jira.coreos.com/browse/ART-673).

The job typically takes 10 to 15 minutes to succeed.  Flakes with Jenkins upstream when downloading plugins is the #1 barrier to success.  Just need to retry again until the Jenkins update center stabilizes.  Once in a while there is a dist git hiccup (dist git is the git server used by brew).  Again, just try again until it settles down.

When the job succeeds, an email is sent to the folks in the mailing lists.  Add yours to the list when submitting the job if it is not listed there.  Submit a PR to update [https://github.com/openshift/aos-cd-jobs/blob/master/jobs/devex/jenkins-plugins/Jenkinsfile](https://github.com/openshift/aos-cd-jobs/blob/master/jobs/devex/jenkins-plugins/Jenkinsfile) to add your email if you'll be doing this long term.

The email sent will contain links that will point you to the dist git commit for the new RPM.

Store the job link and dist git link in the original PR, like [this comment](https://github.com/openshift/jenkins/pull/829#issuecomment-477686871).

AND THE LAST PIECE !!!!! .... the openshift/jenkins images produced by OpenShift's Brew are listed at [https://brewweb.engineering.redhat.com/brew/packageinfo?packageID=67183](https://brewweb.engineering.redhat.com/brew/packageinfo?packageID=67183).  Make sure the brew registry in in your docker config's insecure registry list (it is `brew-pulp-docker01.web.prod.ext.phx2.redhat.com:8888`).  If you are on the Red Hat network or VPN, you can download the images and try them out.  Also, QA/QE goes to this registry to verify bugs/features.

NOTE: when they build those images in brew, they actually modify the `Dockerfile.rhel7` files to facilitate:
* some **magic** with respect to the yum repositories used that is particular to running in this quarantined environment
* switch the `INSTALL_JENKINS_VIA_RPMS` so our build scripts do not attempt to download plugins, but rather install the RPMs

The file which triggers these things for the main image is at [https://github.com/openshift/ocp-build-data/blob/openshift-4.0/images/openshift-jenkins-2.yml](https://github.com/openshift/ocp-build-data/blob/openshift-4.0/images/openshift-jenkins-2.yml) for the 4.0 release.  There are branches for the other releases.  For the agent images you have [https://github.com/openshift/ocp-build-data/blob/openshift-4.0/images/jenkins-slave-base-rhel7.yml](https://github.com/openshift/ocp-build-data/blob/openshift-4.0/images/jenkins-slave-base-rhel7.yml), [https://github.com/openshift/ocp-build-data/blob/openshift-4.0/images/jenkins-agent-maven-35-rhel7.yml](https://github.com/openshift/ocp-build-data/blob/openshift-4.0/images/jenkins-agent-maven-35-rhel7.yml), and [https://github.com/openshift/ocp-build-data/blob/openshift-4.0/images/jenkins-agent-nodejs-8-rhel7.yml](https://github.com/openshift/ocp-build-data/blob/openshift-4.0/images/jenkins-agent-nodejs-8-rhel7.yml). 
