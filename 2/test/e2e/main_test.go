package e2e

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"testing"
)

func TestMain(m *testing.M) {
	var err error
	kubeConfig, err = getConfig()
	if err != nil {
		fmt.Printf("failed to get kubeconfig: %v\n", err)
		os.Exit(1)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\ninterrupt received, cleaning up test namespaces...")
		cleanupNamespaces()
		os.Exit(1)
	}()

	exitCode := m.Run()
	cleanupNamespaces()
	os.Exit(exitCode)
}
