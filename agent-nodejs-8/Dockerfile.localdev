FROM quay.io/openshift/origin-jenkins-agent-base:v4.0

MAINTAINER Akram Ben Aissi <abenaiss@redhat.com>

ENV NODEJS_VERSION=8 \
    NPM_CONFIG_PREFIX=$HOME/.npm-global \
    PATH=$HOME/node_modules/.bin/:$HOME/.npm-global/bin/:$PATH \
    BASH_ENV=/usr/local/bin/scl_enable \
    ENV=/usr/local/bin/scl_enable \
    PROMPT_COMMAND=". /usr/local/bin/scl_enable" \
    LANG=en_US.UTF-8 \
    LC_ALL=en_US.UTF-8


COPY contrib/bin/scl_enable /usr/local/bin/scl_enable
COPY contrib/bin/configure-agent /usr/local/bin/configure-agent
COPY contrib/openshift/CentOS-Base.repo /etc/yum.repos.d/CentOS-Base.repo

# Install NodeJS
RUN  curl http://mirror.centos.org/centos-7/7/os/x86_64/RPM-GPG-KEY-CentOS-7 -o /etc/pki/rpm-gpg/RPM-GPG-KEY-CentOS-7 && \
    INSTALL_PKGS="make gcc-c++" && \
    DISABLES="--disablerepo=rhel-server-extras --disablerepo=rhel-server --disablerepo=rhel-fast-datapath --disablerepo=rhel-server-optional --disablerepo=rhel-server-ose --disablerepo=rhel-server-rhscl" && \
    ln -s /usr/lib/node_modules/nodemon/bin/nodemon.js /usr/bin/nodemon && \
    yum $DISABLES install -y --setopt=tsflags=nodocs --disableplugin=subscription-manager $INSTALL_PKGS && \
    rpm -V $INSTALL_PKGS && \
    curl -sL https://rpm.nodesource.com/setup_8.x | bash - && \
    yum $DISABLES install -y --setopt=tsflags=nodocs --disableplugin=subscription-manager nodejs && \
    rpm -V nodejs && \
    yum clean all -y 

RUN chown -R 1001:0 $HOME && \
    chmod -R g+rw $HOME

USER 1001
