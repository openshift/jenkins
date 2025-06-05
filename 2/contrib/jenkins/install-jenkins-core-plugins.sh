#! /bin/bash -eu
set -o pipefail


jenkins_version=$(cat /opt/openshift/jenkins-version.txt)
echo "Jenkins version: ${jenkins_version}"
if [[ "${INSTALL_JENKINS_VIA_RPMS}" == false ]]; then
    PLUGIN_LIST="$1"
    echo "Plugin list wil be take from file: " $PLUGIN_LIST
    YUM_FLAGS=" "
    shift  # Shift the script arguments. So $1 will be dropped in favor of $2
    if [ "$#" == "1" ]; then
        YUM_FLAGS="$1"
    fi
    # Builds on Konflux will copy `jenkins.war` from a dependent container image to /usr/lib/jenkins
    # If the file is already present, we skip the upstream RPM installation
    if [[ -f "/usr/lib/jenkins/jenkins.war" ]]; then
        echo "jenkins.war already exists, skipping upstream RPM installation"
    else
        echo "Installing jenkins.war from upstream RPM"
	curl https://pkg.jenkins.io/redhat-stable/jenkins.repo -o /etc/yum.repos.d/jenkins.repo
        rpm --import https://pkg.jenkins.io/redhat-stable/jenkins-ci.org.key
        rpm --import https://pkg.jenkins.io/redhat-stable/jenkins.io.key
        rpm --import https://pkg.jenkins.io/redhat-stable/jenkins.io-2023.key

        yum -y $YUM_FLAGS --setopt=tsflags=nodocs --disableplugin=subscription-manager install jenkins-${jenkins_version}
        yum $YUM_FLAGS clean all
    fi    
    
    /usr/local/bin/install-plugins.sh $PLUGIN_LIST
else
    yum install -y --disableplugin=subscription-manager jenkins-2.* jenkins-2-plugins
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
