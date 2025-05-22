Configuration files
-------------------------------

The layering and s2i build flows noted above for updating the list of plugins can also be used to update the configuration injected into the Jenkins deployment.  However, don't forget the note about copying of config data and Persistent Volumes in the [startup notes](#startup-notes-for-the-jenkins-core-image).

A typical scenario employed by our users has been extending the Jenkins image to add groovy init scripts to customize your Jenkins installation.

A quick recipe of how to do that via layering would be:

* mkdir -p contrib/openshift/configuration/init.groovy.d
* create a contrib/openshift/configuration/init.groovy.d/foo.groovy file with whatever groovy init steps you desire
* create a Dockerfile with (adjusting the image ref as you see fit)

```
FROM registry.redhat.io/ocp-tools-4/jenkins-rhel9:v4.17.0
COPY ./contrib/openshift /opt/openshift
```

And then update your Jenkins deployment to use the resulting image directly, or update the ImageStreamTag reference you Jenkins deployment is employing, with our new image.  During startup,
the existing run script your new image inherits from this repositories Jenkins image will copy the groovy init script to the appropriate spot under the Jenkins home directory.


Usage
---------------------------------

For this, we will assume that you are using an `registry.redhat.io/ocp-tools-4/jenkins-rhel9` for v4.x.

If you want to set only the mandatory environment variables and store the database
in the `/tmp/jenkins` directory on the host filesystem, execute the following command:

```
$ docker run -d -e JENKINS_PASSWORD=<password> -v /tmp/jenkins:/var/lib/jenkins ocp-tools-4/jenkins-rhel9:v4.17.0
```


Jenkins admin user
---------------------------------

Authenticating into a Jenkins server running within the OpenShift Jenkins image is controlled by the [OpenShift Login plugin](https://github.com/openshift/jenkins-openshift-login-plugin), taking into account:

* Whether or not the container is running in an OpenShift Pod
* How the [environment variables](https://github.com/openshift/jenkins#environment-variables) recognized by the image are set

See the [OpenShift Login plugin documentation](https://github.com/openshift/jenkins-openshift-login-plugin) for details on how it manages authentication.

However, when the default authentication mechanism for Jenkins is used, if you are using the OpenShift Jenkins image, you log in with the user name `admin`, supplying the password specified by the `JENKINS_PASSWORD` environment variable set on the container. If you do not override `JENKINS_PASSWORD`, the default password for `admin` is `password`.
