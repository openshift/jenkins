Jenkins Docker Image
====================

This repository contains Dockerfiles for a Jenkins Docker image intended for use with [OpenShift v3](https://github.com/openshift/origin)

For an example of how to use it, [see this sample.](https://github.com/openshift/origin/blob/master/examples/jenkins/README.md)

The image is pushed to DockerHub as openshift/jenkins-16-centos7.

Versions
---------------------------------
Jenkins versions currently provided are:
* jenkins-1.6x

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
    $ make build TARGET=rhel7 VERSION=1.6
    ```

*  **CentOS7 based image**

	This image is available on DockerHub. To download it run:

	```
	$ docker pull openshift/jenkins-16-centos7
	```

	To build a Jenkins image from scratch run:

	```
	$ git clone https://github.com/openshift/jenkins.git
	$ cd jenkins
	$ make build VERSION=1.6
	```

**Notice: By omitting the `VERSION` parameter, the build/test action will be performed
on all provided versions of Jenkins. Since we are currently providing only version `1.6`,
you can omit this parameter.**


Environment variables
---------------------------------

The image recognizes the following environment variables that you can set during
initialization by passing `-e VAR=VALUE` to the Docker run command.

|    Variable name          |    Description                              |
| :------------------------ | -----------------------------------------   |
|  `JENKINS_PASSWORD`       | Password for the 'admin' account            |



You can also set the following mount points by passing the `-v /host:/container` flag to Docker.

|  Volume mount point         | Description              |
| :-------------------------- | ------------------------ |
|  `/var/lib/jenkins`         | Jenkins config directory |

**Notice: When mouting a directory from the host into the container, ensure that the mounted
directory has the appropriate permissions and that the owner and group of the directory
matches the user UID or name which is running inside the container.**


Usage
---------------------------------

For this, we will assume that you are using the `openshift/jenkins-16-centos7` image.
If you want to set only the mandatory environment variables and store the database
in the `/tmp/jenkins` directory on the host filesystem, execute the following command:

```
$ docker run -d -e JENKINS_PASSWORD=<password> -v /tmp/jenkins:/var/lib/jenkins openshift/jenkins-16-centos7
```


Jenkins admin user
---------------------------------

The admin user name is set to `admin` and you have to to specify the password by
setting the `JENKINS_PASSWORD` environment variable. This process is done
upon initialization.


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
    $ make test TARGET=rhel7 VERSION=5.5
    ```

*  **CentOS based image**

    ```
    $ cd jenkins
    $ make test VERSION=1.6
    ```

**Notice: By omitting the `VERSION` parameter, the build/test action will be performed
on all provided versions of Jenkins. Since we are currently providing only version `1.6`,
you can omit this parameter.**
