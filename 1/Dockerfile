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

ENV JENKINS_VERSION=1.609 \
    HOME=/var/lib/jenkins \
    JENKINS_HOME=/var/lib/jenkins

LABEL k8s.io.description="Jenkins is a continuous integration server" \
      k8s.io.display-name="Jenkins 1.609" \
      openshift.io.expose-services="8080:http" \
      openshift.io.tags="jenkins,jenkins1,ci"

# 8080 for main web interface, 50000 for slave agents
EXPOSE 8080 50000

#RUN  curl -L -o /tmp/git.hpi http://updates.jenkins-ci.org/latest/git.hpi
#RUN  curl -L -o /tmp/git-client.hpi http://updates.jenkins-ci.org/latest/git-client.hpi
#RUN  curl -L -o /tmp/scm-api.hpi http://updates.jenkins-ci.org/latest/scm-api.hpi
#RUN  curl -L -o /tmp/ssh-credentials.hpi http://updates.jenkins-ci.org/latest/ssh-credentials.hpi
#RUN  curl -L -o /tmp/credentials.hpi http://updates.jenkins-ci.org/latest/credentials.hpi

VOLUME ["/var/lib/jenkins"]

RUN curl http://pkg.jenkins-ci.org/redhat/jenkins.repo -o /etc/yum.repos.d/jenkins.repo && \
    rpm --import http://pkg.jenkins-ci.org/redhat/jenkins-ci.org.key && \
    yum install -y tar zip unzip java-1.8.0-openjdk jenkins-1.609 && \
    yum clean all  && \
    chmod og+rw /var/lib/jenkins && \
    curl -L https://github.com/openshift/origin/releases/download/v1.0.2/openshift-origin-v1.0.2-325c7b7-linux-amd64.tar.gz | tar -zx && \
    mv oc /usr/local/bin && \
    rm oadm && \
    rm openshift

COPY ./contrib/openshift /opt/openshift
COPY ./contrib/jenkins /usr/local/bin

RUN chmod -R og+rw /opt/openshift

USER 1001

ENTRYPOINT ["/usr/local/bin/run-jenkins"]
