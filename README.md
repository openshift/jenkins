# OpenShift Jenkins Images

## Introduction
This repository contains Dockerfiles for building Jenkins Master and Agent images intended for use with [OKD 4](https://www.okd.io/) and [Red Hat OpenShift 4](https://www.redhat.com/en/technologies/cloud-computing/openshift).

## Hosted Images
All OpenShift 4 images (including the ones from this repository) are based
off of the [Red Hat Universal Base Image 8](https://catalog.redhat.com/software/containers/ubi8/5c647760bed8bd28d0e38f9f).

**NOTE:** Only the 64-bit JVM is available in all images.
### Community
These images are available via [quay.io](https://quay.io) and are community supported.
* [quay.io/openshift/origin-jenkins](https://quay.io/openshift/origin-jenkins)
* [quay.io/openshift/origin-jenkins-agent-nodejs](https://quay.io/openshift/origin-jenkins-agent-nodejs)
* [quay.io/openshift/origin-jenkins-agent-maven](https://quay.io/openshift/origin-jenkins-agent-maven)
* [quay.io/openshift/origin-jenkins-agent-base](https://quay.io/openshift/origin-jenkins-agent-base)

**NOTE:** The jenkins-agent-maven and jenkins-agent-nodejs image are no longer maintained as of version 4.11 and no longer published as of version 4.17.

### Red Hat OpenShift
These images are available via the [Red Hat Catalog](https://catalog.redhat.com) for customers with subscriptions.
#### 4.10 and lower
These images are intended for OpenShift 4.10 and lower.
* [openshift4/ose-jenkins](https://catalog.redhat.com/software/containers/openshift4/ose-jenkins/5cdd918ad70cc57c44b2d279)
* [openshift4/ose-jenkins-agent-base](https://catalog.redhat.com/software/containers/openshift4/ose-jenkins-agent-base/5cdd8e2fbed8bd5717d66e77)
* [openshift4/ose-jenkins-agent-maven](https://catalog.redhat.com/software/containers/openshift4/ose-jenkins-agent-maven/5cdd8fe55a13467289f615e7)
* [openshift4/ose-jenkins-agent-nodejs-12-rhel8](https://catalog.redhat.com/software/containers/openshift4/ose-jenkins-agent-nodejs-12-rhel8/5f6c39da1fa29796579cdff7)

#### 4.11 to 4.15
These images are intended for OpenShift 4.11 and 4.15.
* [ocp-tools-4/jenkins-rhel8](https://catalog.redhat.com/software/containers/ocp-tools-4/jenkins-rhel8/5fe1f38288e9c2f788526306)
* [ocp-tools-4/jenkins-agent-base-rhel8](https://catalog.redhat.com/software/containers/ocp-tools-4/jenkins-agent-base-rhel8/6241e3457847116cf8577aea)

#### 4.16 and higher
These images are intended for OpenShift 4.16 and higher.
* [ocp-tools-4/jenkins-rhel9](https://catalog.redhat.com/software/containers/ocp-tools-4/jenkins-rhel9/65dc9063b7db2e8b83a5b299)
* [ocp-tools-4/jenkins-agent-base-rhel9](https://catalog.redhat.com/software/containers/ocp-tools-4/jenkins-agent-base-rhel9/65dc9063b7db2e8b83a5b29e)

**NOTE:** The jenkins-agent-maven and jenkins-agent-nodejs image are no longer maintained or published as of version 4.11.

## Building
Please see [BUILDING.md](https://github.com/openshift/jenkins/blob/master/BUILDING.md).

## Basic Usage
Please see [BASIC_USAGE.md](https://github.com/openshift/jenkins/blob/master/BASIC_USAGE.md).

## Advanced Usage
Please see [ADVANCED_USAGE.md](https://github.com/openshift/jenkins/blob/master/ADVANCED_USAGE.md).

## Plugins
Please see [PLUGINS.md](https://github.com/openshift/jenkins/blob/master/PLUGINS.md).

## Security
Please see [SECURITY.md](https://github.com/openshift/jenkins/blob/master/SECURITY.md).

## Testing
Please see [TESTING.md](https://github.com/openshift/jenkins/blob/master/TESTING.md).

## Contributing
Please see [CONTRIBUTING.md](https://github.com/openshift/jenkins/blob/master/CONTRIBUTING.md).
