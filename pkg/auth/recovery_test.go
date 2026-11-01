package auth

import "testing"

func TestPasswordVersion(t *testing.T) {
	a, b := passwordVersion("hash-one"), passwordVersion("hash-two")
	if a == b {
		t.Fatal("different passwords must give different versions")
	}
	if a != passwordVersion("hash-one") {
		t.Fatal("the same password must give the same version")
	}
	if len(a) != 16 {
		t.Fatalf("version %q should be 16 hex characters", a)
	}
}
