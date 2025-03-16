package test

import (
	"os"
	"testing"

	"github.com/containers/podman/v4/pkg/specgen"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/jenkins/pkg/podman"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Base Slave Suite")
}

var podmancli *podman.Client
var imageName string

var _ = BeforeSuite(func() {
	var err error
	podmancli, err = podman.NewEnvClient()
	Expect(err).NotTo(HaveOccurred())

	imageName = os.Getenv("IMAGE_NAME")
	if imageName == "" {
		imageName = "openshift/jenkins-slave-base-rhel9-candidate"
	}
})

var _ = Describe("Base slave testing", func() {
	var id string

	AfterEach(func() {
		if CurrentGinkgoTestDescription().Failed {
			By("printing container logs")
			logs, err := podmancli.ContainerLogs(id)
			Expect(err).NotTo(HaveOccurred())
			_, err = GinkgoWriter.Write(logs)
			Expect(err).NotTo(HaveOccurred())
		}

		err := podmancli.ContainerStop(id, 60)
		Expect(err).NotTo(HaveOccurred())

		_, err = podmancli.ContainerRemove(id)
		Expect(err).NotTo(HaveOccurred())

	})

	It("should contain a runnable oc", func() {
		var err error
		sgen := specgen.NewSpecGenerator(imageName, false)
		sgen.Terminal = true
		sgen.Command = []string{"oc"}
		sgen.Entrypoint = []string{"/bin/bash", "-l", "-c"}
		id, err = podmancli.ContainerCreate(sgen)
		Expect(err).NotTo(HaveOccurred())

		err = podmancli.ContainerStart(id)
		Expect(err).NotTo(HaveOccurred())

		code, err := podmancli.ContainerWait(id)
		Expect(err).NotTo(HaveOccurred())
		Expect(code).To(Equal(int32(0)))
	})
})
