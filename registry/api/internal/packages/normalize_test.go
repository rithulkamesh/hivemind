package packages

import (
	"testing"
)

func TestNormalizeName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"My_Package", "my-package"},
		{"my-package", "my-package"},
		{"My.Package", "my-package"},
		{"MY_PACKAGE_NAME", "my-package-name"},
		{"simple", "simple"},
		{"", ""},
		{"_leading", "leading"},
		{"trailing_", "trailing"},
		{"__double__", "double"},
		{"MiXeD.CaSe_PkG", "mixed-case-pkg"},
		{"a.b.c", "a-b-c"},
		{"UPPER", "upper"},
		{"a---b", "a---b"}, // multiple dashes preserved (only leading/trailing stripped)
		{"pkg_v2.0", "pkg-v2-0"},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := NormalizeName(tc.input)
			if got != tc.want {
				t.Errorf("NormalizeName(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}
