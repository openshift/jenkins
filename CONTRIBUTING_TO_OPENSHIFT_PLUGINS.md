# Development and release process for the OpenShift Jenkins plugins

The OpenShift organization currently maintains several plugins related to integration between V3 OpenShift and V2 Jenkins.

|     |      |      |   |
| -------------- | --------------------------   | ------------------------   | -------------- |
| `OpenShift Client Plugin` | [OpenShift Repo](https://github.com/openshift/jenkins-client-plugin) | [Jenkins Repo](https://github.com/jenkinsci/openshift-client-plugin) | [Wiki](https://wiki.jenkins.io/display/JENKINS/OpenShift+Client+Plugin) |
| `OpenShift Sync Plugin` | [OpenShift Repo](https://github.com/openshift/jenkins-sync-plugin) | [Jenkins Repo](https://github.com/jenkinsci/openshift-sync-plugin) | [Wiki](https://wiki.jenkins.io/display/JENKINS/OpenShift+Sync+Plugin) |
| `OpenShift Login Plugin` | [OpenShift Repo](https://github.com/openshift/jenkins-openshift-login-plugin) | [Jenkins Repo](https://github.com/jenkinsci/openshift-login-plugin) | [Wiki](https://wiki.jenkins.io/display/JENKINS/OpenShift+Login+Plugin) |
| `OpenShift Pipeline Plugin` | [OpenShift Repo](https://github.com/openshift/jenkins-plugin) | [Jenkins Repo](https://github.com/jenkinsci/openshift-pipeline-plugin) | [Wiki](https://wiki.jenkins.io/display/JENKINS/OpenShift+Pipeline+Plugin) |

The *OpenShift Pipeline Plugin* is only in maintenance at this point.

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

If you choose to run an external Jenkins server and you would like to test interaction with an OpenShift master, you will need to manually configure the plugin.  See each plugin's README or the OpenShift documentation
for the specifics.

An example flow when running in an OpenShift pod:

1. Clone this git repository:
    ```
    git clone https://github.com/openshift/<plugin in question>-plugin.git
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

Each of the current OpenShift Jenkins plugins has a PR-Testing subfolder which contains instructions and scripts that
facilitate executing the same set of tests executed when we test and merge PRs.  You will also need an OpenShift Origin
development environment (see [https://github.com/openshift/origin](https://github.com/openshift/origin) for details).  Each plugin's tests are actually a subset of the OpenShift Origin extended test suite.

## Actual PR Testing

Each plugin repository under [https://github.com/openshift](https://github.com/openshift) allows for the use of the `[test]` and `[merge]` tags within a PR, via the OpenShift CI/CD infrastructure, that trigger the test and merge processing respectively.  There are a multitude of related resources which makes this happen.

### Actual PR related jobs for the plugins on [https://ci.openshift.redhat.com/jenkins/](https://ci.openshift.redhat.com/jenkins/)

Merge jobs:

* [https://ci.openshift.redhat.com/jenkins/view/All/job/merge_pull_request_jenkins_client_plugin/](https://ci.openshift.redhat.com/jenkins/view/All/job/merge_pull_request_jenkins_client_plugin/)
* [https://ci.openshift.redhat.com/jenkins/view/All/job/merge_pull_request_jenkins_openshift_login_plugin/](https://ci.openshift.redhat.com/jenkins/view/All/job/merge_pull_request_jenkins_openshift_login_plugin/)
* [https://ci.openshift.redhat.com/jenkins/view/All/job/merge_pull_request_jenkins_plugin/](https://ci.openshift.redhat.com/jenkins/view/All/job/merge_pull_request_jenkins_plugin/)
* [https://ci.openshift.redhat.com/jenkins/view/All/job/merge_pull_request_jenkins_sync_plugin/](https://ci.openshift.redhat.com/jenkins/view/All/job/merge_pull_request_jenkins_sync_plugin/)

PR test jobs:

* [https://ci.openshift.redhat.com/jenkins/view/All/job/test_pull_request_jenkins_client_plugin/](https://ci.openshift.redhat.com/jenkins/view/All/job/test_pull_request_jenkins_client_plugin/)
* [https://ci.openshift.redhat.com/jenkins/view/All/job/test_pull_request_jenkins_plugin/](https://ci.openshift.redhat.com/jenkins/view/All/job/test_pull_request_jenkins_plugin/)
* [https://ci.openshift.redhat.com/jenkins/view/All/job/test_pull_request_jenkins_openshift_login_plugin/](https://ci.openshift.redhat.com/jenkins/view/All/job/test_pull_request_jenkins_openshift_login_plugin/)
* [https://ci.openshift.redhat.com/jenkins/view/All/job/test_pull_request_jenkins_sync_plugin/](https://ci.openshift.redhat.com/jenkins/view/All/job/test_pull_request_jenkins_sync_plugin/)

Branch test jobs (common logic for PR test and merge):

* [https://ci.openshift.redhat.com/jenkins/view/All/job/test_branch_jenkins_client_plugin/](https://ci.openshift.redhat.com/jenkins/view/All/job/test_branch_jenkins_client_plugin/)
* [https://ci.openshift.redhat.com/jenkins/view/All/job/test_branch_jenkins_openshift_login_plugin/](https://ci.openshift.redhat.com/jenkins/view/All/job/test_branch_jenkins_openshift_login_plugin/)
* [https://ci.openshift.redhat.com/jenkins/view/All/job/test_branch_jenkins_plugin/](https://ci.openshift.redhat.com/jenkins/view/All/job/test_branch_jenkins_plugin/)
* [https://ci.openshift.redhat.com/jenkins/view/All/job/test_branch_jenkins_sync_plugin/](https://ci.openshift.redhat.com/jenkins/view/All/job/test_branch_jenkins_sync_plugin/)

### Artifacts in https://github.com/openshift/aos-cd-jobs

The files in `aos-cd-jobs` serve as the input for creating the jobs above in [https://ci.openshift.redhat.com/jenkins](https://ci.openshift.redhat.com/jenkins).
Read the various instructions in that repository if you ever want to update the merge and/or test jobs
for the plugins.

* [https://github.com/openshift/aos-cd-jobs/blob/master/sjb/config/test_cases/test_branch_jenkins_plugin.yml](https://github.com/openshift/aos-cd-jobs/blob/master/sjb/config/test_cases/test_branch_jenkins_plugin.yml)
* [https://github.com/openshift/aos-cd-jobs/blob/master/sjb/config/test_cases/test_branch_jenkins_client_plugin.yml](https://github.com/openshift/aos-cd-jobs/blob/master/sjb/config/test_cases/test_branch_jenkins_client_plugin.yml)
* [https://github.com/openshift/aos-cd-jobs/blob/master/sjb/config/test_cases/test_branch_jenkins_sync_plugin.yml](https://github.com/openshift/aos-cd-jobs/blob/master/sjb/config/test_cases/test_branch_jenkins_sync_plugin.yml)
* [https://github.com/openshift/aos-cd-jobs/blob/master/sjb/config/test_cases/test_branch_jenkins_openshift_login_plugin.yml](https://github.com/openshift/aos-cd-jobs/blob/master/sjb/config/test_cases/test_branch_jenkins_openshift_login_plugin.yml)
* [https://github.com/openshift/aos-cd-jobs/blob/master/sjb/config/test_cases/test_pull_request_jenkins_plugin.yml](https://github.com/openshift/aos-cd-jobs/blob/master/sjb/config/test_cases/test_pull_request_jenkins_plugin.yml)
* [https://github.com/openshift/aos-cd-jobs/blob/master/sjb/config/test_cases/test_pull_request_jenkins_client_plugin.yml](https://github.com/openshift/aos-cd-jobs/blob/master/sjb/config/test_cases/test_pull_request_jenkins_client_plugin.yml)
* [https://github.com/openshift/aos-cd-jobs/blob/master/sjb/config/test_cases/test_pull_request_jenkins_sync_plugin.yml](https://github.com/openshift/aos-cd-jobs/blob/master/sjb/config/test_cases/test_pull_request_jenkins_sync_plugin.yml)
* [https://github.com/openshift/aos-cd-jobs/blob/master/sjb/config/test_cases/test_pull_request_jenkins_openshift_login_plugin.yml](https://github.com/openshift/aos-cd-jobs/blob/master/sjb/config/test_cases/test_pull_request_jenkins_openshift_login_plugin.yml)
* [https://github.com/openshift/aos-cd-jobs/blob/master/sjb/config/test_cases/merge_pull_request_jenkins_plugin.yml](https://github.com/openshift/aos-cd-jobs/blob/master/sjb/config/test_cases/merge_pull_request_jenkins_plugin.yml)
* [https://github.com/openshift/aos-cd-jobs/blob/master/sjb/config/test_cases/merge_pull_request_jenkins_client_plugin.yml](https://github.com/openshift/aos-cd-jobs/blob/master/sjb/config/test_cases/merge_pull_request_jenkins_client_plugin.yml)
* [https://github.com/openshift/aos-cd-jobs/blob/master/sjb/config/test_cases/merge_pull_request_jenkins_sync_plugin.yml](https://github.com/openshift/aos-cd-jobs/blob/master/sjb/config/test_cases/merge_pull_request_jenkins_sync_plugin.yml)
* [https://github.com/openshift/aos-cd-jobs/blob/master/sjb/config/test_cases/merge_pull_request_jenkins_openshift_login_plugin.yml](https://github.com/openshift/aos-cd-jobs/blob/master/sjb/config/test_cases/merge_pull_request_jenkins_openshift_login_plugin.yml)

### Extended tests

In the job configuration scripts in `aos-cd-jobs` the ginkgo focus capability is leveraged such that each of the 4 plugins in turn
leverage a subset of the tests in these two extended test files:

* [https://github.com/openshift/origin/blob/master/test/extended/builds/pipeline.go](https://github.com/openshift/origin/blob/master/test/extended/builds/pipeline.go) 
* [https://github.com/openshift/origin/blob/master/test/extended/image_ecosystem/jenkins_plugin.go](https://github.com/openshift/origin/blob/master/test/extended/image_ecosystem/jenkins_plugin.go)

These test are structured such that they inspect environment variables that dictate whether:

1. The extended tests are run against the current plugin version installed in the OpenShift Jenkins CentOS7 image.
2. The extended tests are run against a newly constructed image based off of the OpenShift Jenkins CentOS7 image, where
the plugin in question has been rebuilt based on the changes in the PR branch and then copied into the new image, replacing
the official current version.


## Process for cutting a release of this plugin

Once we've merged changes into one of the OpenShift org GitHub repositories for a given plugin, we need to transfer the associated commit to the corresponding JenkinsCI org GitHub repository and follow [the upstream Jenkins project release process](https://wiki.jenkins.io/display/JENKINS/Hosting+Plugins) when we have deemed changes suitable for inclusion into the CentOS7 OpenShift Jenkins image.  We may immediately start the process of updating the RHEL7 image as well, or we may wait and let certain features soak with various upstream users prior to inclusion in RHEL7.

### Accounts and configuration (both local and upstream in various remote, Jenkins related resources) 

Some key specifics from [the upstream Jenkins project release process](https://wiki.jenkins.io/display/JENKINS/Hosting+Plugins):

* You need a login/account via [https://accounts.jenkins.io/](https://accounts.jenkins.io/) .... by extension it should also give you access to [https://issues.jenkins-ci.org](https://issues.jenkins-ci.org).  See [https://wiki.jenkins-ci.org/display/JENKINS/User+Account+on+Jenkins](https://wiki.jenkins-ci.org/display/JENKINS/User+Account+on+Jenkins).
* You should add this account to your `~/.m2/settings.xml`.  The release process noted above has details on how to do that, as well as workarounds for potential hiccups.  Read thoroughly.
* Someone on the OpenShift Developer Experience team who already has access will need to give you the necessary permissions in the files for each plugin at [https://github.com/jenkins-infra/repository-permissions-updater/tree/master/permissions](https://github.com/jenkins-infra/repository-permissions-updater/tree/master/permissions).  Existing users will need to construct a PR that adds your ID to the list of folks.
* Similarly for the [https://github.com/jenkinsci](https://github.com/jenkinsci) repositories for each plugin, we'll need to update each repositories administrator lists, etc.
with your github ID.

### Where the officially cut versions of each plugin are hosted

For our Jenkins image repository to include particular versions of our plugins in the image, the plugin versions in question
need to be available at these locations, depending on the particular plugin of course.  These are the official landing spots 
for a newly minted version of a particular plugin.

* [https://updates.jenkins.io/download/plugins/openshift-client](https://updates.jenkins.io/download/plugins/openshift-client)
* [https://updates.jenkins.io/download/plugins/openshift-pipeline](https://updates.jenkins.io/download/plugins/openshift-pipeline)
* [https://updates.jenkins.io/download/plugins/openshift-sync](https://updates.jenkins.io/download/plugins/openshift-sync)
* [https://updates.jenkins.io/download/plugins/openshift-login](https://updates.jenkins.io/download/plugins/openshift-login)

We as of yet have not had to pay attention to them, but the CI jobs over on CloudBee's Jenkins server for our 4 plugins are:

* [https://jenkins.ci.cloudbees.com/job/plugins/job/openshift-pipeline-plugin/](https://jenkins.ci.cloudbees.com/job/plugins/job/openshift-pipeline-plugin/)
* [https://jenkins.ci.cloudbees.com/job/plugins/job/openshift-client-plugin/](https://jenkins.ci.cloudbees.com/job/plugins/job/openshift-client-plugin/)
* [https://jenkins.ci.cloudbees.com/job/plugins/job/openshift-sync-plugin/](https://jenkins.ci.cloudbees.com/job/plugins/job/openshift-sync-plugin/)
* [https://jenkins.ci.cloudbees.com/job/plugins/job/openshift-login-plugin/](https://jenkins.ci.cloudbees.com/job/plugins/job/openshift-login-plugin/)

These kick in when we cut the official version at the Jenkins Update Center for a given plugin.


### Set up local repository to cut release:

To cut a new release of any of our 4 plugins, you will set up a local clone of the [https://github.com/jenkinsci](https://github.com/jenkinsci) repository, like [https://github.com/jenkinsci/openshift-pipeline-plugin](https://github.com/jenkinsci/openshift-pipeline-plugin),
and then transfer the necessary commits from the corresponding [https://github.com/openshift](https://github.com/openshift) repository, like [https://github.com/openshift/jenkins-plugin](https://github.com/openshift/jenkins-plugin).

Two approaches have evolved with the four plugins:

#### Commits between [https://github.com/openshift](https://github.com/openshift) and [https://github.com/jenkinsci](https://github.com/jenkinsci) repositories are in sync

This condition still exists for the `openshift-pipeline` plugin.  As such, the process and maintenance is a bit simpler, as we can `git rebase`.
Set up your git remotes so origin is the `https://github.com/jenkinsci/<plugin dir>` repository, and upstream is the `https://github.com/openshift/<plugin dir>/` repository.  Using `openshift-pipeline` as an example (and substitute the 
other plugin names if working with those plugins):

1. From the parent directory you've chosen for you local repository, run `git clone git@github.com:jenkinsci/openshift-pipeline-plugin.git`
1. Change directories into `openshift-pipeline-plugin`, and run `git remote add upstream git://github.com/openshift/jenkins-plugin`
1. Then pull and rebase the latest changes from [https://github.com/openshift/jenkins-plugin](https://github.com/openshift/jenkins-plugin) with the following:

```
	$ git checkout master	
	$ git fetch upstream	
	$ git fetch upstream --tags	
	$ git rebase upstream/master	
	$ git push origin master	
	$ git push origin --tags
```

#### Commits between [https://github.com/openshift](https://github.com/openshift) and [https://github.com/jenkinsci](https://github.com/jenkinsci) repositories are no longer in sync

This condition now exists for the `openshift-sync`, `openshift-login`, and `openshift-client` plugins.  As such, we have to `git cherry-pick`.  The recipe is the same as above, except the `git rebase upstream/master` is replaced with either a `git cherry-pick <commit id>` or `git cherry-pick -m 1 <merge commit id>` based on whether you are pulling in a direct commit or merge commit.  Also, the `git fetch` will be something like `git fetch git://github.com/openshift/jenkins-sync-plugin` when you are getting ready to pick commits from openshift to jenkinsci, and `git fetch git@github.com:jenkinsci/openshift-sync-plugin.git` when your are getting ready to pick commits from jenkinsci to openshift.

### Submit the new release to the Jenkins organization

After pushing the desired commits to the [https://github.com/jenkinsci](https://github.com/jenkinsci) repository for the plugin in question, you can now actuall initiate the process to create a new version of the plugin in the Jenkins Update Center.

Prerequisite: your Git ID should have push access to the [https://github.com/openshift](https://github.com/openshift) and [https://github.com/jenkinsci](https://github.com/jenkinsci) repositories for this plugin; your Jenkins ID (again see [https://wiki.jenkins-ci.org/display/JENKINS/User+Account+on+Jenkins](https://wiki.jenkins-ci.org/display/JENKINS/User+Account+on+Jenkins)) is listed in the permission file for the plugin, like [https://github.com/jenkins-infra/repository-permissions-updater/blob/master/permissions/plugin-openshift-pipeline.yml](https://github.com/jenkins-infra/repository-permissions-updater/blob/master/permissions/plugin-openshift-pipeline.yml).  Given these assumptions:

1. Then run `mvn release:prepare release:perform`
1. You'll minimally be prompted for the `release version`, `release tag`, and the `new development version`.  Default choices will be provided for each, and the defaults are typically acceptable, so you can just hit the enter key for all three prompts.  As an example, if we are currently at v1.0.36, it will provide 1.0.37 for the new `release version` and `release tag`.  For the `new development version` it will provide 1.0.38-SNAPSHOT, which is again acceptable.  The only time you *might* have to override the default provided is if we currently depend on a SNAPSHOT version of openshift-restclient-java (e.g. `5.3.0-SNAPSHOT`).  This occurs when we add new features to openshift-restclient-java, but the eclipse team has not cut a new, official release (which will typically look like `5.3.0-FINAL`).  If we are in such a mode, you will get prompted about moving off the SNAPSHOT version (the default provided would be `5.3.0`), but override this (i.e. type in `5.3.0-SNAPSHOT`).	
1. The `mvn release:prepare release:perform` command will take a few minutes to build the plugin and go through various verifications, followed by a push of the built artifacts up to Jenkins.  This typically works without further involvement but has failed for various reasons in the past.  If so, to retry with the same release version, you will need to call `git reset --hard HEAD~2` to back off the two commits created as part of publishing the release (the "release commits", where the pom.xml is updated to reflect the new version and the next snapshot version), as well as use `git tag` to delete both the local and remote version of the corresponding tag.  After deleting the commits and tags, use `git push -f` to update the commits at [https://github.com/jenkinsci/openshift-pipeline-plugin](https://github.com/jenkinsci/openshift-pipeline-plugin). Address whatever issues you have (you might have to solicit help on the Jenkins developer group: [https://groups.google.com/forum/#!forum/jenkinsci-dev](https://groups.google.com/forum/#!forum/jenkinsci-dev)) and try again.
1. If `mvn release:prepare release:perform` completes successfully, those "release commits" will look something like if you ran `git log -2`:

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

### Why the commits may be out of sync between [https://github.com/openshift](https://github.com/openshift) and [https://github.com/jenkinsci](https://github.com/jenkinsci) / Why the separate repositories / How to interpret the differences

#### Testing:

* We can employ our CI server and extended test framework for respositories in [https://github.com/openshift](https://github.com/openshift).  We cannot for [https://github.com/jenkinsci](https://github.com/jenkinsci).  For this reason we maintain the separate repositories, where run OpenShift CI against incoming changes before publishing them to the "official" [https://github.com/jenkinsci](https://github.com/jenkinsci) repositories.

#### Release cycle (when we get to the end of a release):

* We can continue development of new features for the next release in the [https://github.com/openshift](https://github.com/openshift) repositories.  But if a new bug for the current release
pops up in the interim, we can make the change in [https://github.com/jenkinsci](https://github.com/jenkinsci) first, then cherry-pick into [https://github.com/openshift](https://github.com/openshift) (employing
our regression testing), and then when ready cut the new version of the plugin in [https://github.com/jenkinsci](https://github.com/jenkinsci).

#### What version is a change really in:

* Through various scenarios and human error, those "release commits" (where the version in the pom.xml is updated) sometimes have landed in [https://github.com/openshift](https://github.com/openshift) repositories after the fact, or out of order with
how they landed in [https://github.com/jenkinsci](https://github.com/jenkinsci) repositories.  The [https://github.com/jenkinsci](https://github.com/jenkinsci) repositories are the *master* in this case.  The plugin version a change is in is expressed by the commit orders in the pom.xml release changes in the various [https://github.com/jenkinsci](https://github.com/jenkinsci) repositories.

### Get the new plugin version commits back to the [https://github.com/openshift](https://github.com/openshift) repositories

If the plugin's commits are in sync between the various [https://github.com/openshift](https://github.com/openshift) and [https://github.com/jenkinsci](https://github.com/jenkinsci) repositories:

* Run `git push https://github.com/openshift/jenkins-plugin.git master` to upload the 2 commits created for cutting the new release to our upstream, development repository, and get the two repositories back in sync.  If coordinating the two repositories via `git rebase ...`, this is necessary.  

If they are not in sync, as noted before, go the `git fetch ...` followed by `git cherry-pick ..` this time in the [https://github.com/jenkinsci](https://github.com/jenkinsci) to [https://github.com/openshift](https://github.com/openshift) direction.

While we only cut new releases of the plugins from [https://github.com/jenkinsci](https://github.com/jenkinsci), pushing the commits back to [https://github.com/openshift](https://github.com/openshift) is helpful when you are working in your local version of that repository, and are say building patches for users.  You'll have some sense of which official release the patch is built off of.

## FINALLY .... updating our OpenShift Jenkins images with the new plugin versions

As referred to previously, the new plugin version will land at `https://updates.jenkins.io/download/plugins/<plugin-name>`.  Monitor that page for the existence of the new version of the plugin.  Warning: the link for the new version can show up, but does not mean the file is available yet.  Click the link to confirm you can download the new version of the plugin.  When you can download the new version file, the new release is available.

At this point, we are back to the steps articulated in [the base plugin installation section of this repository's README](https://github.com/openshift/jenkins/blob/master/README.md#base-set-of-plugins).  You'll 
modify the text file with the new version for whatever OpenShift plugin you have cut a new version for, create a new PR, and confirm (either via `[test]` or `[merge]` in the PR) the new version of the plugin does in fact exist at the Jenkins Update center and can be successfully downloaded and installed.  Assuming so, you will go through the RHEL7 process also articulated in [the base plugin installation section of this repository's README](https://github.com/openshift/jenkins/blob/master/README.md#base-set-of-plugins), and voila!

## Opportunities for optimization/automation/enhancement

If we can get:

* an "OpenShift CI/CD account" registered with Jenkins/CloudBees 
* have that account given to the appropriate permissions in the [https://github.com/jenkinsci](https://github.com/jenkinsci) git repositories
* have that account added to the permission yml files
* have a .m2/settings.xml file constructed for that account with the account's ID/password that is baked into an image our CI/CD (aos-cd-jobs in particular) can use

Then in theory we could automate at least good portions of this process.  Probably the only thorns in the side would be merge conflicts with git commits between repositories and possible recovery from `mvn release:prepare release:perform` hiccups.

Use of the extended tests in this repository when updating versions of Jenkins or versions of the workflow/pipeline plugins could be prudent as a regression test.

Images and templates for development/testing environments ... worthwhile?  Certainly there are varying degrees of sophistication, how far one could go.