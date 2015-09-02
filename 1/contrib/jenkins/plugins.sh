#! /bin/bash
#
# Originally copied from https://github.com/jenkinsci/docker
# You can set JENKINS_UC to change the default URL to Jenkins update center
#
# Usage:
#
# FROM openshift/jenkins-1-centos7
# COPY plugins.txt /plugins.txt
# RUN /usr/local/bin/plugins.sh /plugins.txt
#
# The format of 'plugins.txt. is:
#
# pluginId:pluginVersion

set -e

# TODO: Move this Dockerfile
JENKINS_UC=${JENKINS_UC:-https://updates.jenkins-ci.org}

while read spec || [ -n "$spec" ]; do
    plugin=(${spec//:/ });
    [[ ${plugin[0]} =~ ^# ]] && continue
    [[ ${plugin[0]} =~ ^\s*$ ]] && continue
    [[ -z ${plugin[1]} ]] && plugin[1]="latest"
    echo "Downloading ${plugin[0]}:${plugin[1]}"

    if [ -z "$JENKINS_UC_DOWNLOAD" ]; then
      JENKINS_UC_DOWNLOAD=$JENKINS_UC/download
    fi
    mkdir -p /opt/openshift/plugins
    curl -sSL -f ${JENKINS_UC_DOWNLOAD}/plugins/${plugin[0]}/${plugin[1]}/${plugin[0]}.hpi \
      -o /opt/openshift/plugins/${plugin[0]}.jpi
done  < $1
chmod -R og+rw /opt/openshift/plugins
