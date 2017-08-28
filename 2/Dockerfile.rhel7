FROM registry.access.redhat.com/openshift3/ose

# Jenkins image for OpenShift
#
# This image provides a Jenkins server, primarily intended for integration with
# OpenShift v3.
#
# Volumes: 
# * /var/jenkins_home
# Environment:
# * $JENKINS_PASSWORD - Password for the Jenkins 'admin' user.

MAINTAINER Ben Parees <bparees@redhat.com>

ENV JENKINS_VERSION=2.46 \
    HOME=/var/lib/jenkins \
    JENKINS_HOME=/var/lib/jenkins \
    JENKINS_UC=https://updates.jenkins-ci.org \
    LANG=en_US.UTF-8 \
    LC_ALL=en_US.UTF-8

LABEL io.k8s.description="Jenkins is a continuous integration server" \
      io.k8s.display-name="Jenkins 2.46" \
      io.openshift.tags="jenkins,jenkins2,ci" \
      io.openshift.expose-services="8080:http" \
      io.openshift.s2i.scripts-url=image:///usr/libexec/s2i

# Labels consumed by Red Hat build service
LABEL com.redhat.component="openshift-jenkins-docker" \
      name="openshift3/jenkins-2-rhel7" \
      version="3.6" \
      architecture="x86_64" \
      release="7"

# 8080 for main web interface, 50000 for slave agents
EXPOSE 8080 50000

RUN yum-config-manager --disable epel >/dev/null || : && \
    yum-config-manager --enable rhel-7-server-ose-onlineint-rpms || : && \
    INSTALL_PKGS="wget rsync gettext git tar zip unzip java-1.8.0-openjdk java-1.8.0-openjdk.i686 java-1.8.0-openjdk-devel java-1.8.0-openjdk-devel.i686 atomic-openshift-clients jenkins-2.46.3 jenkins-2-plugins" && \
    yum install -y $INSTALL_PKGS && \
    rpm -V $INSTALL_PKGS && \
    yum clean all  && \
    localedef -f UTF-8 -i en_US en_US.UTF-8

COPY ./contrib/openshift /opt/openshift
COPY ./contrib/jenkins /usr/local/bin
ADD ./contrib/s2i /usr/libexec/s2i

# Remove the base-plugins.txt file because it's only used for Centos
# and its presence in the rhel image is confusing.
RUN rm /opt/openshift/base-plugins.txt && \ 
    mkdir -p /opt/openshift/plugins && \
    # NOTE: When adding new Jenkins plugin, you have to create the symlink for the
    # HPI file created by rpm to /opt/openshift/plugins folder.
    for FILENAME in /usr/lib/jenkins/*hpi ; do ln -s $FILENAME /opt/openshift/plugins/`basename $FILENAME .hpi`.jpi; done &&\
    chmod 775 /etc/passwd && \
    chmod -R 775 /etc/alternatives && \
    chmod -R 775 /var/lib/alternatives && \
    chmod -R 775 /usr/lib/jvm && \
    chmod 775 /usr/bin && \
    chmod 775 /usr/lib/jvm-exports && \
    chmod 775 /usr/share/man/man1 && \
    chmod 775 /var/lib/origin && \
    unlink /usr/bin/java && \
    unlink /usr/bin/jjs && \
    unlink /usr/bin/keytool && \
    unlink /usr/bin/orbd && \
    unlink /usr/bin/pack200 && \
    unlink /usr/bin/policytool && \
    unlink /usr/bin/rmid && \
    unlink /usr/bin/rmiregistry && \
    unlink /usr/bin/servertool && \
    unlink /usr/bin/tnameserv && \
    unlink /usr/bin/unpack200 && \
    unlink /usr/lib/jvm-exports/jre && \
    unlink /usr/share/man/man1/java.1.gz && \
    unlink /usr/share/man/man1/jjs.1.gz && \
    unlink /usr/share/man/man1/keytool.1.gz && \
    unlink /usr/share/man/man1/orbd.1.gz && \
    unlink /usr/share/man/man1/pack200.1.gz && \
    unlink /usr/share/man/man1/policytool.1.gz && \
    unlink /usr/share/man/man1/rmid.1.gz && \
    unlink /usr/share/man/man1/rmiregistry.1.gz && \
    unlink /usr/share/man/man1/servertool.1.gz && \
    unlink /usr/share/man/man1/tnameserv.1.gz && \
    unlink /usr/share/man/man1/unpack200.1.gz && \
    # need to create <plugin>.pinned files when upgrading "core" plugins like credentials or subversion that are bundled with the jenkins server
    # Currently jenkins v2 does not embed any plugins, but for reference:
    # touch /opt/openshift/plugins/credentials.jpi.pinned && \
    rmdir /var/log/jenkins && \
    /usr/local/bin/fix-permissions /opt/openshift && \
    chown -R 1001:0 /opt/openshift && \
    # the prior chown doesn't traverse the /opt/openshift/plugins links .. this one will assist fix-permission/assemble for extension builds like master/slave
    chown 1001:0 /usr/lib/jenkins/*hpi && \ 
    /usr/local/bin/fix-permissions /var/lib/jenkins && \
    /usr/local/bin/fix-permissions /var/log

VOLUME ["/var/lib/jenkins"]

USER 1001
ENTRYPOINT []
CMD ["/usr/libexec/s2i/run"]
