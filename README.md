Jenkins Docker Image
====================

This repository contains Dockerfiles for Jenkins Docker images intended for
use with [OpenShift v3 and OpenShift v4](https://github.com/openshift/origin)

For an example of how to use it, [see this sample.](https://github.com/openshift/origin/blob/master/examples/jenkins/README.md)

The images are pushed to DockerHub as openshift/jenkins-2-centos7, openshift/jenkins-slave-base-centos7, openshift/jenkins-agent-maven-35-centos7, and openshift/jenkins-agent-nodejs-8-centos7.

The slave-maven and slave-nodejs for both centos7 and rhel7  are being deprecated as part of v3.10 of OpenShift.
Additionally, development of these images will cease as of v3.10.  And they are removed from this repository as
part of the v4.0 development cycle.

Support for the [OpenShift Pipeline Plugin](https://github.com/openshift/jenkins-plugin) stopped with v3.11 of OpenShift.
The plugin itself and any samples around it are removed as part of v4.0.

For more information about using these images with OpenShift, please see the
official [OpenShift Documentation](https://docs.okd.io/latest/using_images/other_images/jenkins.html).

Versions
---------------------------------
Jenkins versions currently provided are:
* [jenkins-2.x](../master/2)

For OpenShift v3, the options for the images` operating system are as follows:

RHEL versions currently supported are:
* RHEL7

CentOS versions currently supported are:
* CentOS7

For OpenShift v4, the operating system choice is reduced.  All OpenShift v4 images (including the ones from this repository) are based
off of the ["Universal Based Image" or "UBI"](https://access.redhat.com/documentation/en-us/red_hat_enterprise_linux_atomic_host/7/html/getting_started_with_containers/using_red_hat_base_container_images_standard_and_minimal#using_rhel_atomic_base_images_minimal).


Installation (OpenShift V3)
---------------------------------
Choose either the CentOS7 or RHEL7 based image:

*  **RHEL7 based image**

You can access these images from the Red Hat Container Catalog. For OpenShift v3 see:
* https://access.redhat.com/containers/#/registry.access.redhat.com/openshift3/jenkins-2-rhel7
* https://access.redhat.com/containers/#/registry.access.redhat.com/openshift3/jenkins-slave-base-rhel7
* https://access.redhat.com/containers/#/registry.access.redhat.com/openshift3/jenkins-agent-maven-35-rhel7
* https://access.redhat.com/containers/#/registry.access.redhat.com/openshift3/jenkins-agent-nodejs-8-rhel7

To build a RHEL7 based image, you need to run Docker build on a properly
subscribed RHEL machine.

    ```
    $ git clone https://github.com/openshift/jenkins.git
    $ cd jenkins
    $ make build TARGET=rhel7 VERSION=2
    ```

Also note, as of 3.11, the RHEL images are hosted at registry.redhat.io as well.  This is the terms based
registry and requires credentials for access.  See [Transitioning the Red Hat container registry](https://www.redhat.com/en/blog/transitioning-red-hat-container-registry) for details:
* registry.redhat.io/openshift3/jenkins-2-rhel7:v3.11
* registry.redhat.io/openshift3/jenkins-agent-nodejs-8-rhel7:v3.11
* registry.redhat.io/openshift3/jenkins-agent-maven-35-rhel7:v3.11
* registry.redhat.io/openshift3/jenkins-slave-base-rhel7:v3.11

The openshift cluster install for 3.11 will ensure that credentials are provided and subsequently available on the nodes
in the cluster to facilitate image pulling.


*  **CentOS7 based image**

The v3.x images are available on DockerHub. An example download command is:

	```
	$ docker pull openshift/jenkins-2-centos7
	```

To build a Jenkins image from scratch run:

	```
	$ git clone https://github.com/openshift/jenkins.git
	$ cd jenkins
	$ make build VERSION=2
	```

**Notice: By omitting the `VERSION` parameter, the build/test action will be performed
on all provided versions of Jenkins.**

If you are curious about the precise level of Jenkins for either `jenkins-2-centos7` or `jenkins-2-rhel7`, then
you can execute:


    $ docker run -it <image spec> /etc/alternatives/java -jar /usr/lib/jenkins/jenkins.war --version


For example:


    $ docker run -it docker.io/openshift/jenkins-2-centos7:latest /etc/alternatives/java -jar /usr/lib/jenkins/jenkins.war --version

Installation (OpenShift V4)
---------------------------------

Starting with v4.0, the images are only available on quay.io for public community support. Their pull specs are:
* quay.io/openshift/origin-jenkins:<release tag>
* quay.io/openshift/origin-jenkins-agent-nodejs:<release tag>
* quay.io/openshift/origin-jenkins-agent-maven:<release tag>
* quay.io/openshift/origin-jenkins-agent-base:<release tag>

Visit quay.io to discover the set of tags for each image.  For example, for the core jenkins image, the tags are [here](https://quay.io/repository/openshift/origin-jenkins?tab=tags)

The images are also still available at the Red Hat Container Catalog for customers with subscriptions,
though with some changes in the naming.

As with the initial introduction in 3.11, given the [transitioning of the Red Hat container registry](https://www.redhat.com/en/blog/transitioning-red-hat-container-registry), the RHEL based images are available at both registry.access.redhat.com and registry.redhat.io.
The terms based registry, registry.redhat.io, which requires credentials for access, is the strategic direction, and
will be the only location for RHEL8 based content when that is available.  The pull secret you obtain from try.openshift.com includes
access to registry.redhat.io.  The image pull specs are:
* registry.redhat.io/openshift4/ose-jenkins:<release tag>
* registry.redhat.io/openshift4/ose-jenkins-agent-nodejs:<release tag>
* registry.redhat.io/openshift4/ose-jenkins-agent-maven:<release tag>
* registry.redhat.io/openshift4/ose-jenkins-agent-base:<release tag>

OpenShift v4 also removes the 32 bit JVM option.  Only 64 bit will be provided for all images.

The `Dockerfile.rhel7` variants still exists, but as part of the `CentOS` vs. `RHEL` distinction no longer existing, the various `Dockerfile` files have been renamed to `Dockerfile.localdev` to more clearly denote that they are for builds on developers' local machines that most likely do not have a Red Hat subscription / entitlement.  The `Dockerfile.localdev` variants are structured to allow building of the images on machines without `RHEL` subscriptions, even though the base images are no longer based on `CentOS`.  Subscriptions are still required for use of `Dockerfile.rhel7`.

With any local builds, if for example you plan on submitting a PR to this repository, you still build the same way as with OpenShift v3 with respect to the `make` invocations.  

Be aware, no support in any way is provided for running images created from any of the `Dockerfile.localdev` files.  And in fact the images hosted on both quay.io and the Red Hat Container Catalog are based off the `Dockerfile.rhel7` files.

And lastly, as part of 4.x cluster installs, the OpenShift Jenkins image version corresponding to the cluster version is part of the image payload for the install.  So the `jenkins` ImageStream in the `openshift` namespace will have image references that point to the image registry associated with your install instead of these public registries noted above.  There is also an ImageStream for each of the agent images in the `openshift` namespace in 4.x installs.


Startup notes for the Jenkins core image
---------------------------------

When you run you startup the main Jenkins image in an OpenShift pod for the first time, it performs various set up actions, including:
* Setting the JVM parameters for the actual start of the Jenkins JVM
* Updating the /etc/passwd so that the random, non-root user ID employed works
* Copies all the default configuration from the image to the appropriate locations under the Jenkins home directory (which maps to the image's volume mount)
* Copies all the plugins to the appropriate locations under the Jenkins home directory

By default, all copies to the Jenkins home directory are only done on the initial startup if a Persistent Volume is employed for the Jenkins deployment.  There are ways to override that behavior by environment variables (see the next section below).  But you can also recycle the PVCs during restarts of your Jenkins deployment if you update the image being used and want to reset the system that way.


Environment variables
---------------------------------

The image recognizes the following environment variables that you can set during
initialization by passing `-e VAR=VALUE` to the Docker run command.

|    Variable name          |    Description                              |
| :------------------------ | -----------------------------------------   |
|  `JENKINS_PASSWORD`       | Password for the 'admin' account when using default Jenkins authentication.            |
| `OPENSHIFT_ENABLE_OAUTH`  | Determines whether the OpenShift Login plugin manages authentication when logging into Jenkins. |
| `OPENSHIFT_PERMISSIONS_POLL_INTERVAL` | Specifies in milliseconds how often the OpenShift Login plugin polls OpenShift for the permissions associated with each user defined in Jenkins. |
| `INSTALL_PLUGINS`         | Comma-separated list of additional plugins to install on startup. The format of each plugin spec is `plugin-id:version` (as in plugins.txt) |
|  `OVERRIDE_PV_CONFIG_WITH_IMAGE_CONFIG`       | When running this image with an OpenShift persistent volume for the Jenkins config directory, the transfer of configuration from the image to the persistent volume is only done the first startup of the image as the persistent volume is assigned by the persistent volume claim creation. If you create a custom image that extends this image and updates configuration in the custom image after the initial startup, by default it will not be copied over, unless you set this environment variable to some non-empty value other than 'false'. |
|  `OVERRIDE_PV_PLUGINS_WITH_IMAGE_PLUGINS`       | When running this image with an OpenShift persistent volume for the Jenkins config directory, the transfer of plugins from the image to the persistent volume is only done the first startup of the image as the persistent volume is assigned by the persistent volume claim creation. If you create a custom image that extends this image and updates plugins in the custom image after the initial startup, by default they will not be copied over, unless you set this environment variable to some non-empty value other than 'false'. |
|  `OVERRIDE_RELEASE_MIGRATION_OVERWRITE`       | When running this image with an OpenShift persistent volume for the Jenkins config directory, and this image is starting in an existing deployment created with an earlier version of this image, unless the environment variable is set to some non-empty value other than 'false', the plugins from the image will replace any versions of those plugins currently residing in the Jenkins plugin directory.  |
|  `SKIP_NO_PROXY_DEFAULT`       | This environment variable applies to the agent/slave images produced by this repository.  By default, the agent/slave images will create/update the 'no_proxy' environment variable with the hostnames for the Jenkins server endpoint and Jenkins JNLP endpoint, as communication flows to endpoints typically should *NOT* go through a HTTP Proxy.  However, if your use case dictates those flows should not be exempt from the proxy, set this environment variable to any non-empty value other than 'false'.    |
|  `ENABLE_FATAL_ERROR_LOG_FILE`       | When running this image with an OpenShift persistent volume claim for the Jenkins config directory, this environment variable allows the fatal error log file to persist if a fatal error occurs. The fatal error file will be located at `/var/lib/jenkins/logs`.   |
|  `NODEJS_SLAVE_IMAGE`  | Setting this value will override the image used for the default NodeJS agent pod configuration.  For 3.x, the default NodeJS agent pod uses `docker.io/openshift/jenkins-agent-nodejs-8-centos7` or `registry.redhat.io/openshift3/jenkins-agent-nodejs-8-rhel7` depending whether you are running the centos or rhel version of the Jenkins image.  This variable must be set before Jenkins starts the first time for it to have an effect. For 4.x, the image is included in the 4.0 payload via an imagestream in the openshift namespace, and the image spec points to the internal image registry.  If you are running this image outside of OpenShift, you must either set this environment variable or manually update the setting to an accessible image spec. |
|  `MAVEN_SLAVE_IMAGE`   | Setting this value overrides the image used for the default maven agent pod configuration.  For 3.x, the default maven agent pod uses `docker.io/openshift/jenkins-agent-maven-35-centos7` or `registry.redhat.io/openshift3/jenkins-agent-maven-35-rhel7` depending whether you are running the centos or rhel version of the Jenkins image.  For 4.x, the image is included in the 4.0 payload via an imagestream in the openshift namespace, and the image spec points to the internal image registry.  If you are running this image outside of OpenShift, you must either set this environment variable or manually update the setting to an accessible image spec. This variable must be set before Jenkins starts the first time for it to have an effect. |
|  `JENKINS_UC_INSECURE`       | When your Jenkins Update Center repository is using a self-signed certificate with an unknown Certificate Authority, this variable allows one to bypass the repository's SSL certificate check. The variable applies to download of the plugin which may occur during Jenkins image build, if you build an extension of the jenkins image or if you run the Jenkins image and leverage one of the options to download additional plugins (use of s2i whith plugins.txt or use of `INSTALL_PLUGINS` environment variable. |
| `MAVEN_MIRROR_URL` | Allows you to specify a [Maven mirror repository](https://maven.apache.org/guides/mini/guide-mirror-settings.html) in the form of `MAVEN_MIRROR_URL=https://nexus.mycompany.com/repository/maven-public`. The mirror repository is used by maven as an additional location for artifacts downloads. For more details on [how maven mirrors works](https://maven.apache.org/guides/mini/guide-mirror-settings.html), you can refer to maven documentation.

You can also set the following mount points by passing the `-v /host:/container` flag to Docker.

|  Volume mount point         | Description              |
| :-------------------------- | ------------------------ |
|  `/var/lib/jenkins`         | Jenkins config directory |

**Notice: When mounting a directory from the host into the container, ensure that the mounted
directory has the appropriate permissions and that the owner and group of the directory
matches the user UID or name which is running inside the container.**

Inclusion of the `oc` binary
---------------------------------

To assist in interacting with the OpenShift API server while using this image, the `oc` binary, the CLI command for OpenShift, has been
installed in the master and slave images defined in this repository.


However, it needs to be noted that backward compatibility is not guaranteed between different versions of `oc` and the OpenShift
API Server.  As such, it is recommended that you align versions of this image present in the nodes of your cluster with your
OpenShift API server.  In other words, you should use the version specific tag instead of the `latest` tag.

|  Jenkins image version      | `oc` client version      |
| :-------------------------- | ------------------------ |
|  `jenkins-*-centos7:v3.7`   | 3.7 `oc` binary          |
|  `jenkins-*-centos7:v3.6`   | 3.6 `oc` binary          |
|  `jenkins-*-centos7:latest` | `oc` binary from `docker.io/openshift/origin:latest` image          |
|  `jenkins-*-rhel7:v3.7`     | 3.7 `oc` binary          |
|  `jenkins-*-rhel7:v3.6`     | 3.6 `oc` binary          |
|  `jenkins-*-rhel7:latest`   | 3.6 `oc` binary \*\*     |


**Notice: the `latest` tag for the RHEL7 images will point to 3.6 indefinitely in order to support users on older clusters with older slave
configurations that point to the "latest" tag.  This way, they will have an older `oc` client which should be able to communicate with both 3.6
and newer versions of OpenShift API Servers.  As the support policy is less stringent for the CentOS7 image, the `latest` tag there will
make the more obvious correlation to the latest built version of OpenShift (which can include pre-GA versions).

**Notice:  There is an additional consideration with the pod configurations for the Kubernetes Plugin; earlier versions of this image
did not specify the "pull always" policy for the default agents/slaves configured.  As a result, users may have older/different images on
your nodes depending when the images were pulled.  Starting with the 3.7 release, the default changed to "pull always" to avoid this problem
in the future.  But if you started using this image prior to 3.7, verification of your Kubernetes plugin configurations for the image pull
policy used is warranted to guarantee consistency around what image is being used on each of your nodes.

The `oc` binary is still included in the v4 images as well.  And the same recommendations around client/server version synchronization still apply.

Jenkins security advisories, the "master" image from this repository, and the `oc` binary
---------------------------------

Any security advisory related updates to Jenkins core or the plugins we include in the OpenShift Jenkins master image will only occur in the v3.11 and v4.x
branches of this repository.

We do support running the v3.11 version of the master image against older v3.x (as far back as v3.4) OpenShift clusters if you want to pick up Jenkins security advisory
updates.  Per the prior section, we advise that you import a version of `oc` into your Jenkins installation that matches your OpenShift
cluster via the "Global Tool Configuration" option in Jenkins either via the UI, CLI, or groovy init scripts.

Our OpenShift Client Plugin has some documentation on doing this [here](https://github.com/openshift/jenkins-client-plugin#setting-up-jenkins-nodes).

Also note for the RHEL image, the v3.11 image examines whether it is running in an OpenShift Pod and what version the cluster is at.  If the cluster is at a version prior to v3.11, the Maven and NodeJS agent example configuration for the kubernetes plugin will point to registry.access.redhat.com for
the image setting.  If the cluster is at v3.11, the image setting will point to the terms based registry at registry.access.io.


Plugins
---------------------------------

### Base set of plugins

An initial set of Jenkins plugins are included in the OpenShift Jenkins images.  The general methodology
is that the CentOS7 image if first updated with any changes to the list of plugins.  After some level
of verification with that image, the RHEL7 image is updated.

#### Plugin installation for CentOS7 V3 Only

The top level list of plugins to install is located [here](2/contrib/openshift/base-plugins.txt).  The
format of the file is:

```
pluginId:pluginVersion
```

For v3, the file is processed by the following call in the [CentOS7 Dockerfile](2/Dockerfile):

```
/usr/local/bin/install-plugins.sh /opt/openshift/base-plugins.txt
```

In v4. that call has been moved to [this script](2/contrib/jenkins/install-jenkins-core-plugins.sh), which is called from
both `Dockerfile.localdev` and `Dockerfile.rhel7`.

Where both [base-plugins.txt](2/contrib/openshift/base-plugins.txt) and [install-plugins.sh](2/contrib/jenkins/install-plugins.sh)
are copied into the image prior to that invocation.

The running of `install-plugins.sh` will download the files listed in `base-plugins.txt`, and then open each plugin's manifest
and download any needed dependencies listed, including upgrading any previously installed dependencies as needed.

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
Visit [the upstream repository](https://github.com/openshift/jenkins-sync-plugin) as well as the [Jenkins plugin wiki](https://wiki.jenkins-ci.org/display/JENKINS/OpenShift+Sync+Plugin).  This plugin facilitates the integration between the OpenShift Jenkinsfile Build Strategy and Jenkins Pipelines.  It also facilitates auto-configuration of the slave pod templates for the Kubernetes Plugin.  See the [OpenShift documentation](https://docs.openshift.com) for more details.

* **OpenShift Login Plugin**
Visit [the upstream repository](https://github.com/openshift/jenkins-openshift-login-plugin) as well as the [Jenkins plugin wiki](https://wiki.jenkins-ci.org/display/JENKINS/OpenShift+Login+Plugin).  This plugin integrates the authentication and authorization of your Jenkins instance with you OpenShift cluster, providing a single sign-on look and feel.  You'll sign into the Jenkins server using the same credentials that you use to sign into the OpenShift Web Console or interact with OpenShift via the `oc` CLI.  See the [OpenShift documentation](https://docs.openshift.com) for more details.

For the above OpenShift Jenkins plugins, each of their READMEs have specifics unique to each of them on how to use and if so desired contribute to their development.  That said, there is a good deal of commonality and shared infrastructure
related to developing, creating new versions, and ultimately updating the images of this repository with those new versions.  If you would like to understand the specifics of that process, please visit our [contribution guide](CONTRIBUTING_TO_OPENSHIFT_JENKINS_IMAGE_AND_PLUGINS.md).

* **Kubernetes Plugin**
Though not originated out of the OpenShift organization, this plugin is invaluable in that it allows slaves to be dynamically provisioned on multiple Docker hosts using [Kubernetes](https://github.com/kubernetes/kubernetes). To learn how to use this plugin, see the [example](https://github.com/openshift/origin/tree/master/examples/jenkins/master-slave) available in the OpenShift Origin repository. For more details about this plugin, visit the [plugin](https://wiki.jenkins-ci.org/display/JENKINS/Kubernetes+Plugin) web site.

Configuration files
-------------------------------

The layering and s2i build flows noted above for updating the list of plugins can also be used to update the configuration injected into the Jenkins deployment.  However, don't forget the note about copying of config data and Persistent Volumes in the [startup notes](#startup-notes-for-the-jenkins-core-image).

A typical scenario employed by our users has been extending the Jenkins image to add groovy init scripts to customize your Jenkins installation.

A quick recipe of how to do that via layering would be:

* mkdir -p contrib/openshift/configuration/init.groovy.d
* create a contrib/openshift/configuration/init.groovy.d/foo.groovy file with whatever groovy init steps you desire
* create a Dockerfile with (adjusting the image ref as you see fit)

```
FROM registry.access.redhat.com/openshift3/jenkins-2-rhel7:v3.11
COPY ./contrib/openshift /opt/openshift
```

And then update your Jenkins deployment to use the resulting image directly, or update the ImageStreamTag reference you Jenkins deployment is employing, with our new image.  During startup,
the existing run script your new image inherits from this repositories Jenkins image will copy the groovy init script to the appropriate spot under the Jenkins home directory.


Usage
---------------------------------

For this, we will assume that you are using an `openshift/jenkins-2-centos7` image for v3.x, or
`quay.io/openshift/origin-jenkins` for v4.x.

If you want to set only the mandatory environment variables and store the database
in the `/tmp/jenkins` directory on the host filesystem, execute the following command:

```
$ docker run -d -e JENKINS_PASSWORD=<password> -v /tmp/jenkins:/var/lib/jenkins openshift/jenkins-2-centos7
```


Jenkins admin user
---------------------------------

Authenticating into a Jenkins server running within the OpenShift Jenkins image is controlled by the [OpenShift Login plugin](https://github.com/openshift/jenkins-openshift-login-plugin), taking into account:

* Whether or not the container is running in an OpenShift Pod
* How the [environment variables](https://github.com/openshift/jenkins#environment-variables) recognized by the image are set

See the [OpenShift Login plugin documentation](https://github.com/openshift/jenkins-openshift-login-plugin) for details on how it manages authentication.

However, when the default authentication mechanism for Jenkins is used, if you are using the OpenShift Jenkins image, you log in with the user name `admin`, supplying the password specified by the `JENKINS_PASSWORD` environment variable set on the container. If you do not override `JENKINS_PASSWORD`, the default password for `admin` is `password`.


Test
---------------------------------

This repository also provides a test framework which checks basic functionality
of the Jenkins image.

With v3, users can choose between testing Jenkins based on a RHEL (where you are running on a platform with subscriptions) or CentOS image.
With v4, there is not CentOS vs. RHEL distinction, but we still use TARGET to control whether subscriptions are required when building the test image,
and we reuse the v3 values (i.e. `rhel7`) for that purpose.

*  **RHEL based image**

    To test a RHEL7 based Jenkins image, you need to run the test on a properly
    subscribed RHEL machine.

    ```
    $ cd jenkins
    $ make test TARGET=rhel7 VERSION=2
    ```

*  **CentOS based image**

    ```
    $ cd jenkins
    $ make test VERSION=2
    ```

**Notice: By omitting the `VERSION` parameter, the build/test action will be performed
on all provided versions of Jenkins. Since we are currently providing only version `2`,
you can omit this parameter.**

## PR testing for this repository

As with the plugins focused on OpenShift integration, see [the contribution guide](CONTRIBUTING_TO_OPENSHIFT_JENKINS_IMAGE_AND_PLUGINS.md).
