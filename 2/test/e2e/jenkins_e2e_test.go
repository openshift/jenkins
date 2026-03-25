package e2e

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"testing"
)

const (
	adminUsername = "admin"
	adminPassword = "password"
)

func TestDefaultUsernamePasswordLogin(t *testing.T) {
	ta := &testArgs{
		t: t,
		templateParams: map[string]string{
			"ENABLE_OAUTH": "false",
		},
	}
	ta = setupThroughJenkinsLaunch(ta.t, ta)
	registerCleanup(ta)

	auth := basicAuth(adminUsername, adminPassword)
	jenkinsURL := getJenkinsURL(ta)
	waitForJenkinsAPI(ta, jenkinsURL, "/api/json")
	checkJenkinsLogin(ta, jenkinsURL, auth)
	logJenkinsVersion(ta, jenkinsURL, auth)
}

func TestUsernamePasswordLoginPersistent(t *testing.T) {
	ta := &testArgs{
		t:        t,
		template: "jenkins-persistent",
		templateParams: map[string]string{
			"ENABLE_OAUTH":             "false",
			"VOLUME_CAPACITY":          "1Gi",
			"JENKINS_IMAGE_STREAM_TAG": "jenkins:oldversion",
		},
		upgradeImageStreamTag: "jenkins:latest",
	}
	ta = setupThroughJenkinsLaunch(ta.t, ta)
	registerCleanup(ta)
	jenkinsURL := getJenkinsURL(ta)

	auth := basicAuth(adminUsername, adminPassword)

	// Step 1: Verify login with default credentials on the old image
	t.Log("step 1: verifying login with default credentials on pre-upgrade image")
	waitForJenkinsAPI(ta, jenkinsURL, "/api/json")
	checkJenkinsLogin(ta, jenkinsURL, auth)
	logJenkinsVersion(ta, jenkinsURL, auth)

	// Step 2: Change the admin password
	t.Log("step 2: changing admin password")
	newPassword := "upgraded-admin-password"
	changeJenkinsPassword(ta, jenkinsURL, adminUsername, newPassword, auth)

	// Step 3: Verify the new password works before upgrade
	t.Log("step 3: verifying login with new password before upgrade")
	newAuth := basicAuth(adminUsername, newPassword)
	checkJenkinsLogin(ta, jenkinsURL, newAuth)

	// Step 4: Upgrade Jenkins to the new image
	t.Logf("step 4: upgrading Jenkins from jenkins:oldversion to %s", ta.upgradeImageStreamTag)
	upgradeJenkinsImage(ta)

	// Step 5: Verify the changed password still works after upgrade (PVC persistence)
	t.Log("step 5: verifying login with changed password after upgrade")
	waitForJenkinsAPI(ta, jenkinsURL, "/api/json")
	checkJenkinsLogin(ta, jenkinsURL, newAuth)
	logJenkinsVersion(ta, jenkinsURL, newAuth)
	t.Log("password persisted successfully across Jenkins upgrade")
}

func TestOCVersionSwitching(t *testing.T) {
	ta := &testArgs{t: t}
	ta = setupThroughJenkinsLaunch(ta.t, ta)
	registerCleanup(ta)

	jenkinsURL := getJenkinsURL(ta)
	waitForJenkinsAPI(ta, jenkinsURL, "/api/json")

	token := getOAuthToken(ta)
	ocAuth := bearerAuth(token)
	checkJenkinsLogin(ta, jenkinsURL, ocAuth)

	bcName := "oc-version-switch-buildconfig"

	createJenkinsPipelineBuildConfig(ta, bcName, ocVersionPipeline)

	// First build: no tool specified, uses the latest oc bundled in the image.
	// Parse the minor version from "Client Version: 4.XX.Y" in the log.

	buildName := startBuild(ta, bcName)
	waitForBuildComplete(ta, buildName)
	buildLog := getBuildLog(ta, buildName, ocAuth)

	rgx := regexp.MustCompile(`Client Version:\s*4\.(\d+)`)
	matches := rgx.FindStringSubmatch(buildLog)
	t.Logf("match logs: %s", matches)
	if len(matches) < 2 {
		t.Fatalf("could not parse oc minor version from default build log:\n%s", buildLog)
	}
	latestMinor, err := strconv.Atoi(matches[1])
	if err != nil {
		t.Fatalf("invalid minor version %q: %v", matches[1], err)
	}
	t.Logf("detected latest oc version: 4.%d", latestMinor)

	const firstMinor = 16
	for minor := firstMinor; minor < latestMinor; minor++ {
		ocTool := fmt.Sprintf("oc-4.%d", minor)
		expectedVersion := fmt.Sprintf("4.%d", minor)
		t.Run(ocTool, func(t *testing.T) {
			updateBuildConfigJenkinsfile(ta, bcName, ocVersionWithToolPipeline(ocTool))
			buildName := startBuild(ta, bcName)
			waitForBuildComplete(ta, buildName)
			assertBuildLog(ta, buildName, ocAuth, expectedVersion)
		})
	}
}

func TestSmokeTest(t *testing.T) {
	ta := &testArgs{t: t}
	ta = setupThroughJenkinsLaunch(ta.t, ta)
	registerCleanup(ta)

	jenkinsURL := getJenkinsURL(ta)
	waitForJenkinsAPI(ta, jenkinsURL, "/api/json")

	token := getOAuthToken(ta)
	auth := bearerAuth(token)

	t.Log("verifying expected plugins are installed")
	checkPluginsInstalled(ta, jenkinsURL, auth, basePlugins)

	t.Log("creating a test job via REST API")
	status := createJenkinsJob(ta, jenkinsURL, "testJob", auth, testJobXML)
	if status != http.StatusOK {
		ta.t.Fatalf("expected 200 when creating testJob, got %d", status)
	}

	// t.Log("verifying the test job exists")
	// status = getJenkinsJob(ta, jenkinsURL, "testJob", auth)
	// if status != http.StatusOK {
	// 	ta.t.Fatalf("expected 200 when getting testJob, got %d", status)
	// }

	t.Log("verifying job creation fails with invalid authentication")
	invalidAuth := bearerAuth("invalid-token")
	status = createJenkinsJob(ta, jenkinsURL, "failJob", invalidAuth, testJobXML)
	if status != http.StatusUnauthorized {
		ta.t.Fatalf("expected 401 when creating job with invalid token, got %d", status)
	}

	t.Log("verifying non-existent job returns 404")
	status = getJenkinsJob(ta, jenkinsURL, "failJob", auth)
	if status != http.StatusNotFound {
		ta.t.Fatalf("expected 404 for non-existent job failJob, got %d", status)
	}

	t.Log("running a pipeline build as final smoke check")
	bcName := "smoke-test-buildconfig"
	createJenkinsPipelineBuildConfig(ta, bcName, smokeTestPipeline)

	buildName := startBuild(ta, bcName)
	waitForBuildComplete(ta, buildName)
	assertBuildLog(ta, buildName, auth, "Jenkins smoke test passed")
}
