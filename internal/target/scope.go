package target

import "net/netip"

// Scope uses IANA's IPv4 special-purpose registry, last reviewed 2026-10-14.
// Special-use addresses are never silently promoted to public merely because
// netip.Addr.IsGlobalUnicast reports true. A new-host redirect proceeds only
// for Public; an explicit target may select any usable unicast scope.
type Scope string

const (
	Public     Scope = "public"
	Private    Scope = "private"
	Loopback   Scope = "loopback"
	LinkLocal  Scope = "link_local"
	SpecialUse Scope = "special_use"
)

var specialPrefixes = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),
	netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("192.0.0.0/24"),
	netip.MustParsePrefix("192.0.2.0/24"),
	netip.MustParsePrefix("192.31.196.0/24"),
	netip.MustParsePrefix("192.52.193.0/24"),
	netip.MustParsePrefix("192.88.99.0/24"),
	netip.MustParsePrefix("192.175.48.0/24"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("198.51.100.0/24"),
	netip.MustParsePrefix("203.0.113.0/24"),
	netip.MustParsePrefix("240.0.0.0/4"),
}

func Classify(ip netip.Addr) Scope {
	if !UsableIPv4(ip) {
		return SpecialUse
	}
	if ip.IsPrivate() {
		return Private
	}
	if ip.IsLoopback() {
		return Loopback
	}
	if ip.IsLinkLocalUnicast() {
		return LinkLocal
	}
	for _, prefix := range specialPrefixes {
		if prefix.Contains(ip) {
			return SpecialUse
		}
	}
	return Public
}

func UsableIPv4(ip netip.Addr) bool {
	if !ip.Is4() || ip.IsUnspecified() || ip.IsMulticast() || ip == netip.MustParseAddr("255.255.255.255") {
		return false
	}
	// These registry ranges have no normal destination semantics.
	if netip.MustParsePrefix("0.0.0.0/8").Contains(ip) || netip.MustParsePrefix("240.0.0.0/4").Contains(ip) {
		return false
	}
	return true
}
