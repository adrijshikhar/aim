package session

import (
	"encoding/hex"
	"strings"
	"testing"
)

func TestGenerateUUID(t *testing.T) {
	id, err := GenerateUUID()
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(id, "-")
	if len(parts) != 5 {
		t.Fatalf("invalid UUID: %q", id)
	}
	for i, size := range []int{8, 4, 4, 4, 12} {
		if len(parts[i]) != size {
			t.Fatalf("invalid UUID group length: %q", id)
		}
	}
	bytes, err := hex.DecodeString(strings.Join(parts, ""))
	if err != nil {
		t.Fatal(err)
	}
	if bytes[6]>>4 != 4 || bytes[8]>>6 != 2 {
		t.Fatalf("incorrect UUID version or variant: %q", id)
	}
}
