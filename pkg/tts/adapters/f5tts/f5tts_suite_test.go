package f5tts_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestF5tts(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "F5tts Suite")
}
