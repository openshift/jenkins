# Contributing

The OpenShift organization maintains several Github repositories that contain all of the source code for our Jenkins images and plugins.

|     |      |      |   |
| -------------- | --------------------------   | ------------------------   | -------------- |
|`Jenkins Server Image`|[OpenShift Repo](https://github.com/openshift/jenkins)|||
|`Jenkins Agent Base Image`|[OpenShift Repo](https://github.com/openshift/jenkins)|||
| `OpenShift Client Plugin` | [OpenShift Repo](https://github.com/openshift/jenkins-client-plugin) | [Jenkins Repo](https://github.com/jenkinsci/openshift-client-plugin) | [Release Website](https://plugins.jenkins.io/openshift-client/) |
| `OpenShift Sync Plugin` | [OpenShift Repo](https://github.com/openshift/jenkins-sync-plugin) | [Jenkins Repo](https://github.com/jenkinsci/openshift-sync-plugin) | [Release Website](https://plugins.jenkins.io/openshift-sync/) |
| `OpenShift Login Plugin` | [OpenShift Repo](https://github.com/openshift/jenkins-openshift-login-plugin) | [Jenkins Repo](https://github.com/jenkinsci/openshift-login-plugin) | [Release Website](https://plugins.jenkins.io/openshift-login/) |



As the development process for each of these plugins are similar in many ways, we are documenting the process here under
the repository for the OpenShift Jenkins images, where each of the plugins will reference this document in their respective
READMEs.

## Constructing your development sandbox

You will need an environment with the following tools installed:

* Maven (the `mvn` command)
* Git
* Java 17
* an IDE or editor of your choosing

### Basic code development flow

Our plugins are constructed such that if they are running in an OpenShift pod, they can determine how to connect to the
associated OpenShift master automatically, and no configuration of the plugin from the Jenkins console is needed.

If you choose to run an external Jenkins server and you would like to test interaction with an OpenShift master, you will need to manually configure the plugin.  See each plugin's README or the OpenShift documentation for the specifics.

An example flow when running in an OpenShift pod:

1. Clone the git repository:
    ```
    git clone https://github.com/openshift/<plugin name>-plugin.git
    ```
1. In the root of the local repository, run maven
    ```
    cd <plugin name>-plugin
    mvn clean package
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

### the "PR-Testing" folders in the plugins

The Dockerfiles and scripts in these folders were used in the pre-Prow days, where the Jenkins jobs would build Docker images with the updated plugin local to the test nodes, and then via environment variables, the extended tests would provision Jenkins either using the standards templates, or specialized ones that would leverage the local test image.

We can most likely remove these, but are holding on in case we have to create non-master branches for back porting fixes to older versions of the plugins in corresponding older openshift/jenkins images.

## Process for cutting a release of a plugin

Once we've merged changes into one of the OpenShift org GitHub repositories for a given plugin, we need to transfer the associated commit to the corresponding JenkinsCI org GitHub repository and follow [the upstream Jenkins project release process](https://wiki.jenkins.io/display/JENKINS/Hosting+Plugins) when we have deemed changes suitable for inclusion into the non-subscription OpenShift Jenkins image.

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

If the PR passes all tests and merges, the api.ci system will promote the jenkins images to quay.io.

The image build from the PR in particular is of interest when it comes to plugin versions within our openshift/jenkins image, and what we have to do in creating the RPM based images hosted on registy.redhat.io/registry.access.redhat.com that we provide subscription level support for.  The PR will have a link to the `ci/prow/images` job.  If you click that link, then the `artifacts` link, then the next `artifacts` link, then `build-logs`, you'll see gzipped output from each of the image builds.  Click the one for the main jenkins image.  If you search for the string `Installed plugins:` you'll find the complete list of every plugin that was installed.  Copy that output to clipboard and paste it into the PR that just merged.  See [https://github.com/openshift/jenkins/pull/829#issuecomment-477637521](https://github.com/openshift/jenkins/pull/829#issuecomment-477637521) as an example.

### Step 2

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
