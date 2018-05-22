Node Slave Image
====================

This repository contains Dockerfiles for a Jenkins Agent Docker image intended for 
use with [OpenShift v3](https://github.com/openshift/origin)

Node Registry Support
---------------------------------
This Node agent image supports using a [Node Mirror/Registry](https://blog.sonatype.com/using-nexus-3-as-your-repository-part-2-npm-packages) manager such as Sonatype Nexus 3 via setting the NPM_MIRROR_URL environment variable on the slave pod