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
mkdir -p /tmp/.plugin-download

while read spec || [ -n "$spec" ]; do
    plugin=(${spec//:/ });
    [[ ${plugin[0]} =~ ^# ]] && continue
    [[ ${plugin[0]} =~ ^\s*$ ]] && continue
    [[ -z ${plugin[1]} ]] && plugin[1]="latest"

    if [ -z "$JENKINS_UC_DOWNLOAD" ]; then
      JENKINS_UC_DOWNLOAD=$JENKINS_UC/download
    fi

    name="${plugin[0]}-${plugin[1]}"
    curl -sSL -f ${JENKINS_UC_DOWNLOAD}/plugins/${plugin[0]}/${plugin[1]}/${plugin[0]}.hpi \
      -o /opt/openshift/plugins/${plugin[0]}.jpi &> /tmp/.plugin-download/${name}.log &
    DOWNLOAD_JOBS+=("$!:${name}")
done  < $1

set +e
# Now wait for all downloads to complete
for job in "${DOWNLOAD_JOBS[@]}"; do
  job_id=$(echo $job | cut -d ':' -f 1)
  job_name=$(echo $job | cut -d ':' -f 2)

  echo "Downloading ${job_name} ..."
  wait ${job_id}
  result=$?
  if [ "$result" != "0" ]; then
    echo "Failed to download ${job_name}:"
    cat /tmp/.plugin-download/${job_name}.log 2>/dev/null
    exit ${result}
  fi
done

# Cleanup and make the plugins world writeable
rm -rf /tmp/.plugin-download
chmod -R og+rw /opt/openshift/plugins
