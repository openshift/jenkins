#!/bin/sh
#
# This script, along with the Dockerfile in this directory, enables
# the use of the openshift internal ci system, and in particular
#
# - prow
# - ci-operator (https://github.com/openshift/ci-operator)
# - configuration for ci-operator defined in https://github.com/openshift/release
# - images
# - openshift api objects
#
# to facilitate the use of test suites defined in openshift/origin
# (https://github.com/openshift/origin) against images produced
# by builds of this repository.
#
# 
# The flow is as follows:
#
# - ci-operator takes a git branch ref for this repo and config specific to this repo
#   and produces the jenkins related images from this repo using that git branch ref,
#   via the openshift docker build strategy (eating our dog food)

# - ci-operator also generates parameter settings for a template in
#   https://github.com/openshift/release; those settings will contain
#   docker pull specs for the images that were built, as well as the
#   test image to use

# - When running the tests, the template is instantiated and sets up
#   a test cluster for execution, using the test image noted above,
#   of tests from openshift/origin; for our purposes, we want to run
#   the e2e/extended tests defined in openshift/origin for openshift
#   pipeline strategy builds and other jenkins/openshift integration 
#   features

# - The template has some plug points (via env vars) to run customizable
#   commands that can massage the set up of the test cluster prior to
#   the running of the tests ... to "prepare" the cluster after it is installed

# - Those plug points are specified via the ci-operator config for
#   jenkins stored in http://github.com/openshift/release

# - In particular, the test image used to run the tests is defined,
#   and for the case of jenkins, we take a base test image defined
#   by the openshift internal CI infrastructure, and extend it 
#   via an openshift docker strategy build, using the Dockerfile
#   in this directory, to inject this script into the test image used
#   as  the "prepare" plug point

# - When this script is called prior to running the tests,
#   it performs the "preparation" we need.

# - namely, the "preparation" is tagging the image to be tested,
#   i.e. the new jenkins image we built
#   in the first step of this flow, which the template seeding noted
#   above insures is set to the env var "IMAGE_PREPARE",
#   into the jenkins image stream in the openshift namespace
#   in the test cluster

# - our existing openshift/origin tests will then pick up this new
#   test image as part of running the tests
#

oc tag --source=docker $IMAGE_PREPARE openshift/jenkins:2
oc tag --source=docker $IMAGE_PREPARE openshift/jenkins:latest
