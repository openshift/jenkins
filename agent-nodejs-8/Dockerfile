FROM openshift/jenkins-slave-base-centos7

MAINTAINER Gabe Montero <gmontero@redhat.com>

ENV NODEJS_VERSION=8 \
    NPM_CONFIG_PREFIX=$HOME/.npm-global \
    PATH=$HOME/node_modules/.bin/:$HOME/.npm-global/bin/:$PATH \
    BASH_ENV=/usr/local/bin/scl_enable \
    ENV=/usr/local/bin/scl_enable \
    PROMPT_COMMAND=". /usr/local/bin/scl_enable" \
    CHROME_BIN=/bin/google-chrome

COPY contrib/bin/scl_enable /usr/local/bin/scl_enable
ADD https://dl.google.com/linux/direct/google-chrome-stable_current_x86_64.rpm google-chrome-stable_current_x86_64.rpm

# Install NodeJS & Google Chrome
RUN yum install -y centos-release-scl-rh && \
    INSTALL_PKGS="rh-nodejs${NODEJS_VERSION} rh-nodejs${NODEJS_VERSION}-npm rh-nodejs${NODEJS_VERSION}-nodejs-nodemon make gcc-c++ nss_wrapper redhat-lsb libXScrnSaver xdg-utils" && \
    ln -s /usr/lib/node_modules/nodemon/bin/nodemon.js /usr/bin/nodemon && \
    yum install -y --setopt=tsflags=nodocs $INSTALL_PKGS && \
    yum -y localinstall google-chrome-stable_current_x86_64.rpm && \
    rm google-chrome-stable_current_x86_64.rpm && \
    rpm -V $INSTALL_PKGS google-chrome-stable && \
    yum clean all -y


RUN chown -R 1001:0 $HOME && \
    chmod -R g+rw $HOME

USER 1001
