FROM openshift/jenkins-slave-base-centos7

MAINTAINER Petter Abrahamsson <petter@redhat.com>

ENV PYTHON_VERSION=3.6.3 \
    BASH_ENV=/usr/local/bin/scl_enable \
    ENV=/usr/local/bin/scl_enable \
    PROMPT_COMMAND=". /usr/local/bin/scl_enable"

COPY contrib/bin/scl_enable /usr/local/bin/scl_enable

USER 0
# install python
RUN yum install -y centos-release-scl-rh && \
    INSTALL_PKGS="rh-python36-python rh-python36-python-virtualenv" && \
    yum install -y --setopt=tsflags=nodocs $INSTALL_PKGS && \
    rpm -V $INSTALL_PKGS && \
    yum clean all -y

RUN chown -R 1001:0 $HOME && \
    chmod -R g+rw $HOME

USER 1001
