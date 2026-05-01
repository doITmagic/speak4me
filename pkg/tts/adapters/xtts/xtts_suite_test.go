package xtts_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestXtts(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Xtts Suite")
}
