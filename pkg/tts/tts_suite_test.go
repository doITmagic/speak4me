package tts_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTts(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Tts Suite")
}
