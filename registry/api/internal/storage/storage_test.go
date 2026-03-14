package storage

import (
	"testing"
)

func TestKey_WithNamespace(t *testing.T) {
	got := Key("myorg", "mypackage", "1.0.0", "mypackage-1.0.0-py3-none-any.whl")
	want := "myorg/mypackage/1.0.0/mypackage-1.0.0-py3-none-any.whl"
	if got != want {
		t.Errorf("Key with namespace = %q, want %q", got, want)
	}
}

func TestKey_WithoutNamespace(t *testing.T) {
	got := Key("", "mypackage", "1.0.0", "mypackage-1.0.0.tar.gz")
	want := "mypackage/1.0.0/mypackage-1.0.0.tar.gz"
	if got != want {
		t.Errorf("Key without namespace = %q, want %q", got, want)
	}
}

func TestKey_SpecialCharacters(t *testing.T) {
	got := Key("", "my-pkg", "2.0.0-rc1", "my_pkg-2.0.0rc1-py3-none-any.whl")
	want := "my-pkg/2.0.0-rc1/my_pkg-2.0.0rc1-py3-none-any.whl"
	if got != want {
		t.Errorf("Key = %q, want %q", got, want)
	}
}
