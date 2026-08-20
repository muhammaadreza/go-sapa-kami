package middleware

import (
	"net/http"
	"net/url"
	"strings"
)

func originAllowlist(origins []string) map[string]struct{} {
	allowed := make(map[string]struct{}, len(origins))
	for _, origin := range origins {
		normalized, ok := normalizeOrigin(origin);
		if ok { allowed[normalized] = struct{}{} }
	}

	return allowed
}

func normalizeOrigin(raw string) (string, bool) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.User != nil || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
		return "", false
	}

	return strings.ToLower(u.Scheme) + "://" + strings.ToLower(u.Host), true
}

func requestOrigin(r *http.Request) (string, bool) {
	origin, ok := normalizeOrigin(r.Header.Get("Origin"));
	if ok { return origin, true }

	referer, err := url.Parse(r.Header.Get("Referer"))
	if err != nil || referer.Scheme == "" || referer.Host == "" { return "", false }

	return normalizeOrigin(referer.Scheme + "://" + referer.Host)
}
