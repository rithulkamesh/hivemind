package packages

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEscapeHTML(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "hello"},
		{"<script>alert('xss')</script>", "&lt;script&gt;alert('xss')&lt;/script&gt;"},
		{`"quoted"`, "&quot;quoted&quot;"},
		{"a & b", "a &amp; b"},
		{"<b>bold</b>", "&lt;b&gt;bold&lt;/b&gt;"},
		{"", ""},
		{"no special chars", "no special chars"},
		{`<a href="http://example.com?a=1&b=2">link</a>`,
			`&lt;a href=&quot;http://example.com?a=1&amp;b=2&quot;&gt;link&lt;/a&gt;`},
		// Multiple ampersands
		{"&&&", "&amp;&amp;&amp;"},
		// Mixed special chars
		{`<"&>`, `&lt;&quot;&amp;&gt;`},
	}
	for _, tc := range tests {
		got := escapeHTML(tc.input)
		if got != tc.want {
			t.Errorf("escapeHTML(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestWantsJSON(t *testing.T) {
	tests := []struct {
		accept string
		want   bool
	}{
		{"application/vnd.pypi.simple.v1+json", true},
		{"application/json", true},
		{"text/html", false},
		{"", false},
		{"application/vnd.pypi.simple.v1+json, text/html", true},
	}
	for _, tc := range tests {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		if tc.accept != "" {
			req.Header.Set("Accept", tc.accept)
		}
		got := wantsJSON(req)
		if got != tc.want {
			t.Errorf("wantsJSON(Accept: %q) = %v, want %v", tc.accept, got, tc.want)
		}
	}
}
