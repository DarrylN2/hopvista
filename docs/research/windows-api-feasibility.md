# Ordinary-user Windows probe feasibility

**Subsequent actual Go check (2026-10-10):** toolchain availability was resolved and unelevated Go IPv4/local IPv6 interop was executed. Remote IPv6 remains unavailable; a short native timeout overrun prevents declaring the new gate passed. See [milestone 1 Go evidence](go-windows-network-feasibility.md). The C# results below remain historical supporting evidence.

Evidence for [Validate ordinary-user Windows probe APIs](https://github.com/DarrylN2/hopvista/issues/8), gathered on 2026-10-06. The repeatable, disposable interop check is [windows-api-check.ps1](windows-api-check.ps1). It makes direct Windows IP Helper calls from PowerShell-hosted C# and one HTTPS request from .NET; it is not HopVista application code or a Go binding.

## Environment and method

- Windows reported `Microsoft Windows NT 10.0.26200.0`, 64-bit process, and **not elevated** (`WindowsPrincipal.IsInRole(Administrator) = false`).
- The restricted command sandbox returned error 5 (`ACCESS_DENIED`) from `IcmpCreateFile` / `Icmp6CreateFile`, and even its `ping.exe` could not contact the IP driver. The same check succeeded when run outside that sandbox, still unelevated. The sandbox result is not evidence that Windows normally requires elevation.
- Calls used `IcmpSendEcho2`, `Icmp6SendEcho2`, and `GetBestRoute2` from `iphlpapi.dll`, matching Microsoft's [IPv4 echo](https://learn.microsoft.com/en-us/windows/win32/api/icmpapi/nf-icmpapi-icmpsendecho2), [IPv6 echo](https://learn.microsoft.com/en-us/windows/win32/api/icmpapi/nf-icmpapi-icmp6sendecho2), [reply layout](https://learn.microsoft.com/en-us/windows/win32/api/ipexport/ns-ipexport-icmpv6_echo_reply_lh), and [route API](https://learn.microsoft.com/en-us/windows/win32/api/netioapi/nf-netioapi-getbestroute2) documentation. The 64-bit option and reply layouts were exercised by parsing actual replies, including an IPv6 address, status, and RTT.
- Private source and gateway addresses are omitted below. The public probe destination was `1.1.1.1`; `192.0.2.1` was used solely to observe a no-reply outcome. Each outcome is one probe, not a reliability estimate.

## Observations

| Check | Observed result |
| --- | --- |
| IPv4 localhost echo | One reply from `127.0.0.1`, status `0` (`IP_SUCCESS`). |
| IPv6 localhost echo | One reply from `::1`, status `0`, with both Hop Limit 64 and Hop Limit 1. The returned address parsed as `::1`. |
| IPv4 gateway echo | One reply from the on-link gateway, status `0`. |
| IPv4 transit hop | Echo to `1.1.1.1` with TTL 1 returned one reply from the private first-hop router, status `11013` (`IP_TTL_EXPIRED_TRANSIT`). |
| IPv4 destination | Echo to `1.1.1.1` with TTL 64 returned one destination reply, status `0`. |
| No reply | Echo to `192.0.2.1` returned zero replies and `GetLastError = 11010` (`IP_REQ_TIMED_OUT`). With a 750 ms request timeout, the observed call returned in 180 ms in the final run; an earlier run returned in 387 ms. A timeout is not evidence that HTTP is unreachable. |
| Route selection | `GetBestRoute2` returned an unspecified next hop for localhost and the on-link gateway, and the same private gateway and interface index as `Get-NetRoute` for `1.1.1.1`. It also returned a selected source address. IPv6 localhost returned an unspecified next hop. |
| Pinned HTTPS exchange | A TCP connection to literal `1.1.1.1:443` authenticated TLS with hostname `one.one.one.one`, sent that HTTP Host value, and received `HTTP/1.1 200 OK`. This proves the hostname can remain distinct from the dialed IP in a local client; Go's transport wiring still needs its own implementation check. |

Microsoft defines status `11013` as TTL expired in transit and `11010` as request timed out. The former is direct evidence that TTL-stepped ICMP can reveal an IPv4 responding hop here; the latter says only that no response arrived for that echo. [IPv4 reply statuses](https://learn.microsoft.com/en-us/windows/win32/api/ipexport/ns-ipexport-icmp_echo_reply)

## Limits and planning consequence

- This machine had **no external IPv6 default route**, so IPv6 transit-hop responses could not be checked. IPv6 localhost validates the API call and Hop Limit option, not a remote IPv6 path. No VPN route was available for comparison. A remote IPv6 and VPN check remains an acceptance case on machines that have those routes; their absence here is not a failure of the API.
- The synchronous ICMP functions accept a timeout but no cancellation token. A caller can stop scheduling more probes on cancellation, then wait for an in-flight call to return by its own short timeout. The check observed a no-reply call return within its 750 ms bound, but did not test cancellation of a pending native call. If sub-second cancellation is required, evaluate the APIs' documented asynchronous event path rather than assuming a Go context interrupts the synchronous call. [IPv4 API](https://learn.microsoft.com/en-us/windows/win32/api/icmpapi/nf-icmpapi-icmpsendecho2), [IPv6 API](https://learn.microsoft.com/en-us/windows/win32/api/icmpapi/nf-icmpapi-icmp6sendecho2)
- Go was not installed, so the exact Go-to-Windows binding was not executed. These results support an ordinary-user IP Helper route on this Windows build and do not justify a C++ component or raw-socket elevation. Keep Go interop layout and remote IPv6 as build/acceptance checks, with explicit partial outcomes if a family is unavailable.
