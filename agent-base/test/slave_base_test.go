package test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/docker/engine-api/types/container"
	"github.com/openshift/jenkins/pkg/docker"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Base Slave Suite")
}

var dockercli *docker.Client
var imageName string

var _ = BeforeSuite(func() {
	var err error
	dockercli, err = docker.NewEnvClient()
	Expect(err).NotTo(HaveOccurred())

	imageName = os.Getenv("IMAGE_NAME")
	if imageName == "" {
		imageName = "openshift/jenkins-slave-base-centos7-candidate"
	}
})

var _ = Describe("Base slave testing", func() {
	var id string

	AfterEach(func() {
		if CurrentGinkgoTestDescription().Failed {
			By("printing container logs")
			logs, err := dockercli.ContainerLogs(id)
			Expect(err).NotTo(HaveOccurred())
			_, err = GinkgoWriter.Write(logs)
			Expect(err).NotTo(HaveOccurred())
		}

		err := dockercli.ContainerStop(id, nil)
		Expect(err).NotTo(HaveOccurred())

		err = dockercli.ContainerRemove(id)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should contain a runnable oc", func() {
		var err error
		id, err = dockercli.ContainerCreate(
			&container.Config{
				Image: imageName,
				Cmd:   []string{"oc"},
				Tty:   true,
			},
			nil)
		Expect(err).NotTo(HaveOccurred())

		err = dockercli.ContainerStart(id)
		Expect(err).NotTo(HaveOccurred())

		code, err := dockercli.ContainerWait(id)
		Expect(err).NotTo(HaveOccurred())
		Expect(code).To(Equal(0))
	})
})
