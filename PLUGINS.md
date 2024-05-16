# Plugins

### Base set of plugins

An initial set of Jenkins plugins are included in the OpenShift Jenkins images.  The general methodology
is that the CentOS7 image if first updated with any changes to the list of plugins.  After some level
of verification with that image, the RHEL7 image is updated.


#### Plugin installation for CentOS7 V4
In v4, that call has been moved to [this script](2/contrib/jenkins/install-jenkins-core-plugins.sh), which is called from
both `Dockerfile.localdev`, `Dockerfile.rhel7` and `Dockerfile.rhel8`.

Where both [base-plugins.txt](2/contrib/openshift/base-plugins.txt) and [install-plugins.sh](2/contrib/jenkins/install-plugins.sh)
are copied into the image prior to that invocation.

The running of `install-plugins.sh` will download the files listed in `base-plugins.txt`, and then open each plugin's manifest
and download any needed dependencies listed, including upgrading any previously installed dependencies as needed.

#### Plugin installation for CentOS7 V4.10+
Starting from `release-4.10`, the `base-plugins.txt` file is instead used to generate `bundle-plugins.txt` which is the comprehensive
list of plugins used by the Jenkins image. To generate this list, developers must run `make plugins-list` prior to commit `base-plugins.txt`.
A git hook is provided to enforce that `bundle-plugins.txt` is always newer than `base-plugins.txt` on every commit attempt. And `openshift-ci`
also runs the `make plugins-list` to be sure that the locally generated list of plugins does not change between the developer commit and the
ci run.
The `bundle-plugins.txt` becomes then the source of truth for the ran, tested and verified plugins list. This file is intended to be used by
anyone who wants to build a Jenkins Image with the exact same set of plugins. Hence, this file is used by the Red Hat internal release
team (ART) and does not alter the existing release process, except that instead of getting the list of plugins from a succesful build, we now get it
from a predefined, pre-test and historized file written to the code repository.



To update the version of a plugin or add a new plugin, construct a PR for this repository that updates `base-plugins.txt` appropriately.
Administrators for this repository will make sure necessary tests are run and merge the PR when things are ready.

When PRs for this repository's `openshift-3*` branches are merged, they kick off associated builds in the [`push_jenkins_images` job on OpenShift's public
Jenkins CI/CD server](https://ci.openshift.redhat.com/jenkins/view/All/job/push_jenkins_images/).  When those builds complete,
new versions of the CentOS7 based versions of the images produced by this repository are pushed to Docker Hub.  See the top of the README for the precise list.

For v4.0, the job definitions for this repository in https://github.com/openshif/release result in our Prow based infrastructure to eventually
mirror the image content on quay.io.

#### Plugin installation for RHEL7 V3 and V4

Only OpenShift developers working for Red Hat can update the list of plugins for the RHEL7 image.  For those developers, visit this
[internal Jenkins server](https://buildvm.openshift.eng.bos.redhat.com:8443/job/devex/job/devex%252Fjenkins-plugins/) and log in (contact our CD team for permissions to this job).  Click the `Build with parameters` link, update the `PLUGIN_LIST` field, and submit the build.  The format of the data for the `PLUGIN_LIST` field is the same as `base-plugins.txt`.

The complete list of plugins (i.e. including dependencies) needs to be provided though.  The most straight forward approach is to mine the output of the CentOS7 build which passed verification for the complete list.  Just search for `Installed plugins:` and leverage copy/paste to compile what is needed.

Although this document will refrain on detailing the precise details, once the build on the internal Jenkins server is complete,
the processes will be set in motion to build the `jenkins-2-plugins` RPM that is installed by the [RHEL7 Dockerfile](2/Dockerfile.rhel7) when the next version of the RHEL7 based OpenShift Jenkins image is built.  When new versions of OpenShift are released, associated versions of the RHEL7 based versions of the images produced by this repository are pushed to the Docker registry provided to RHEL7 subscribers.

Some reference links for the OpenShift Jenkins developers and where things cross over with the CD/CL/Atomic/RHEL teams:
* http://pkgs.devel.redhat.com/cgit/rpms/?q=jenkins
* https://brewweb.engineering.redhat.com/brew/search?match=glob&type=package&terms=*jenkins*

### Adding plugins or updating existing plugins

A combination of the contents of this repository and the capabilities of OpenShift allow for a variety of ways to modify
the list of plugins either for the images directly produced from this repository, or by creating images which build
from the images directly produced from this repository.

The specifics for each approach are detailed below.

#### Installing using layering

In order to install additional Jenkins plugins, the OpenShift Jenkins image provides a way similar to how
the [initial set of plugins are added](#plugin-installation-for-centos7-v3-only) to this image that will allow one
to add or update by layering on top of this image. The derived image will provide the same functionality
as described in this documentation, in addition it will also include all plugins you list in the `plugins.txt` file.

To create a derived image in this fashion, create the following `Dockerfile`:

```
FROM openshift/jenkins-2-centos7
COPY plugins.txt /opt/openshift/configuration/plugins.txt
RUN /usr/local/bin/install-plugins.sh /opt/openshift/configuration/plugins.txt
```

The format of `plugins.txt` file is:

```
pluginId:pluginVersion
```

For example, to install the github Jenkins plugin, you specify following to `plugins.txt`:

```
github:1.11.3
```

After this, just run `docker build -t my_jenkins_image -f Dockerfile`.

#### Installing using S2I build

The [s2i](https://github.com/openshift/source-to-image) tool allows you to do additional modifications of this Jenkins image.
For example, you can use S2I to copy custom Jenkins Jobs definitions, additional
plugins or replace the default `config.xml` file with your own configuration.

To do that, you can either use the standalone `s2i` tool, that will produce the
customized Docker image or you can use OpenShift Source build strategy.

In order to include your modifications in Jenkins image, you need to have a Git
repository with following directory structure:

* `./plugins` folder that contains binary Jenkins plugins you want to copy into Jenkins
* `./plugins.txt` file that list the plugins you want to install (see the section above)
* `./configuration/jobs` folder that contains the Jenkins job definitions
* `./configuration/config.xml` file that contains your custom Jenkins configuration

Note that the `./configuration` folder will be copied into `/var/lib/jenkins`
folder, so you can also include additional files (like `credentials.xml`, etc.).

To build your customized Jenkins image, you can then execute following command:

```console
$ s2i build https://github.com/your/repository openshift/jenkins-2-centos7 your_image_name
```
NOTE:  if instead of adding a plugin you want to replace an existing plugin via dropping the binary plugin in the `./plugins` directory,
make sure the filename ends in `.jpi`.

####  Installing on Startup

The `INSTALL_PLUGINS` environment variable may be used to install a set of plugins on startup. When using a
persistent volume for `/var/lib/jenkins`, plugin installation will only happen on the initial run of the image.

In the following example, the Groovy and Pull Request Builder plugins are installed

```
INSTALL_PLUGINS=groovy:1.30,ghprb:1.35.0
```

### Plugins focused on integration with OpenShift

A subset of the plugins included by the images of this repository play a direct part in integrating between Jenkins and OpenShift.

* **OpenShift Client Plugin**
Visit [the upstream repository](https://github.com/openshift/jenkins-client-plugin) as well as the [Jenkins plugin wiki](https://wiki.jenkins-ci.org/display/JENKINS/OpenShift+Client+Plugin).  With the lessons learned from OpenShift Pipeline Plugin, as well as adjustments to the rapid evolutions of both Jenkins and OpenShift, this plugin, with its fluent styled syntax and use of the `oc` binary (exposing all the capabilities of that command), is the preferred choice for interacting with OpenShift via either Jenkins Pipeline or Freestyle jobs.

* **OpenShift Sync Plugin**
Visit [the upstream repository](https://github.com/openshift/jenkins-sync-plugin) as well as the [Jenkins plugin wiki](https://wiki.jenkins-ci.org/display/JENKINS/OpenShift+Sync+Plugin).  This plugin facilitates the integration between the OpenShift Jenkinsfile Build Strategy and Jenkins Pipelines.  It also facilitates auto-configuration of the agent pod templates for the Kubernetes Plugin.  See the [OpenShift documentation](https://docs.openshift.com) for more details.

* **OpenShift Login Plugin**
Visit [the upstream repository](https://github.com/openshift/jenkins-openshift-login-plugin) as well as the [Jenkins plugin wiki](https://wiki.jenkins-ci.org/display/JENKINS/OpenShift+Login+Plugin).  This plugin integrates the authentication and authorization of your Jenkins instance with you OpenShift cluster, providing a single sign-on look and feel.  You'll sign into the Jenkins server using the same credentials that you use to sign into the OpenShift Web Console or interact with OpenShift via the `oc` CLI.  See the [OpenShift documentation](https://docs.openshift.com) for more details.

For the above OpenShift Jenkins plugins, each of their READMEs have specifics unique to each of them on how to use and if so desired contribute to their development.  That said, there is a good deal of commonality and shared infrastructure
related to developing, creating new versions, and ultimately updating the images of this repository with those new versions.  If you would like to understand the specifics of that process, please visit our [contribution guide](CONTRIBUTING_TO_OPENSHIFT_JENKINS_IMAGE_AND_PLUGINS.md).

* **Kubernetes Plugin**
Though not originated out of the OpenShift organization, this plugin is invaluable in that it allows agents to be dynamically provisioned on multiple Docker hosts using [Kubernetes](https://github.com/kubernetes/kubernetes). To learn how to use this plugin, see the [example](https://github.com/openshift/origin/tree/master/examples/jenkins/master-agent) available in the OpenShift Origin repository. For more details about this plugin, visit the [plugin](https://wiki.jenkins-ci.org/display/JENKINS/Kubernetes+Plugin) web site.
