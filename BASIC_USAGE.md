# Basic Uage of Jenkins Images

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
|  `SKIP_NO_PROXY_DEFAULT`       | This environment variable applies to the agent images produced by this repository.  By default, the agent images will create/update the 'no_proxy' environment variable with the hostnames for the Jenkins server endpoint and Jenkins JNLP endpoint, as communication flows to endpoints typically should *NOT* go through a HTTP Proxy.  However, if your use case dictates those flows should not be exempt from the proxy, set this environment variable to any non-empty value other than 'false'.    |
|  `ENABLE_FATAL_ERROR_LOG_FILE`       | When running this image with an OpenShift persistent volume claim for the Jenkins config directory, this environment variable allows the fatal error log file to persist if a fatal error occurs. The fatal error file will be located at `/var/lib/jenkins/logs`.   |
|  `NODEJS_AGENT_IMAGE`  | Setting this value will override the image used for the default NodeJS agent pod configuration.  For 3.x, the default NodeJS agent pod uses `docker.io/openshift/jenkins-agent-nodejs-8-centos7` or `registry.redhat.io/openshift3/jenkins-agent-nodejs-8-rhel7` depending whether you are running the centos or rhel version of the Jenkins image.  This variable must be set before Jenkins starts the first time for it to have an effect. For 4.x, the image is included in the 4.0 payload via an imagestream in the openshift namespace, and the image spec points to the internal image registry.  If you are running this image outside of OpenShift, you must either set this environment variable or manually update the setting to an accessible image spec. |
|  `MAVEN_AGENT_IMAGE`   | Setting this value overrides the image used for the default maven agent pod configuration.  For 3.x, the default maven agent pod uses `docker.io/openshift/jenkins-agent-maven-35-centos7` or `registry.redhat.io/openshift3/jenkins-agent-maven-35-rhel7` depending whether you are running the centos or rhel version of the Jenkins image.  For 4.x, the image is included in the 4.0 payload via an imagestream in the openshift namespace, and the image spec points to the internal image registry.  If you are running this image outside of OpenShift, you must either set this environment variable or manually update the setting to an accessible image spec. This variable must be set before Jenkins starts the first time for it to have an effect. |
|  `JENKINS_UC_INSECURE`       | When your Jenkins Update Center repository is using a self-signed certificate with an unknown Certificate Authority, this variable allows one to bypass the repository's SSL certificate check. The variable applies to download of the plugin which may occur during Jenkins image build, if you build an extension of the jenkins image or if you run the Jenkins image and leverage one of the options to download additional plugins (use of s2i whith plugins.txt or use of `INSTALL_PLUGINS` environment variable. |
| `MAVEN_MIRROR_URL` | Allows you to specify a [Maven mirror repository](https://maven.apache.org/guides/mini/guide-mirror-settings.html) in the form of `MAVEN_MIRROR_URL=https://nexus.mycompany.com/repository/maven-public`. The mirror repository is used by maven as an additional location for artifacts downloads. For more details on [how maven mirrors works](https://maven.apache.org/guides/mini/guide-mirror-settings.html), you can refer to maven documentation.
| `USE_JAVA_VERSION`	|	Allows you to set the Java version used by the JNLP Client for the agent images. By default the value is set to `java-17`. To use Java 8 for the value for this env var to `java-8`.	|
| `AGENT_BASE_IMAGE`	|	Setting this value overrides the image used for the 'jnlp' container in the sample kubernetes plug-in PodTemplates provided with this image.  Otherwise, the image from the 'jenkins-agent-base:latest' ImageStreamTag in the 'openshift' namespace is used.	|
| `JAVA_BUILDER_IMAGE`	|	Setting this value overrides the image used for the 'java-builder' container in the sample kubernetes plug-in PodTemplates provided with this image.  Otherwise, the image from the 'java:latest' ImageStreamTag in the 'openshift' namespace is used.	|
| `NODEJS_BUILDER_IMAGE`	|	Setting this value overrides the image used for the 'nodejs-builder' container in the sample kubernetes plug-in PodTemplates provided with this image.  Otherwise, the image from the 'nodejs:latest' ImageStreamTag in the 'openshift' namespace is used.	|
| `JAVA_FIPS_OPTIONS`  | Per this [OpenJDK support article](https://docs.redhat.com/en/documentation/red_hat_build_of_openjdk/17/html-single/configuring_red_hat_build_of_openjdk_17_on_rhel_with_fips/index#fips_settings) control how the JVM operates when running on a FIPS node. |

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
installed in the master and agent images defined in this repository.

**NOTE:** Backward compatibility is not guaranteed between differeing versions of `oc` and the
API Server.  It is recommended that you align versions of this image present in the nodes of your cluster with your OpenShift API server.
