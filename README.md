Jenkins Docker Image
====================

This repository contains Dockerfiles for Jenkins Docker images intended for
use with [OpenShift v3](https://github.com/openshift/origin)

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

RHEL versions currently supported are:
* RHEL7

CentOS versions currently supported are:
* CentOS7


Installation
---------------------------------
Choose either the CentOS7 or RHEL7 based image:

*  **RHEL7 based image**

    You can access these images from the Red Hat Container Catalog. See:
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

*  **CentOS7 based image**

	The images are available on DockerHub. An example download command is:

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
|  `NODEJS_SLAVE_IMAGE`  | Setting this value will override the image used for the default NodeJS agent pod configuration.  The default NodeJS agent pod uses `docker.io/openshift/jenkins-agent-nodejs-8-centos7` or `registry.redhat.io/openshift3/jenkins-agent-nodejs-8-rhel7` depending whether you are running the centos or rhel version of the Jenkins image.  This variable must be set before Jenkins starts the first time for it to have an effect.  |
|  `MAVEN_SLAVE_IMAGE`   | Setting this value overrides the image used for the default maven agent pod configuration.  The default maven agent pod uses `docker.io/openshift/jenkins-agent-maven-35-centos7` or `registry.redhat.io/openshift3/jenkins-agent-maven-35-rhel7` depending whether you are running the centos or rhel version of the Jenkins image.  This variable must be set before Jenkins starts the first time for it to have an effect.




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


Plugins
---------------------------------

### Base set of plugins

An initial set of Jenkins plugins are included in the OpenShift Jenkins images.  The general methodology
is that the CentOS7 image if first updated with any changes to the list of plugins.  After some level
of verification with that image, the RHEL7 image is updated.

#### Plugin installation for CentOS7

The top level list of plugins to install is located [here](2/contrib/openshift/base-plugins.txt).  The
format of the file is:

```
pluginId:pluginVersion
```

The file is processed by the following call in the [CentOS7 Dockerfile](2/Dockerfile):

```
/usr/local/bin/install-plugins.sh /opt/openshift/base-plugins.txt
```

Where both [base-plugins.txt](2/contrib/openshift/base-plugins.txt) and [install-plugins.sh](2/contrib/jenkins/install-plugins.sh)
are copied into the image prior to that invocation.

The running of `install-plugins.sh` will download the files listed in `base-plugins.txt`, and then open each plugin's manifest
and download any needed dependencies listed, including upgrading any previously installed dependencies as needed.

To update the version of a plugin or add a new plugin, construct a PR for this repository that updates `base-plugins.txt` appropriately.
Administrators for this repository will make sure necessary tests are run and merge the PR when things are ready.

When PRs for this repository are merged, they kick off associated builds in the [`push_jenkins_images` job on OpenShift's public
Jenkins CI/CD server](https://ci.openshift.redhat.com/jenkins/view/All/job/push_jenkins_images/).  When those builds complete,
new versions of the CentOS7 based versions of the images produced by this repository are pushed to Docker Hub.  See the top of the README for the precise list.

#### Plugin installation for RHEL7

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
the [initial set of plugins are added](#plugin-installation-for-centos7) to this image that will allow one
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
related to developing, creating new versions, and ultimately updating the images of this repository with those new versions.  If you would like to understand the specifics of that process, please visit our [plugin contribution guide](CONTRIBUTING_TO_OPENSHIFT_PLUGINS.md).

* **Kubernetes Plugin**
Though not originated out of the OpenShift organization, this plugin is invaluable in that it allows slaves to be dynamically provisioned on multiple Docker hosts using [Kubernetes](https://github.com/kubernetes/kubernetes). To learn how to use this plugin, see the [example](https://github.com/openshift/origin/tree/master/examples/jenkins/master-slave) available in the OpenShift Origin repository. For more details about this plugin, visit the [plugin](https://wiki.jenkins-ci.org/display/JENKINS/Kubernetes+Plugin) web site.

Usage
---------------------------------

For this, we will assume that you are using the `openshift/jenkins-2-centos7` image.
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

Users can choose between testing Jenkins based on a RHEL or CentOS image.

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
