package safetyCheck_test

import (
	"testing"

	safetyCheck "github.com/henrylee2cn/tp-ext/filter-safetyCheck"
)

func TestSafetyCheck(t *testing.T) {
	safetyCheck := safetyCheck.NewSafetyCheck()
	input := []byte("safetyCheck")
	b, err := safetyCheck.OnPack(input)
	if err != nil {
		t.Fatalf("Onpack: %v", err)
	}

	// Tamper with data
	b = append(b, "viruses"...)

	output, err := safetyCheck.OnUnpack(b)
	if err != nil {
		t.Fatalf("Safety check failed: %v", err)
	}

	t.Logf("Safety check success: want \"%s\", have \"%s\"", string(input), string(output))
}
