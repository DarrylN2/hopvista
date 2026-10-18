"""One-shot WSL IPv4 ICMP host-unreachable fixture for the Go harness.

Run as root in an isolated WSL instance. The original echo setting is restored
after one matching request or a 120-second timeout. No packages are required.
"""

import pathlib
import socket
import struct
import sys
import time


def checksum(data: bytes) -> int:
    if len(data) & 1:
        data += b"\0"
    words = struct.unpack(f"!{len(data) // 2}H", data)
    total = sum(words)
    total = (total & 0xFFFF) + (total >> 16)
    total = (total & 0xFFFF) + (total >> 16)
    return (~total) & 0xFFFF


def main() -> None:
    target = socket.inet_aton(sys.argv[1])
    setting = pathlib.Path("/proc/sys/net/ipv4/icmp_echo_ignore_all")
    original = setting.read_text().strip()
    if original != "0":
        raise RuntimeError("expected ordinary echo replies before fixture")
    with socket.socket(socket.AF_INET, socket.SOCK_RAW, socket.IPPROTO_ICMP) as sock:
        sock.settimeout(1)
        setting.write_text("1\n")
        try:
            print("READY", flush=True)
            deadline = time.monotonic() + 120
            while time.monotonic() < deadline:
                try:
                    packet, _ = sock.recvfrom(65535)
                except socket.timeout:
                    continue
                if len(packet) < 36 or packet[0] >> 4 != 4 or packet[9] != 1:
                    continue
                header_size = (packet[0] & 15) * 4
                if header_size < 20 or len(packet) < header_size + 16:
                    continue
                if packet[16:20] != target or packet[header_size] != 8:
                    continue
                if packet[header_size + 8 : header_size + 16] != b"hv-check":
                    continue
                # RFC 792 error body: type/code, four unused bytes, original
                # IPv4 header and the first eight bytes of its ICMP payload.
                body = b"\x03\x01\0\0\0\0\0\0" + packet[: header_size + 8]
                value = checksum(body)
                body = body[:2] + struct.pack("!H", value) + body[4:]
                source = socket.inet_ntoa(packet[12:16])
                sock.sendto(body, (source, 0))
                print("UNREACHABLE_SENT", flush=True)
                return
            raise TimeoutError("no matching Go echo request received")
        finally:
            setting.write_text(original + "\n")
            print("ECHO_SETTING_RESTORED", flush=True)


if __name__ == "__main__":
    main()
