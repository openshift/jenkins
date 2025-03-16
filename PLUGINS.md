# Plugins

### Base set of plugins

An initial set of Jenkins plugins are included in the OpenShift Jenkins images.  The general methodology
is that the UBI image if first updated with any changes to the list of plugins.  After some level
of verification with that image, the RHEL9 image is updated.

#### Plugin installation for V4
In v4, that call has been moved to [this script](2/contrib/jenkins/install-jenkins-core-plugins.sh), which is called from
`Dockerfile.localdev`, `Dockerfile.rhel8` and `Dockerfile.rhel9`.

Where both [base-plugins.txt](2/contrib/openshift/base-plugins.txt) and [install-plugins.sh](2/contrib/jenkins/install-plugins.sh)
are copied into the image prior to that invocation.

The running of `install-plugins.sh` will download the files listed in `base-plugins.txt`, and then open each plugin's manifest
and download any needed dependencies listed, including upgrading any previously installed dependencies as needed.

#### Plugin installation for V4.11+

Starting from `release-4.11`, manually update `base-plugins.txt` and `bundle-plugins.txt` to be the same.

### Adding plugins or updating existing plugins

A combination of the contents of this repository and the capabilities of OpenShift allow for a variety of ways to modify
the list of plugins either for the images directly produced from this repository, or by creating images which build
from the images directly produced from this repository.

The specifics for each approach are detailed below.

#### Installing using layering

In order to install additional Jenkins plugins, the OpenShift Jenkins image provides a way that will allow one
to add or update by layering on top of this image. The derived image will provide the same functionality
as described in this documentation, in addition it will also include all plugins you list in the `plugins.txt` file.

To create a derived image in this fashion, create the following `Dockerfile`:

```
FROM registry.redhat.io/ocp-tools-4/jenkins-rhel9
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
$ s2i build https://github.com/your/repository registry.redhat.io/ocp-tools-4/jenkins-rhel9 your_image_name
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
