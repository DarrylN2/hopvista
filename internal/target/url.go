// Package target validates target URLs and selects one IPv4 address for each
// hostname in a diagnostic run. It performs no HTTP or ICMP connection.
package target

import (
	"net/netip"
	"net/url"
	"strconv"
	"strings"
)

type Problem struct {
	Code    string
	Message string
}

func (p *Problem) Error() string { return p.Message }

type URL struct {
	value    *url.URL
	hostname string
	hostKey  string
	port     int
	literal  netip.Addr
}

func (u URL) String() string   { return u.value.String() }
func (u URL) Hostname() string { return u.hostname }
func (u URL) Port() int        { return u.port }

func Parse(raw string) (URL, *Problem) {
	parsed, err := url.Parse(raw)
	if err != nil || parsed == nil || !parsed.IsAbs() || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Opaque != "" || parsed.Host == "" || parsed.Hostname() == "" || parsed.User != nil {
		return URL{}, &Problem{Code: "invalid_target_url", Message: "Enter an absolute HTTP or HTTPS URL without credentials."}
	}
	if strings.HasSuffix(parsed.Host, ":") {
		return URL{}, &Problem{Code: "invalid_target_url", Message: "The target URL has an invalid port."}
	}
	port := 80
	if parsed.Scheme == "https" {
		port = 443
	}
	if rawPort := parsed.Port(); rawPort != "" {
		port, err = strconv.Atoi(rawPort)
		if err != nil || port < 1 || port > 65535 {
			return URL{}, &Problem{Code: "invalid_target_url", Message: "The target URL has an invalid port."}
		}
	}
	host := parsed.Hostname()
	if strings.ContainsAny(host, "\r\n\t ") {
		return URL{}, &Problem{Code: "invalid_target_url", Message: "The target URL has an invalid host."}
	}
	var literal netip.Addr
	if ip, err := netip.ParseAddr(host); err == nil {
		if ip.Is6() {
			return URL{}, &Problem{Code: "ipv6_not_supported_in_v1", Message: "IPv6 targets are not supported in v1."}
		}
		if !UsableIPv4(ip) {
			return URL{}, &Problem{Code: "invalid_target_url", Message: "This IPv4 destination cannot be probed."}
		}
		literal = ip
	} else if strings.HasPrefix(parsed.Host, "[") || strings.Contains(host, ":") || !validHostname(host) {
		return URL{}, &Problem{Code: "invalid_target_url", Message: "The target URL has an invalid host."}
	}
	parsed.Fragment, parsed.RawFragment = "", ""
	return URL{value: parsed, hostname: host, hostKey: strings.TrimSuffix(strings.ToLower(host), "."), port: port, literal: literal}, nil
}

func validHostname(host string) bool {
	if len(host) == 0 || len(host) > 253 {
		return false
	}
	for _, label := range strings.Split(strings.TrimSuffix(host, "."), ".") {
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, c := range label {
			if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-') {
				return false
			}
		}
	}
	return true
}

func resolveRedirect(from URL, location string) (URL, *Problem) {
	if location == "" {
		return URL{}, &Problem{Code: "invalid_redirect", Message: "The redirect has no Location."}
	}
	ref, err := url.Parse(location)
	if err != nil {
		return URL{}, &Problem{Code: "invalid_redirect", Message: "The redirect Location is invalid."}
	}
	next := from.value.ResolveReference(ref)
	parsed, problem := Parse(next.String())
	if problem != nil {
		if problem.Code == "invalid_target_url" {
			problem.Code = "invalid_redirect"
		}
		return URL{}, problem
	}
	return parsed, nil
}
