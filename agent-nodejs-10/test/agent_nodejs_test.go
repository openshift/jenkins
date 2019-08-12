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
	RunSpecs(t, "NodeJS Agent Suite")
}

var dockercli *docker.Client
var imageName string

var _ = BeforeSuite(func() {
	var err error
	dockercli, err = docker.NewEnvClient()
	Expect(err).NotTo(HaveOccurred())

	imageName = os.Getenv("IMAGE_NAME")
	if imageName == "" {
		imageName = "openshift/jenkins-agent-nodejs-10-centos7-candidate"
	}
})

var _ = Describe("NodeJS agent testing", func() {
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

	It("should contain a runnable npm", func() {
		// the default entrypoint for the image assumes if you supply more than
		// one arg, you are trying to invoke the slave logic, so we have to
		// override the entrypoint to run complicated commands for testing.

		var err error
		id, err = dockercli.ContainerCreate(
			&container.Config{
				Image:      imageName,
				Entrypoint: []string{"/bin/bash", "-c", "npm --version"},
				Tty:        true,
			},
			nil)
		Expect(err).NotTo(HaveOccurred())

		err = dockercli.ContainerStart(id)
		Expect(err).NotTo(HaveOccurred())

		code, err := dockercli.ContainerWait(id)
		Expect(err).NotTo(HaveOccurred())
		Expect(code).To(Equal(0))
	})

	It("should contain a runnable node", func() {
		// the default entrypoint for the image assumes if you supply more than
		// one arg, you are trying to invoke the slave logic, so we have to
		// override the entrypoint to run complicated commands for testing.

		var err error
		id, err = dockercli.ContainerCreate(
			&container.Config{
				Image:      imageName,
				Entrypoint: []string{"/bin/bash", "-c", "node --version"},
				Tty:        true,
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
