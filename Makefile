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
VERSIONS="2 slave-base"
BUNDLE_PLUGINS="$(shell pwd)/2/contrib/openshift/bundle-plugins.txt"
REF=$(shell mktemp -d)
JENKINS_WAR="$(shell mktemp -d)/jenkins.war"
JENKINS_FILE := "$(shell pwd)/2/contrib/openshift/jenkins-version.txt"
ARTIFACTS_OUTPUT_FILE := $(shell pwd)/artifacts.lock.yaml
ARTIFACTS_IMAGE := localhost/artifacts-gen:latest
PULLSECRET ?=
BASE_IMAGE := registry.redhat.io/openshift4/ose-cli-rhel9:v4.18
RPMS_LOCK_REPO_URL := https://github.com/konflux-ci/rpm-lockfile-prototype.git
RPMS_LOCK_REPO_TMPDIR := $(shell mktemp -d)/rpm-lockfile-prototype
RPMS_LOCK_IN_FILE := $(shell pwd)/rpms.in.yaml
RPMS_LOCK_FILE := $(shell pwd)/rpms.lock.yaml
REPO_FILE := $(shell pwd)/ubi.repo
RPMS_LOCK_IMAGE := localhost/rpm-lockfile-prototype:latest
TEKTON_PIPELINE_FILE := $(shell pwd)/.tekton/build-pipeline.yaml

ifeq ($(TARGET),rhel8)
	OS := rhel8
 else
 	OS := centos8
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

## Build artifact generation image using podman
.PHONY: build-artifact-gen-image
build-artifact-gen-image:
	@echo "Building $(ARTIFACTS_IMAGE) image ..."
	@podman build -f Containerfile -t $(ARTIFACTS_IMAGE) tools/artifacts-gen
	@echo "✓ Built $(ARTIFACTS_IMAGE) image"

## Generate artifacts.lock.yaml using podman
.PHONY: artifacts-gen
artifacts-gen: build-artifact-gen-image
	@echo "Generating $$(basename $(ARTIFACTS_OUTPUT_FILE)) with podman ..."
	@touch $$(basename $(ARTIFACTS_OUTPUT_FILE))
	@podman run --rm \
		--userns=keep-id \
		-v $(BUNDLE_PLUGINS):/data/$$(basename $(BUNDLE_PLUGINS)).txt:ro \
		-v $(JENKINS_FILE):/data/$$(basename $(JENKINS_FILE)).txt:ro \
		-v $(ARTIFACTS_OUTPUT_FILE):/output/$$(basename $(ARTIFACTS_OUTPUT_FILE)):rw \
		$(ARTIFACTS_IMAGE) \
		--plugins /data/$$(basename $(BUNDLE_PLUGINS)).txt \
		--jenkins /data/$$(basename $(JENKINS_FILE)).txt \
		--output /output/$$(basename $(ARTIFACTS_OUTPUT_FILE)) || { rm -f $(ARTIFACTS_OUTPUT_FILE); exit 1; }
	@echo "✓ Generated $$(basename $(ARTIFACTS_OUTPUT_FILE))"

## Verify artifacts.lock.yaml is up-to-date using podman
.PHONY: verify-artifacts
verify-artifacts:
	@echo "Verifying $$(basename $(ARTIFACTS_OUTPUT_FILE)) is up-to-date with podman ..."
	@touch $(ARTIFACTS_OUTPUT_FILE).tmp
	@if podman run --rm \
		--userns=keep-id \
		-v $(BUNDLE_PLUGINS):/data/$$(basename $(BUNDLE_PLUGINS)).txt:ro \
		-v $(JENKINS_FILE):/data/$$(basename $(JENKINS_FILE)).txt:ro \
		-v $(ARTIFACTS_OUTPUT_FILE).tmp:/output/$$(basename $(ARTIFACTS_OUTPUT_FILE)).tmp:rw \
		$(ARTIFACTS_IMAGE) \
		--plugins /data/$$(basename $(BUNDLE_PLUGINS)).txt \
		--jenkins /data/$$(basename $(JENKINS_FILE)).txt \
		--output /output/$$(basename $(ARTIFACTS_OUTPUT_FILE)).tmp; then \
		if diff -q $(ARTIFACTS_OUTPUT_FILE) $(ARTIFACTS_OUTPUT_FILE).tmp; then \
			echo "✓ $$(basename $(ARTIFACTS_OUTPUT_FILE)) is up-to-date"; \
		else \
			echo "✗ $$(basename $(ARTIFACTS_OUTPUT_FILE)) is out of sync!"; \
			echo "  Run: make artifacts-gen"; \
			rm -f $(ARTIFACTS_OUTPUT_FILE).tmp; \
			exit 1; \
		fi; \
		rm -f $(ARTIFACTS_OUTPUT_FILE).tmp; \
	else \
		echo "✗ Failed to generate verification file!"; \
		echo "  Run: make artifacts-gen"; \
		rm -f $(ARTIFACTS_OUTPUT_FILE).tmp; \
		exit 1; \
	fi

.PHONY: build-rpm-lock-image
build-rpm-lock-image:
	@echo "Building $$(basename $(RPMS_LOCK_REPO_TMPDIR)) image ..."
	@rm -rf $(RPMS_LOCK_REPO_TMPDIR)
	@git clone $(RPMS_LOCK_REPO_URL) $(RPMS_LOCK_REPO_TMPDIR)
	@podman build -f $(RPMS_LOCK_REPO_TMPDIR)/Containerfile -t $(RPMS_LOCK_IMAGE) $(RPMS_LOCK_REPO_TMPDIR)
	@echo "✓ Built $(RPMS_LOCK_IMAGE) image"

.PHONY: rpms-lock
rpms-lock: build-rpm-lock-image
	@if [ -z "$(PULLSECRET)" ]; then \
		echo "Error: PULLSECRET is not set. Please provide a path to the pullsecret file."; \
		echo "Usage: make rpms-lock PULLSECRET=/path/to/pullsecret"; \
		exit 1; \
	fi
	@echo "Generating $$(basename $(RPMS_LOCK_FILE)) file with podman ..."
	@podman run --rm \
		--userns=keep-id \
		-e REGISTRY_AUTH_FILE=/work/$$(basename $(PULLSECRET)) \
		-v $(RPMS_LOCK_FILE):/work/$$(basename $(RPMS_LOCK_FILE)):rw \
		-v $(RPMS_LOCK_IN_FILE):/work/$$(basename $(RPMS_LOCK_IN_FILE)):ro \
		-v $(REPO_FILE):/work/$$(basename $(REPO_FILE)):ro \
		-v $(PULLSECRET):/work/$$(basename $(PULLSECRET)):ro \
		$(RPMS_LOCK_IMAGE) \
		--image $(BASE_IMAGE) \
		--outfile=/work/$$(basename $(RPMS_LOCK_FILE)) \
		/work/$$(basename $(RPMS_LOCK_IN_FILE))
	@echo "✓ Generated $$(basename $(RPMS_LOCK_FILE))"

.PHONY: update-tekton-tasks
update-tekton-tasks:
	@echo "Updating Tekton tasks in $(TEKTON_PIPELINE_FILE) file ..."
	@./hack/update-tekton-task-bundles.sh $(TEKTON_PIPELINE_FILE)
	@if [ "$$(uname)" = "Darwin" ]; then rm -f $(TEKTON_PIPELINE_FILE)-e; fi
	@echo "✓ Updated Tekton tasks"
