# HopVista

HopVista helps a person understand connection delays between their machine and a website.

## Language

**Target URL**:
The website address a person submits for diagnosis, including its scheme, host, port, path, and query when present.
_Avoid_: Destination, service address

**Diagnostic run**:
A bounded collection of measurements taken from the person's machine for one target URL, including separate IPv4 and IPv6 results when both are available.
_Avoid_: Scan, session, test

**Observed finding**:
A statement about one diagnostic run that is directly supported by its recorded measurements and identifies the relevant target URL or redirect exchange and address family.
_Avoid_: Root cause, verdict

**Run result**:
The measurements, outcomes, and observed findings collected for one diagnostic run, including any partial results.
_Avoid_: Report, session data
