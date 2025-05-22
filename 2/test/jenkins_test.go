package test

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/jenkins/pkg/jenkins"
	"github.com/openshift/jenkins/pkg/podman"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Jenkins Suite (v2)")
}

var podmancli *podman.Client
var imageName string

var _ = BeforeSuite(func() {
	var err error
	podmancli, err = podman.NewEnvClient()
	Expect(err).NotTo(HaveOccurred())

	imageName = os.Getenv("IMAGE_NAME")
	if imageName == "" {
		imageName = "openshift/jenkins-2-rhel9-candidate"
	}
})

var _ = Describe("Jenkins testing (v2)", func() {
	var j *jenkins.Jenkins
	var imageNamesToRemove []string

	BeforeEach(func() {
		var err error
		j = jenkins.NewJenkins(podmancli)
		vcr, err := podmancli.VolumeCreate()
		j.Volume = vcr.Name
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		if CurrentGinkgoTestDescription().Failed {
			By("printing container logs")
			logs, err := podmancli.ContainerLogs(j.ID)
			Expect(err).NotTo(HaveOccurred())
			_, err = GinkgoWriter.Write(logs)
			Expect(err).NotTo(HaveOccurred())
		}

		_, err := podmancli.ContainerStopAndRemove(j.ID, 60)
		Expect(err).NotTo(HaveOccurred())

		err = podmancli.VolumeRemove(j.Volume)
		Expect(err).NotTo(HaveOccurred())

		for _, imageName := range imageNamesToRemove {
			_, errs := podmancli.ImagesRemove([]string{imageName})
			Expect(len(errs)).To(Equal(0))
		}
		imageNamesToRemove = nil
	})

	basePlugins := []string{
		"ace-editor",
		"branch-api",
		"credentials",
		"durable-task",
		"cloudbees-folder",
		"git",
		"git-server",
		"git-client",
		"plain-credentials",
		"scm-api",
		"script-security",
		"structs",
	}

	additionalPlugins := []string{
		"ansicolor",
		"greenballs",
	}

	smokeTest := func(password, invalidpassword string, createJob bool, expectedPlugins, nonExpectedPlugins []string) {
		By("loading plugins correctly")
		logs, err := podmancli.ContainerLogs(j.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(logs).NotTo(ContainSubstring("Failed Loading plugin"))

		By("having the right plugins installed")
		code, out, err := podmancli.ContainerExec(j.ID, []string{"ls", "/var/lib/jenkins/plugins"})
		Expect(err).NotTo(HaveOccurred())
		Expect(code).To(Equal(0))
		files := strings.Split(string(out), "\n")

		for _, elem := range expectedPlugins {
			Expect(files).To(ContainElement(elem + ".jpi"))
		}
		for _, elem := range nonExpectedPlugins {
			Expect(files).NotTo(ContainElement(elem + ".jpi"))
		}

		if createJob {
			By("creating a test job")
			resp, err := j.CreateJob("testJob", password, "testdata/testjob.xml")
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
		}

		By("checking the test job exists")
		resp, err := j.GetJob("testJob", password)
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(http.StatusOK))

		By("failing to create a test job with an invalid password")
		resp, err = j.CreateJob("failJob", invalidpassword, "testdata/testjob.xml")
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))

		By("checking the test job doesn't exist")
		resp, err = j.GetJob("failJob", password)
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
	}

	It("should pass a smoke test", func() {
		By("starting Jenkins")
		err := j.Start(imageName, []string{})
		Expect(err).NotTo(HaveOccurred())

		smokeTest("password", "invalidpassword", true, basePlugins, additionalPlugins)

		By("restarting Jenkins with a new password")
		_, err = podmancli.ContainerStopAndRemove(j.ID, 30)
		Expect(err).NotTo(HaveOccurred())

		err = j.Start(imageName, []string{"JENKINS_PASSWORD=newpassword"})
		Expect(err).NotTo(HaveOccurred())

		smokeTest("newpassword", "password", false, basePlugins, additionalPlugins)
	})

	It("should install plugins at startup", func() {
		By("starting Jenkins")
		err := j.Start(imageName, []string{"INSTALL_PLUGINS=ansicolor:0.4.1,greenballs"})
		Expect(err).NotTo(HaveOccurred())

		var expectedPlugins []string
		expectedPlugins = append(expectedPlugins, basePlugins...)
		expectedPlugins = append(expectedPlugins, additionalPlugins...)
		smokeTest("password", "invalidpassword", true, expectedPlugins, nil)
	})

	It("should pass a smoke test after an s2i build", func() {
		s2i, err := exec.LookPath("s2i")
		if err != nil {
			Skip("s2i binary not found")
		}

		By("running s2i build")
		destImage := fmt.Sprintf("jenkins-test-s2i-%d", rand.Intn(1e9))

		By("set up podman debug")
		ctx, cancel := context.WithTimeout(context.Background(), 25*time.Minute)
		cmdstrs := []string{"ps", "-ef"}
		go podmancli.ExecInActiveContainers(GinkgoWriter, ctx, cmdstrs)
		go podmancli.InspectActiveContainers(GinkgoWriter, ctx)

		cmd := exec.Cmd{
			Path: s2i,
			Args: []string{
				s2i,
				"build",
				"--pull-policy=never",
				"--loglevel=5",
				"testdata/s2i",
				imageName,
				destImage,
			},
			Stdout: GinkgoWriter,
			Stderr: GinkgoWriter,
		}
		err = cmd.Run()
		Expect(err).NotTo(HaveOccurred())
		cancel()

		imageNamesToRemove = append(imageNamesToRemove, destImage)

		By("starting Jenkins")
		err = j.Start(destImage, nil)
		Expect(err).NotTo(HaveOccurred())

		var expectedPlugins []string
		expectedPlugins = append(expectedPlugins, basePlugins...)
		expectedPlugins = append(expectedPlugins, additionalPlugins...)
		smokeTest("password", "invalidpassword", true, expectedPlugins, nil)

		By("checking sample-app-test job exists")
		resp, err := j.GetJob("sample-app-test", "password")
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(http.StatusOK))

		By("checking files laid down by s2i exist")
		code, _, err := podmancli.ContainerExec(j.ID, []string{"stat", "/var/lib/jenkins/plugins/sample.jpi.pinned"})
		Expect(err).NotTo(HaveOccurred())
		Expect(code).To(Equal(0))

		code, _, err = podmancli.ContainerExec(j.ID, []string{"stat", "/var/lib/jenkins/jobs/sample-app-test/config.xml"})
		Expect(err).NotTo(HaveOccurred())
		Expect(code).To(Equal(0))

		code, _, err = podmancli.ContainerExec(j.ID, []string{"grep", "-q", "s2i-test-config", "/var/lib/jenkins/config.xml"})
		Expect(err).NotTo(HaveOccurred())
		Expect(code).To(Equal(0))
	})

	It("should handle spaces in command line arguments correctly", func() {
		By("starting Jenkins")
		err := j.Start(imageName, []string{`JENKINS_JAVA_OVERRIDES=-Dcontains\ space -Dnospace`})
		Expect(err).NotTo(HaveOccurred())

		By("checking resolved command line arguments")
		_, bytes, err := podmancli.ContainerExec(j.ID, []string{"find", "/proc", "-name", "cmdline"})
		Expect(err).NotTo(HaveOccurred())
		output := string(bytes)
		lines := strings.Split(output, "\n")
		found := false
		for _, line := range lines {
			_, bytes, err = podmancli.ContainerExec(j.ID, []string{"cat", line})
			if err != nil {
				continue
			}
			cat := string(bytes)
			if strings.Contains(cat, `-Dcontains space`) && strings.Contains(cat, `-Dnospace`) {
				found = true
				break
			}
		}
		Expect(found).To(BeTrue())
	})
})
