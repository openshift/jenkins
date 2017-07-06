FROM registry.access.redhat.com/openshift3/jenkins-slave-base-rhel7

MAINTAINER Ben Parees <bparees@redhat.com>

# Labels consumed by Red Hat build service
LABEL com.redhat.component="jenkins-slave-maven-rhel7-docker" \
      name="openshift3/jenkins-slave-maven-rhel7" \
      version="3.6" \
      architecture="x86_64" \
      release="4" \
      io.k8s.display-name="Jenkins Slave Maven" \
      io.k8s.description="The jenkins slave maven image has the maven tools on top of the jenkins slave base image." \
      io.openshift.tags="openshift,jenkins,slave,maven"

ENV MAVEN_VERSION=3.3 \
    BASH_ENV=/usr/local/bin/scl_enable \
    ENV=/usr/local/bin/scl_enable \
    PROMPT_COMMAND=". /usr/local/bin/scl_enable"

# Install Maven
RUN yum-config-manager --enable rhel-server-rhscl-7-rpms && \    
    yum-config-manager --disable epel >/dev/null || : && \
    INSTALL_PKGS="java-1.8.0-openjdk-devel rh-maven33*" && \
    yum install -y $INSTALL_PKGS && \
    rpm -V ${INSTALL_PKGS//\*/} && \
    yum clean all -y && \
    mkdir -p $HOME/.m2

# When bash is started non-interactively, to run a shell script, for example it
# looks for this variable and source the content of this file. This will enable
# the SCL for all scripts without need to do 'scl enable'.
ADD contrib/bin/scl_enable /usr/local/bin/scl_enable
ADD contrib/bin/configure-slave /usr/local/bin/configure-slave
ADD ./contrib/settings.xml $HOME/.m2/

RUN chown -R 1001:0 $HOME && \
    chmod -R g+rw $HOME

USER 1001
