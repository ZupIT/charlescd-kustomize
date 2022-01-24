package kustomize_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestKustomize(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Kustomize Suite")
}
