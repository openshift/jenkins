# S2I image to Jenkins slave convertor

This directory contains the convertor which transforms the S2I builder image
into a Jenkins slave image.

### Jenkins Slaves

#### Manual Setup

You can use any Docker image as a Jenkins Slave as long as it runs either JNLP
client or the swarm plugin client. For example, look at these two scripts:

* [run-jnlp-client](https://github.com/mfojtik/jenkins-ci/blob/master/jenkins-slave/contrib/openshift/run-jnlp-client)
* [run-swarm-client](https://github.com/mfojtik/jenkins-ci/blob/master/jenkins-slave/contrib/openshift/run-swarm-client)

Once you have this Docker Image, you have to manually configure Jenkins Master
to use these images as a slaves. Follow the steps in the
[jenkins-kubernetes-plugin](https://github.com/jenkinsci/kubernetes-plugin#running-in-kubernetes-google-container-engine)
documentation.

#### Tagging existing ImageStream as Jenkins Slave

If you have your Jenkins Slave image imported in OpenShift and available as an
ImageStream, you can tell Jenkins Master to automatically add it as a Kubernetes
Plugin slave. To do that, you have to set following labels:

```json
{
  "kind": "ImageStream",
  "apiVersion": "v1",
  "metadata": {
    "name": "jenkins-slave-image",
    "labels": {
      "role": "jenkins-slave"
    },
    "annotations": {
      "slave-label": "my-slave",
      "slave-directory": "/opt/app-root/jenkins"
    }
  },
  "spec": {}
}
```

The `role=jenkins-slave` label is mandatory, but the annotations are optional.
If the `slave-label` annotations is not set, Jenkins will use the ImageStream name as
label. If the `slave-directory` is not set, Jenkins will use default
*/opt/app-root/jenkins* directory.

Make sure that the Jenkins slave directory is world writeable.

#### Using this template

```console
$ oc create -f https://raw.githubusercontent.com/openshift/jenkins/master/examples/slave/s2i-slave-template.json
```

The `s2i-to-jenkins-slave` template defines a
[BuildConfig](https://docs.openshift.org/latest/dev_guide/builds.html#defining-a-buildconfig)
that uses the [Docker
Strategy](https://docs.openshift.org/latest/dev_guide/builds.html#docker-strategy-options)
to rebuild the S2I image (or any compatible image) to serve as a Jenkins Slave.
For that, we have to install JRE to run the `slave.jar`, setup nss_wrapper to
provide the username for the random UID the container will run as and a shell
script that we launch as an entrypoint from Jenkins.

When you choose `s2i-to-jenkins-slave` template, you have to specify the image
name you want to convert. The default value is `ruby-20-centos7`, but you can
change it to any available ImageStream you have in OpenShift.

Once you instantiate the template, go to *Browse/Builds* where you can see that
the build was started. You have to wait till the build finishes and the
ImageStream contains the Docker image for slaves.

The labels are annotations are automatically set for this ImageStream, so you
can ignore the section above.
