package openwa

import "net/url"

func validProxy(raw string, kind ProxyType) bool {
	if raw == "" {
		return kind == ""
	}
	u, err := url.Parse(raw)
	if err != nil || u.Hostname() == "" || u.RawQuery != "" || u.Fragment != "" || containsControl(raw) {
		return false
	}
	switch ProxyType(u.Scheme) {
	case ProxyTypeHTTP, ProxyTypeHTTPS, ProxyTypeSOCKS4, ProxyTypeSOCKS5:
	default:
		return false
	}
	return kind == "" || kind == ProxyType(u.Scheme)
}
