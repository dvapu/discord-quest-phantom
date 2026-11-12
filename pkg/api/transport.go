package api

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

type regionInfoKey struct{}

type localeOverrideKey struct{}

// WithRegionInfo attaches a region to a request context. Every HTTP call made
// with the derived context is sent through that region's proxy (when one is
// configured) and carries matching locale/timezone headers.
func WithRegionInfo(ctx context.Context, info RegionInfo) context.Context {
	return context.WithValue(ctx, regionInfoKey{}, info)
}

// RegionInfoFromContext extracts the region attached with WithRegionInfo.
func RegionInfoFromContext(ctx context.Context) (RegionInfo, bool) {
	if ctx == nil {
		return RegionInfo{}, false
	}
	info, ok := ctx.Value(regionInfoKey{}).(RegionInfo)
	return info, ok
}

// withLocaleOverride pins the locale/timezone for a single request context.
func withLocaleOverride(ctx context.Context, locale, timezone string) context.Context {
	return context.WithValue(ctx, localeOverrideKey{}, RegionInfo{Locale: locale, Timezone: timezone})
}

func localeOverrideFromContext(ctx context.Context) (RegionInfo, bool) {
	if ctx == nil {
		return RegionInfo{}, false
	}
	info, ok := ctx.Value(localeOverrideKey{}).(RegionInfo)
	return info, ok
}

// RegionTransport routes requests to the proxy that belongs to the region in
// the request context. Requests without a region use the direct connection.
type RegionTransport struct {
	mu      sync.RWMutex
	direct  http.RoundTripper
	proxies map[string]http.RoundTripper
}

// NewRegionTransport builds a transport with a direct fallback.
func NewRegionTransport() *RegionTransport {
	return &RegionTransport{
		direct:  cloneDefaultTransport(),
		proxies: make(map[string]http.RoundTripper),
	}
}

func cloneDefaultTransport() *http.Transport {
	base, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return &http.Transport{}
	}
	return base.Clone()
}

// SetProxy registers (or replaces) the HTTP/SOCKS5 proxy for a region code.
func (t *RegionTransport) SetProxy(code, proxyURL string) error {
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" {
		return fmt.Errorf("region proxy code must not be empty")
	}
	parsed, err := url.Parse(proxyURL)
	if err != nil {
		return fmt.Errorf("invalid proxy URL: %w", err)
	}
	switch strings.ToLower(parsed.Scheme) {
	case "http", "https", "socks5", "socks5h":
	default:
		return fmt.Errorf("unsupported proxy scheme %q (use http, https, socks5 or socks5h)", parsed.Scheme)
	}
	if parsed.Host == "" {
		return fmt.Errorf("proxy URL has no host")
	}
	transport := cloneDefaultTransport()
	transport.Proxy = http.ProxyURL(parsed)
	transport.MaxIdleConnsPerHost = 4
	t.mu.Lock()
	if t.proxies == nil {
		t.proxies = make(map[string]http.RoundTripper)
	}
	previous := t.proxies[code]
	t.proxies[code] = transport
	t.mu.Unlock()
	if closer, ok := previous.(interface{ CloseIdleConnections() }); ok {
		closer.CloseIdleConnections()
	}
	return nil
}

// RemoveProxy drops the proxy registration for one region code.
func (t *RegionTransport) RemoveProxy(code string) {
	t.mu.Lock()
	delete(t.proxies, strings.ToUpper(strings.TrimSpace(code)))
	t.mu.Unlock()
}

// HasProxy reports whether a proxy is registered for a region code.
func (t *RegionTransport) HasProxy(code string) bool {
	t.mu.RLock()
	_, ok := t.proxies[strings.ToUpper(strings.TrimSpace(code))]
	t.mu.RUnlock()
	return ok
}

// RoundTrip implements http.RoundTripper.
func (t *RegionTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if info, ok := RegionInfoFromContext(req.Context()); ok {
		t.mu.RLock()
		transport := t.proxies[strings.ToUpper(strings.TrimSpace(info.Code))]
		t.mu.RUnlock()
		if transport != nil {
			return transport.RoundTrip(req)
		}
	}
	return t.direct.RoundTrip(req)
}

// CloseIdleConnections releases keep-alive connections held by every route.
func (t *RegionTransport) CloseIdleConnections() {
	t.mu.RLock()
	direct := t.direct
	proxies := make([]http.RoundTripper, 0, len(t.proxies))
	for _, transport := range t.proxies {
		proxies = append(proxies, transport)
	}
	t.mu.RUnlock()
	if closer, ok := direct.(interface{ CloseIdleConnections() }); ok {
		closer.CloseIdleConnections()
	}
	for _, transport := range proxies {
		if closer, ok := transport.(interface{ CloseIdleConnections() }); ok {
			closer.CloseIdleConnections()
		}
	}
}
