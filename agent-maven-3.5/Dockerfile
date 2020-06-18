FROM openshift/jenkins-slave-base-centos7:v3.11

MAINTAINER OpenShift Developer Services <jenkins+openshift-dev-services@redhat.com>

ENV MAVEN_VERSION=3.5 \
    BASH_ENV=/usr/local/bin/scl_enable \
    ENV=/usr/local/bin/scl_enable \
    PROMPT_COMMAND=". /usr/local/bin/scl_enable" \
    MAVEN_OPTS="-Duser.home=$HOME"

# Install Maven
RUN INSTALL_PKGS="java-11-openjdk-devel java-1.8.0-openjdk-devel rh-maven35*" && \
    x86_EXTRA_RPMS=$(if [ "$(uname -m)" == "x86_64" ]; then echo -n java-11-openjdk-devel.i686 java-1.8.0-openjdk-devel.i686; fi) && \
    yum install -y centos-release-scl-rh && \
    yum install -y --setopt=protected_multilib=false --enablerepo=centosplus $INSTALL_PKGS $x86_EXTRA_RPMS && \
    rpm -V java-11-openjdk-devel.x86_64 java-1.8.0-openjdk-devel.x86_64 rh-maven35 $x86_EXTRA_RPMS && \
    yum clean all -y && \
    mkdir -p $HOME/.m2 

# When bash is started non-interactively, to run a shell script, for example it
# looks for this variable and source the content of this file. This will enable
# the SCL for all scripts without need to do 'scl enable'.
ADD contrib/bin/* /usr/local/bin/
ADD ./contrib/settings.xml $HOME/.m2/

RUN chown -R 1001:0 $HOME && \
    chmod -R g+rw $HOME

USER 1001
