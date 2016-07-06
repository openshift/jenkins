FROM openshift/jenkins-slave-base-centos7

MAINTAINER Ben Parees <bparees@redhat.com>

ENV MAVEN_VERSION=3.3.9

# Install Maven
RUN INSTALL_PKGS="java-1.8.0-openjdk-devel" && \
    yum install -y --enablerepo=centosplus $INSTALL_PKGS && \
    rpm -V $INSTALL_PKGS && \
    yum clean all -y && \
    (curl -v https://www.apache.org/dist/maven/maven-3/$MAVEN_VERSION/binaries/apache-maven-$MAVEN_VERSION-bin.tar.gz | \
    tar -zx -C /usr/local) && \
    ln -sf /usr/local/apache-maven-$MAVEN_VERSION/bin/mvn /usr/local/bin/mvn && \
    mkdir -p $HOME/.m2

ADD ./contrib/settings.xml $HOME/.m2/

RUN chown -R 1001:0 $HOME && \
    chmod -R g+rw $HOME

USER 1001
