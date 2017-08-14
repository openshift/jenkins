FROM openshift/jenkins-slave-base-centos7

MAINTAINER Ben Parees <bparees@redhat.com>

ENV MAVEN_VERSION=3.3 \
    GRADLE_VERSION=3.5 \
    BASH_ENV=/usr/local/bin/scl_enable \
    ENV=/usr/local/bin/scl_enable \
    PROMPT_COMMAND=". /usr/local/bin/scl_enable" \
    PATH=$PATH:/opt/gradle-3.5/bin

# Install Maven
RUN INSTALL_PKGS="java-1.8.0-openjdk-devel rh-maven33*" && \
    yum install -y centos-release-scl-rh && \
    yum install -y --enablerepo=centosplus $INSTALL_PKGS && \
    curl -LOk https://services.gradle.org/distributions/gradle-${GRADLE_VERSION}-bin.zip && \
    unzip gradle-${GRADLE_VERSION}-bin.zip -d /opt && \
    rm -f gradle-${GRADLE_VERSION}-bin.zip && \
    rpm -V ${INSTALL_PKGS//\*/} && \
    yum clean all -y && \
    mkdir -p $HOME/.m2 && \
    mkdir -p $HOME/.gradle

# When bash is started non-interactively, to run a shell script, for example it
# looks for this variable and source the content of this file. This will enable
# the SCL for all scripts without need to do 'scl enable'.
ADD contrib/bin/scl_enable /usr/local/bin/scl_enable
ADD contrib/bin/configure-slave /usr/local/bin/configure-slave
ADD ./contrib/settings.xml $HOME/.m2/
ADD ./contrib/init.gradle $HOME/.gradle/

RUN chown -R 1001:0 $HOME && \
    chmod -R g+rw $HOME

USER 1001
