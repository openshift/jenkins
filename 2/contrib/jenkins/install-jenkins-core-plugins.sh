#! /bin/bash -eu

set -o pipefail

if [[ "${INSTALL_JENKINS_VIA_RPMS}" == "false" ]]; then
    curl https://pkg.jenkins.io/redhat-stable/jenkins.repo -o /etc/yum.repos.d/jenkins.repo
    rpm --import https://pkg.jenkins.io/redhat-stable/jenkins-ci.org.key
    yum -y --setopt=tsflags=nodocs install jenkins-2.150.2-1.1
    rpm -V jenkins-2.150.2-1.1
    yum clean all
    /usr/local/bin/install-plugins.sh "$@"
else
    yum install -y jenkins-2.* jenkins-2-plugins
    rpm -V jenkins-2.* jenkins-2-plugins
    yum clean all
fi


