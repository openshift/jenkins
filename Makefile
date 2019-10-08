# The master images follow the normal numbering scheme in which the
# major version is used as the directory name and incorporated into
# the image name (jenkins-2-centos7 in this case).  For the slave
# images we are not versioning them (they actually pull their
# jars from the jenkins master, so they don't have a jenkins version,
# so the only thing we'd version is the maven/nodejs version).
# Since these are basically samples we are just going to maintain one
# version (at least that is the initial goal).  This naming system
# can be revisited in the future if we decide we need either jenkins
# or <platform> version numbers in the names.
VERSIONS="2 slave-base agent-maven-3.5 agent-nodejs-8 agent-nodejs-10"

ifeq ($(TARGET),rhel7)
	OS := rhel7
else
	OS := centos7
endif

.PHONY: build
build:
	VERSIONS=$(VERSIONS) hack/build.sh $(OS) $(VERSION)

.PHONY: test
test:
	VERSIONS=$(VERSIONS) TAG_ON_SUCCESS=$(TAG_ON_SUCCESS) TEST_MODE=true hack/build.sh $(OS) $(VERSION)
