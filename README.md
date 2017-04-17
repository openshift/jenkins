Jenkins Docker Image
====================

This repository contains Dockerfiles for a Jenkins Docker image intended for 
use with [OpenShift v3](https://github.com/openshift/origin)

For an example of how to use it, [see this sample.](https://github.com/openshift/origin/blob/master/examples/jenkins/README.md)

The image is pushed to DockerHub as openshift/jenkins-1-centos7.

For more information about using these images with OpenShift, please see the
official [OpenShift Documentation](https://docs.openshift.org/latest/using_images/other_images/jenkins.html).

Versions
---------------------------------
Jenkins versions currently provided are:
* [jenkins-1.6x](../master/1) - DEPRECATED
* [jenkins-2.x](../master/2)

RHEL versions currently supported are:
* RHEL7

CentOS versions currently supported are:
* CentOS7


Installation
---------------------------------
Choose either the CentOS7 or RHEL7 based image:

*  **RHEL7 based image**

    To build a RHEL7 based image, you need to run Docker build on a properly
    subscribed RHEL machine.

    ```
    $ git clone https://github.com/openshift/jenkins.git
    $ cd jenkins
    $ make build TARGET=rhel7 VERSION=1
    ```

*  **CentOS7 based image**

	This image is available on DockerHub. To download it run:

	```
	$ docker pull openshift/jenkins-1-centos7
	```

	To build a Jenkins image from scratch run:

	```
	$ git clone https://github.com/openshift/jenkins.git
	$ cd jenkins
	$ make build VERSION=1
	```

**Notice: By omitting the `VERSION` parameter, the build/test action will be performed
on all provided versions of Jenkins. Since we are currently providing only version `1`,
you can omit this parameter.**


Environment variables
---------------------------------

The image recognizes the following environment variables that you can set during
initialization by passing `-e VAR=VALUE` to the Docker run command.

|    Variable name          |    Description                              |
| :------------------------ | -----------------------------------------   |
|  `JENKINS_PASSWORD`       | Password for the 'admin' account when using default Jenkin authentication.            |
| `OPENSHIFT_ENABLE_OAUTH` | Determines whether the OpenShift Login plugin manages authentication when logging into Jenkins. |
| `OPENSHIFT_PERMISSIONS_POLL_INTERVAL` | Specifies in milliseconds how often the OpenShift Login plugin polls OpenShift for the permissions associated with each user defined in Jenkins. |
| `INSTALL_PLUGINS`         | Comma-separated list of additional plugins to install on startup. The format of each plugin spec is `plugin-id:version` (as in plugins.txt) |



You can also set the following mount points by passing the `-v /host:/container` flag to Docker.

|  Volume mount point         | Description              |
| :-------------------------- | ------------------------ |
|  `/var/lib/jenkins`         | Jenkins config directory |

**Notice: When mounting a directory from the host into the container, ensure that the mounted
directory has the appropriate permissions and that the owner and group of the directory
matches the user UID or name which is running inside the container.**


Plugins
---------------------------------

#### Installing using layering

In order to install additional Jenkins plugins, the OpenShift Jenkins image provides a way
how to add those by layering on top of this image. The derived image, will provide the same functionality
as described in this documentation, in addition it will also include all plugins you list in the `plugins.txt` file.

To create derived image, you have to write following `Dockerfile`:

```
FROM openshift/jenkins-1-centos7
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
$ s2i build https://github.com/your/repository openshift/jenkins-1-centos7 your_image_name
```
NOTE:  if instead of adding a plugin you want to replace an existing plugin via dropping the binary plugin in the `./plugins` directory,
make sure the filename ends in `.jpi`.

####  Installing on Startup

The INSTALL_PLUGINS environment variable may be used to install a set of plugins on startup. When using a
persistent volume for /var/lib/jenkins, plugin installation will only happen on the initial run of the image.

In the following example, the Groovy and Pull Request Builder plugins are installed

```
INSTALL_PLUGINS=groovy:1.30,ghprb:1.35.0
```

#### Plugins of note

* **OpenShift Pipeline Plugin**
Visit [the upstream repository](https://github.com/openshift/jenkins-plugin), as well an example use of the plugin's capabilities with the [OpenShift Sample Job](https://github.com/openshift/jenkins/tree/master/1/contrib/openshift/configuration/jobs/OpenShift%20Sample) included in this image. For more details visit the Jenkins [plugin](https://wiki.jenkins-ci.org/display/JENKINS/OpenShift+Pipeline+Plugin) website.

* **OpenShift Client Plugin**
Visit [the upstream repository](https://github.com/openshift/jenkins-client-plugin) as well as the [Jenkins plugin wiki](https://wiki.jenkins-ci.org/display/JENKINS/OpenShift+Client+Plugin).  With the lessons learned from OpenShift Pipeline Plugin, as well as adjustments to the rapid evolutions of both Jenkins and OpenShift, this experimental plugin, currently included in the Centos images for this repository, is viewed as the long term replacement for OpenShift Pipeline Plugin.

* **OpenShift Sync Plugin**
Visit [the upstream repository](https://github.com/openshift/jenkins-sync-plugin) as well as the [Jenkins plugin wiki](https://wiki.jenkins-ci.org/display/JENKINS/OpenShift+Sync+Plugin).  This plugin facilitates the integration between the OpenShift Jenkinsfile Build Strategy and Jenkins Pipelines.  See the [OpenShift documentation](https://docs.openshift.com) for more details. 

* **Kubernetes Plugin**
This plugin allows slaves to be dynamically provisioned on multiple Docker hosts using [Kubernetes](https://github.com/kubernetes/kubernetes). To learn how to use this plugin, see the [example](https://github.com/openshift/origin/tree/master/examples/jenkins/master-slave) available in the OpenShift Origin repository. For more details about plugin, visit the [plugin](https://wiki.jenkins-ci.org/display/JENKINS/Kubernetes+Plugin) web site.

Usage
---------------------------------

For this, we will assume that you are using the `openshift/jenkins-1-centos7` image.
If you want to set only the mandatory environment variables and store the database
in the `/tmp/jenkins` directory on the host filesystem, execute the following command:

```
$ docker run -d -e JENKINS_PASSWORD=<password> -v /tmp/jenkins:/var/lib/jenkins openshift/jenkins-1-centos7
```


Jenkins admin user
---------------------------------

The admin user name is set to `admin`.  There are now two supported means of authenticating:

* If running outside of OpenShift, or running in OpenShift with the environment variable `OPENSHIFT_ENABLE_OAUTH` set to `false` on the container, default Jenkins authentication is used. 
You log in with the user name `admin`, supplying the password specified by the `JENKINS_PASSWORD` environment variable. If you do not override `JENKINS_PASSWORD`, the default password for `admin` is `password`.
* If running in OpenShift and the environment variable `OPENSHIFT_ENABLE_OAUTH` is set to a value other than `false` on the container, the [OpenShift Login plugin](https://github.com/openshift/jenkins-openshift-login-plugin)
manages the login process, and to login you specify valid credentials as required by the identity provider used by OpenShift.  In this case, the predefined `admin` user in the default Jenkins user database is now ignored.
Unless there is an `admin` user defined within OpenShift with sufficient permissions to the project Jenkins is running in, you will not be able to do anything with Jenkins by logging in as `admin`.  

A quick reminder on OpenShift identity providers: if, for example, the default OpenShift identity provider `Allow All` is used, you can provide any non-empty
string as the password for any valid user for the OpenShift project Jenkins is running in.  Otherwise, if `Allow All` is not used as the identity provider, then valid credentials stored with your identity provider must be provided.

For non-browser, direct HTTP or HTTPS access to Jenkins when the [OpenShift Login plugin](https://github.com/openshift/jenkins-openshift-login-plugin) manages authentication, a HTTP bearer token authentication header must be supplied
with an OpenShift token which has sufficient permissions to access the project that Jenkins is running in. A suggested token to use is a token associated with the serviceaccount for the project Jenkins in running in.  If you started
Jenkins using the example [jenkins-ephemeral](https://github.com/openshift/origin/blob/master/examples/jenkins/jenkins-ephemeral-template.json) or [jenkins-persistent](https://github.com/openshift/origin/blob/master/examples/jenkins/jenkins-persistent-template.json) templates, the commands to display the token are:

	```
	$ oc describe serviceaccount jenkins
	$ oc describe secret <serviceaccount secret name>
	``` 

Once authenticated, OpenShift roles determine which Jenkins permissions you have.  Any user with the OpenShift `admin` role for the OpenShift project Jenkins is running in will have the same permissions as those assigned to an administrative user within Jenkins.
Users with the `edit` or `view` roles for the OpenShift project Jenkins is running in will have progressively reduced permissions within Jenkins.

For the `view` role, the Jenkins permissions are:

* hudson.model.Hudson.READ
* hudson.model.Item.READ
* com.cloudbees.plugins.credentials.CredentialsProvider.VIEW

For the `edit` role, in addition to the permissions available to `view`:

* hudson.model.Item.BUILD
* hudson.model.Item.CONFIGURE
* hudson.model.Item.CREATE
* hudson.model.Item.DELETE
* hudson.model.Item.WORKSPACE
* hudson.scm.SCM.TAG
* jenkins.model.Jenkins.RUN_SCRIPTS

Users authenticated against OpenShift OAuth will be added to the Jenkins authorization matrix upon their first successful login.

Permissions for users in Jenkins can be changed in OpenShift after those users are initially established in Jenkins.  The OpenShift Login plugin polls the OpenShift API server for permissions and will update the permissions stored in
Jenkins for each Jenkins user with the permissions retrieved from OpenShift.  Technically speaking, you can change the permissions for a Jenkins user from the Jenkins UI as well, but those changes will be overwritten the next
time the poll occurs.

You can control how often the polling occurs with the `OPENSHIFT_PERMISSIONS_POLL_INTERVAL` environment variable.  The default polling interval when no environment variable is set is 5 minutes.



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
on all provided versions of Jenkins. Since we are currently providing only version `1`,
you can omit this parameter.**
