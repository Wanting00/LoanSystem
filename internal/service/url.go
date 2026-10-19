package service

import (
	"net/url"
	"path"
	"strings"
)

func parseHTTPURL(value string) (*url.URL, bool) {
	parsedURL, err := url.Parse(value)
	if err != nil ||
		!parsedURL.IsAbs() ||
		(!strings.EqualFold(parsedURL.Scheme, "http") && !strings.EqualFold(parsedURL.Scheme, "https")) ||
		parsedURL.Host == "" ||
		parsedURL.Hostname() == "" ||
		parsedURL.User != nil {
		return nil, false
	}

	return parsedURL, true
}

func isPDFOrJPEGURL(value string) bool {
	parsedURL, valid := parseHTTPURL(value)
	if !valid {
		return false
	}

	switch strings.ToLower(path.Ext(parsedURL.Path)) {
	case ".pdf", ".jpg", ".jpeg":
		return true
	default:
		return false
	}
}
