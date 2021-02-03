Maven Slave Image
====================

This repository contains Dockerfiles for a Jenkins Agent Docker image intended for 
use with [OpenShift v3](https://github.com/openshift/origin)

Maven Repository Support
---------------------------------
This Maven agent image supports using a Maven Mirror/repository manager such as Sonatype Nexus via setting the MAVEN_MIRROR_URL environment variable on the slave pod