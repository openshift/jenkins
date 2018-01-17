package test

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/jenkins/pkg/docker"
	"github.com/openshift/jenkins/pkg/jenkins"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Jenkins Suite (v1)")
}

var dockercli *docker.Client
var imageName string

var _ = BeforeSuite(func() {
	var err error
	dockercli, err = docker.NewEnvClient()
	Expect(err).NotTo(HaveOccurred())

	imageName = os.Getenv("IMAGE_NAME")
	if imageName == "" {
		imageName = "openshift/jenkins-1-centos7-candidate"
	}
})

var _ = Describe("Jenkins v1 testing", func() {
	var j *jenkins.Jenkins
	var imageNamesToRemove []string

	BeforeEach(func() {
		var err error
		j = jenkins.NewJenkins(dockercli)
		j.Volume, err = dockercli.VolumeCreate()
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		if CurrentGinkgoTestDescription().Failed {
			By("printing container logs")
			logs, err := dockercli.ContainerLogs(j.ID)
			Expect(err).NotTo(HaveOccurred())
			_, err = GinkgoWriter.Write(logs)
			Expect(err).NotTo(HaveOccurred())
		}

		err := dockercli.ContainerStopAndRemove(j.ID, nil)
		Expect(err).NotTo(HaveOccurred())

		err = dockercli.VolumeRemove(j.Volume)
		Expect(err).NotTo(HaveOccurred())

		for _, imageName := range imageNamesToRemove {
			err = dockercli.ImageRemove(imageName)
			Expect(err).NotTo(HaveOccurred())
		}
		imageNamesToRemove = nil
	})

	smokeTest := func(password, invalidpassword string, createJob bool) {
		By("loading plugins correctly")
		logs, err := dockercli.ContainerLogs(j.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(logs).NotTo(ContainSubstring("Failed Loading plugin"))

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

	It("should pass a smoke test in 64-bit mode", func() {
		By("starting Jenkins")
		err := j.Start(imageName, []string{"OPENSHIFT_JENKINS_JVM_ARCH=x86_64"})
		Expect(err).NotTo(HaveOccurred())

		smokeTest("password", "invalidpassword", true)

		By("restarting Jenkins with a new password")
		err = dockercli.ContainerStopAndRemove(j.ID, docker.Duration(30*time.Second))
		Expect(err).NotTo(HaveOccurred())

		err = j.Start(imageName, []string{"OPENSHIFT_JENKINS_JVM_ARCH=x86_64", "JENKINS_PASSWORD=newpassword"})
		Expect(err).NotTo(HaveOccurred())

		smokeTest("newpassword", "password", false)
	})

	It("should pass a smoke test in 32-bit mode", func() {
		By("starting Jenkins")
		err := j.Start(imageName, []string{"OPENSHIFT_JENKINS_JVM_ARCH=i386"})
		Expect(err).NotTo(HaveOccurred())

		smokeTest("password", "invalidpassword", true)

		By("restarting Jenkins with a new password")
		err = dockercli.ContainerStopAndRemove(j.ID, docker.Duration(30*time.Second))
		Expect(err).NotTo(HaveOccurred())

		err = j.Start(imageName, []string{"OPENSHIFT_JENKINS_JVM_ARCH=i386", "JENKINS_PASSWORD=newpassword"})
		Expect(err).NotTo(HaveOccurred())

		smokeTest("newpassword", "password", false)
	})

	It("should pass a smoke test after an s2i build", func() {
		s2i, err := exec.LookPath("s2i")
		if err != nil {
			Skip("s2i binary not found")
		}

		By("running s2i build")
		destImage := fmt.Sprintf("jenkins-test-s2i-%d", rand.Intn(1e9))

		cmd := exec.Cmd{
			Path: s2i,
			Args: []string{
				s2i,
				"build",
				"--pull-policy=never",
				"testdata/s2i",
				imageName,
				destImage,
			},
			Stdout: GinkgoWriter,
			Stderr: GinkgoWriter,
		}
		err = cmd.Run()
		Expect(err).NotTo(HaveOccurred())

		imageNamesToRemove = append(imageNamesToRemove, destImage)

		By("starting Jenkins")
		err = j.Start(destImage, nil)
		Expect(err).NotTo(HaveOccurred())

		smokeTest("password", "invalidpassword", true)

		By("checking sample-app-test job exists")
		resp, err := j.GetJob("sample-app-test", "password")
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(http.StatusOK))

		By("checking files laid down by s2i exist")
		code, _, err := dockercli.ContainerExec(j.ID, []string{"stat", "/var/lib/jenkins/plugins/sample.jpi.pinned"})
		Expect(err).NotTo(HaveOccurred())
		Expect(code).To(Equal(0))

		code, _, err = dockercli.ContainerExec(j.ID, []string{"stat", "/var/lib/jenkins/plugins/ansicolor.jpi"})
		Expect(err).NotTo(HaveOccurred())
		Expect(code).To(Equal(0))

		code, _, err = dockercli.ContainerExec(j.ID, []string{"stat", "/var/lib/jenkins/jobs/sample-app-test/config.xml"})
		Expect(err).NotTo(HaveOccurred())
		Expect(code).To(Equal(0))

		code, _, err = dockercli.ContainerExec(j.ID, []string{"grep", "-q", "s2i-test-config", "/var/lib/jenkins/config.xml"})
		Expect(err).NotTo(HaveOccurred())
		Expect(code).To(Equal(0))
	})
})
