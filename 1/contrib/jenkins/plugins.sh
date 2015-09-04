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
DOWNLOAD_JOBS=()
mkdir -p /opt/openshift/plugins

while read spec || [ -n "$spec" ]; do
    plugin=(${spec//:/ });
    [[ ${plugin[0]} =~ ^# ]] && continue
    [[ ${plugin[0]} =~ ^\s*$ ]] && continue
    [[ -z ${plugin[1]} ]] && plugin[1]="latest"

    if [ -z "$JENKINS_UC_DOWNLOAD" ]; then
      JENKINS_UC_DOWNLOAD=$JENKINS_UC/download
    fi

    curl -sSL -f ${JENKINS_UC_DOWNLOAD}/plugins/${plugin[0]}/${plugin[1]}/${plugin[0]}.hpi \
      -o /opt/openshift/plugins/${plugin[0]}.jpi &>/dev/null &
    DOWNLOAD_JOBS+=("$!:${plugin[0]}-${plugin[1]}")
done  < $1

# Now wait for all downloads to complete
for job in "${DOWNLOAD_JOBS[@]}"; do
  job_id=$(echo $job | cut -d ':' -f 1)
  job_name=$(echo $job | cut -d ':' -f 2)

  echo "Downloading ${job_name} ..."
  wait ${job_id}
  result=$?
  if [ "$result" != "0" ]; then
    echo "Failed to download ${job_name}: ${result}" && exit ${result}
  fi
done

chmod -R og+rw /opt/openshift/plugins
