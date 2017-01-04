FROM openshift/jenkins-slave-base-centos7

MAINTAINER Adam Saleh <asaleh@redhat.com>

# Labels consumed by Red Hat build service
LABEL com.redhat.component="jenkins-slave-ruby-rhel7-docker" \
      name="openshift3/jenkins-slave-ruby-rhel7" \
      version="1" \
      architecture="x86_64" \
      release="1" \
      io.k8s.display-name="Jenkins Slave Ruby" \
      io.k8s.description="The jenkins slave ruby image has the ruby tools on top of the jenkins slave base image." \
      io.openshift.tags="openshift,jenkins,slave,ruby"

ENV RUBY_VERSION=2.3 \
    BASH_ENV=/usr/local/bin/scl_enable \
    ENV=/usr/local/bin/scl_enable \
    PROMPT_COMMAND=". /usr/local/bin/scl_enable"

COPY contrib/bin/scl_enable /usr/local/bin/scl_enable

# Install Ruby
RUN yum install -y centos-release-scl-rh && \
    INSTALL_PKGS="rh-ruby23 rh-ruby23-ruby-devel rh-ruby23-rubygem-rake rh-ruby23-rubygem-bundler" && \
    yum install -y --setopt=tsflags=nodocs $INSTALL_PKGS && \
    rpm -V $INSTALL_PKGS && \
    yum clean all -y

RUN chown -R 1001:0 $HOME && \
    chmod -R g+rw $HOME

USER 1001
