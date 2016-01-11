FROM centos:centos7

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

ENV JENKINS_VERSION=1.625 \
    HOME=/var/lib/jenkins \
    JENKINS_HOME=/var/lib/jenkins

LABEL k8s.io.description="Jenkins is a continuous integration server" \
      k8s.io.display-name="Jenkins 1.625" \
      openshift.io.expose-services="8080:http" \
      openshift.io.tags="jenkins,jenkins1,ci" \
      io.openshift.s2i.scripts-url=image:///usr/libexec/s2i

# 8080 for main web interface, 50000 for slave agents
EXPOSE 8080 50000

RUN curl http://pkg.jenkins-ci.org/redhat/jenkins.repo -o /etc/yum.repos.d/jenkins.repo && \
    rpm --import http://pkg.jenkins-ci.org/redhat/jenkins-ci.org.key && \
    yum -y --setopt=tsflags=nodocs install epel-release && \
    yum -y --setopt=tsflags=nodocs install gettext git tar zip unzip java-1.8.0-openjdk jenkins-1.625 nss_wrapper && \
    yum clean all  && \
    localedef -f UTF-8 -i en_US en_US.UTF-8 && \
    curl -L https://github.com/openshift/origin/releases/download/v1.0.8/openshift-origin-v1.0.8-6a2b026-linux-amd64.tar.gz | tar -zx && \
    mv oc /usr/local/bin && \
    rm oadm && \
    rm openshift

COPY ./contrib/openshift /opt/openshift
COPY ./contrib/jenkins /usr/local/bin
ADD ./contrib/s2i /usr/libexec/s2i

RUN /usr/local/bin/plugins.sh /opt/openshift/base-plugins.txt && \
    chown -R 1001:0 /opt/openshift && \
    /usr/local/bin/fix-permissions /opt/openshift && \
    /usr/local/bin/fix-permissions /var/lib/jenkins

VOLUME ["/var/lib/jenkins"]

USER 1001
CMD ["/usr/libexec/s2i/run"]
