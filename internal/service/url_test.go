package service

import "testing"

func TestIsPDFOrJPEGURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{name: "PDF", url: "https://example.com/agreement.pdf", want: true},
		{name: "JPEG", url: "https://example.com/photo.jpeg", want: true},
		{name: "JPG extension is case insensitive", url: "http://example.com/photo.JPG?download=1", want: true},
		{name: "unsupported extension", url: "https://example.com/agreement.png", want: false},
		{name: "extension is not at end of path", url: "https://example.com/file.pdf/other", want: false},
		{name: "relative URL", url: "/agreement.pdf", want: false},
		{name: "missing host", url: "https:///agreement.pdf", want: false},
		{name: "unsupported scheme", url: "ftp://example.com/agreement.pdf", want: false},
		{name: "userinfo is rejected", url: "https://user:password@example.com/agreement.pdf", want: false},
		{name: "empty URL", url: "", want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := isPDFOrJPEGURL(test.url); got != test.want {
				t.Errorf("isPDFOrJPEGURL(%q) = %t, want %t", test.url, got, test.want)
			}
		})
	}
}
