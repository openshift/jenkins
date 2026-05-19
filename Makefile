# The master images follow the normal numbering scheme in which the
# major version is used as the directory name and incorporated into
# the image name (jenkins-2-rhel9 in this case).  For the slave
# images we are not versioning them (they actually pull their
# jars from the jenkins master, so they don't have a jenkins version,
# so the only thing we'd version is the maven/nodejs version).
# Since these are basically samples we are just going to maintain one
# version (at least that is the initial goal).  This naming system
# can be revisited in the future if we decide we need either jenkins
# or <platform> version numbers in the names.
VERSIONS="2 slave-base"

BUNDLE_PLUGINS="$(shell pwd)/2/contrib/openshift/bundle-plugins.txt"
REF=$(shell mktemp -d)
JENKINS_WAR="$(shell mktemp -d)/jenkins.war"
ifeq ($(TARGET),rhel9)
	OS := rhel9
else ifeq ($(TARGET),rhel8)
	OS := rhel8
else
 	OS := ubi
endif

.PHONY: build
build:
	VERSIONS=$(VERSIONS) hack/build.sh $(OS) $(VERSION)

.PHONY: test
test:
	VERSIONS=$(VERSIONS) TAG_ON_SUCCESS=$(TAG_ON_SUCCESS) TEST_MODE=true hack/build.sh $(OS) $(VERSION)

.PHONY: smoke
smoke:
	@echo "Testing the jenkins template based install on openshift"
	@./scripts/test-jenkins-template-install.sh

.PHONY: e2e
e2e:
	@echo "Starting e2e tests from 2/test directory"
	@echo "IMAGE_NAME set in environment variable with value: $(IMAGE_NAME)"
	@cd 2/test && go test

.PHONY: plugins-list
plugins-list:
	@echo "Do not use this command, manually update base-plugins.txt and bundle-plugins.txt to be the same"

.PHONY: verify
verify:
	./scripts/verify.sh

.PHONY: build-development-image
build-development-images:
	./scripts/build-development-images.sh
