package password_test

import (
	"testing"

	"github.com/rakshithgoud453/myvault/internal/password"
)

func TestGeneratePasswordLength(t *testing.T) {
	lengths := []int{8, 16, 24, 32, 64}

	for _, l := range lengths {
		pw, err := password.Generate(l)
		if err != nil {
			t.Fatalf("Generate(%d) failed: %v", l, err)
		}
		if len(pw) != l {
			t.Errorf("expected length %d, got %d", l, len(pw))
		}
	}
}

func TestGeneratePasswordRandomness(t *testing.T) {
	p1, err1 := password.Generate(20)
	p2, err2 := password.Generate(20)

	if err1 != nil || err2 != nil {
		t.Fatalf("Generate failed: %v, %v", err1, err2)
	}

	if p1 == p2 {
		t.Errorf("two consecutive password generations produced identical results: %s", p1)
	}
}
