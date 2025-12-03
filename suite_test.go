package modbus_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	// "github.com/onsi/gomega/format"
)

func TestRtu(t *testing.T) {
	// format.TruncatedDiff = false
	// format.MaxLength = 50000

	RegisterFailHandler(Fail)
	RunSpecs(t, "Modbus TCP Suite")
}
