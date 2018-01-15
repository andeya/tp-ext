package md5Hash_test

import (
	"testing"

	md5Hash "github.com/henrylee2cn/tp-ext/xfer-md5Hash"
)

func TestMd5Hash(t *testing.T) {
	md5Check := md5Hash.New()
	input := []byte("md5")
	b, err := md5Check.OnPack(input)
	if err != nil {
		t.Fatalf("Onpack: %v", err)
	}

	// Tamper with data
	// b = append(b, "viruses"...)

	output, err := md5Check.OnUnpack(b)
	if err != nil {
		t.Fatalf("Md5 check failed: %v", err)
	}

	t.Logf("Md5 check success: want \"%s\", have \"%s\"", string(input), string(output))
}
