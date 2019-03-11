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
    # Remove the base-plugins.txt file because it's only used for Centos
    # and its presence in the rhel image is confusing.
    rm /opt/openshift/base-plugins.txt
    mkdir -p /opt/openshift/plugins
    # we symlink the rpm installed plugins from /usr/lib/jenkins to /opt/openshift/plugins so that
    # future upgrades of the image and their RPM install automatically get picked by jenkins;
    # we use symlinks vs. actual files to delineate whether the user has overridden a plugin (and
    # by extension taken over its future maintenance)
    for FILENAME in /usr/lib/jenkins/*hpi ; do ln -s $FILENAME /opt/openshift/plugins/`basename $FILENAME .hpi`.jpi; done
    chown 1001:0 /usr/lib/jenkins/*hpi
fi


