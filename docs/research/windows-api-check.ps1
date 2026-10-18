# Disposable feasibility check for HopVista issue #8. Run unelevated on Windows.
# The script sends only a few ICMP requests and reads the route table.
Add-Type -TypeDefinition @'
using System;
using System.Runtime.InteropServices;
public static class HopVistaIcmpCheck {
    [DllImport("iphlpapi.dll", SetLastError = true)] public static extern IntPtr IcmpCreateFile();
    [DllImport("iphlpapi.dll", SetLastError = true)] public static extern IntPtr Icmp6CreateFile();
    [DllImport("iphlpapi.dll", SetLastError = true)] public static extern bool IcmpCloseHandle(IntPtr handle);
    [DllImport("iphlpapi.dll", SetLastError = true)] public static extern uint IcmpSendEcho2(
        IntPtr handle, IntPtr evt, IntPtr apc, IntPtr context, uint destination,
        byte[] request, ushort requestSize, IntPtr options, IntPtr reply, uint replySize, uint timeout);
    [DllImport("iphlpapi.dll", SetLastError = true)] public static extern uint Icmp6SendEcho2(
        IntPtr handle, IntPtr evt, IntPtr apc, IntPtr context, IntPtr source, IntPtr destination,
        byte[] request, ushort requestSize, IntPtr options, IntPtr reply, uint replySize, uint timeout);
    [DllImport("iphlpapi.dll", SetLastError = true)] public static extern uint GetBestRoute2(
        IntPtr interfaceLuid, uint interfaceIndex, IntPtr source, IntPtr destination,
        uint sortOptions, IntPtr route, IntPtr bestSource);
}
'@

function New-NativeBuffer([byte[]]$Bytes) {
    $ptr = [Runtime.InteropServices.Marshal]::AllocHGlobal($Bytes.Length)
    [Runtime.InteropServices.Marshal]::Copy($Bytes, 0, $ptr, $Bytes.Length)
    return $ptr
}

function New-Sockaddr([string]$Address) {
    $ip = [Net.IPAddress]::Parse($Address)
    $bytes = [byte[]]::new(28)
    if ($ip.AddressFamily -eq [Net.Sockets.AddressFamily]::InterNetwork) {
        $bytes[0] = 2
        [Array]::Copy($ip.GetAddressBytes(), 0, $bytes, 4, 4)
    } else {
        $bytes[0] = 23
        [Array]::Copy($ip.GetAddressBytes(), 0, $bytes, 8, 16)
        [Array]::Copy([BitConverter]::GetBytes([uint32]$ip.ScopeId), 0, $bytes, 24, 4)
    }
    return New-NativeBuffer $bytes
}

function Invoke-Icmp([string]$Address, [byte]$Ttl, [uint32]$Timeout = 750) {
    $ip = [Net.IPAddress]::Parse($Address)
    $v6 = $ip.AddressFamily -eq [Net.Sockets.AddressFamily]::InterNetworkV6
    $handle = if ($v6) { [HopVistaIcmpCheck]::Icmp6CreateFile() } else { [HopVistaIcmpCheck]::IcmpCreateFile() }
    if ($handle -eq [IntPtr]::Zero -or $handle -eq [IntPtr]::new(-1)) {
        return [pscustomobject]@{ Address=$Address; Ttl=$Ttl; Error="create: $([Runtime.InteropServices.Marshal]::GetLastWin32Error())" }
    }
    $request = [Text.Encoding]::ASCII.GetBytes('hv-check')
    $reply = [Runtime.InteropServices.Marshal]::AllocHGlobal(512)
    $options = New-NativeBuffer ([byte[]]@($Ttl, 0, 0, 0, 0, 0, 0, 0))
    $source = [IntPtr]::Zero
    $destination = [IntPtr]::Zero
    $clock = [Diagnostics.Stopwatch]::StartNew()
    try {
        if ($v6) {
            $source = New-Sockaddr '::'
            $destination = New-Sockaddr $Address
            $count = [HopVistaIcmpCheck]::Icmp6SendEcho2($handle, [IntPtr]::Zero, [IntPtr]::Zero,
                [IntPtr]::Zero, $source, $destination, $request, [uint16]$request.Length,
                $options, $reply, 512, $Timeout)
            $statusOffset = 28
            $rttOffset = 32
            # IPV6_ADDRESS_EX stores its 16 address bytes after port (2) and flowinfo (4).
            $addressOffset = 6
            $addressLength = 16
        } else {
            $destinationValue = [BitConverter]::ToUInt32($ip.GetAddressBytes(), 0)
            $count = [HopVistaIcmpCheck]::IcmpSendEcho2($handle, [IntPtr]::Zero, [IntPtr]::Zero,
                [IntPtr]::Zero, $destinationValue, $request, [uint16]$request.Length,
                $options, $reply, 512, $Timeout)
            $statusOffset = 4
            $rttOffset = 8
            $addressOffset = 0
            $addressLength = 4
        }
        $errorCode = if ($count -eq 0) { [Runtime.InteropServices.Marshal]::GetLastWin32Error() } else { 0 }
        $status = if ($count -gt 0) { [Runtime.InteropServices.Marshal]::ReadInt32($reply, $statusOffset) } else { $null }
        $rtt = if ($count -gt 0) { [Runtime.InteropServices.Marshal]::ReadInt32($reply, $rttOffset) } else { $null }
        $replyAddress = $null
        if ($count -gt 0) {
            $raw = [byte[]]::new($addressLength)
            [Runtime.InteropServices.Marshal]::Copy([IntPtr]::Add($reply, $addressOffset), $raw, 0, $addressLength)
            $replyAddress = [Net.IPAddress]::new($raw).ToString()
        }
        $clock.Stop()
        return [pscustomobject]@{ Address=$Address; Ttl=$Ttl; Count=$count; Status=$status; RttMs=$rtt; ReplyFrom=$replyAddress; LastError=$errorCode; ElapsedMs=$clock.ElapsedMilliseconds }
    } finally {
        if ($source -ne [IntPtr]::Zero) { [Runtime.InteropServices.Marshal]::FreeHGlobal($source) }
        if ($destination -ne [IntPtr]::Zero) { [Runtime.InteropServices.Marshal]::FreeHGlobal($destination) }
        [Runtime.InteropServices.Marshal]::FreeHGlobal($options)
        [Runtime.InteropServices.Marshal]::FreeHGlobal($reply)
        [void][HopVistaIcmpCheck]::IcmpCloseHandle($handle)
    }
}

function Invoke-BestRoute([string]$Address) {
    $destination = New-Sockaddr $Address
    $route = New-NativeBuffer ([byte[]]::new(512))
    $source = New-NativeBuffer ([byte[]]::new(28))
    try {
        $code = [HopVistaIcmpCheck]::GetBestRoute2([IntPtr]::Zero, 0, [IntPtr]::Zero, $destination, 0, $route, $source)
        if ($code -ne 0) { return [pscustomobject]@{ Address=$Address; ReturnCode=$code } }
        $ifIndex = [Runtime.InteropServices.Marshal]::ReadInt32($route, 8)
        # MIB_IPFORWARD_ROW2: NET_LUID(8), index(4), IP_ADDRESS_PREFIX(32), NextHop(28).
        $nextHop = [IntPtr]::Add($route, 44)
        $family = [Runtime.InteropServices.Marshal]::ReadInt16($nextHop, 0)
        $count = if ($family -eq 2) { 4 } elseif ($family -eq 23) { 16 } else { 0 }
        $offset = if ($family -eq 2) { 4 } else { 8 }
        $next = $null
        if ($count -gt 0) {
            $raw = [byte[]]::new($count)
            [Runtime.InteropServices.Marshal]::Copy([IntPtr]::Add($nextHop, $offset), $raw, 0, $count)
            $next = [Net.IPAddress]::new($raw).ToString()
        }
        $sourceFamily = [Runtime.InteropServices.Marshal]::ReadInt16($source, 0)
        $sourceCount = if ($sourceFamily -eq 2) { 4 } elseif ($sourceFamily -eq 23) { 16 } else { 0 }
        $sourceAddress = $null
        if ($sourceCount -gt 0) {
            $sourceRaw = [byte[]]::new($sourceCount)
            $sourceOffset = if ($sourceFamily -eq 2) { 4 } else { 8 }
            [Runtime.InteropServices.Marshal]::Copy([IntPtr]::Add($source, $sourceOffset), $sourceRaw, 0, $sourceCount)
            $sourceAddress = [Net.IPAddress]::new($sourceRaw).ToString()
        }
        return [pscustomobject]@{ Address=$Address; ReturnCode=$code; InterfaceIndex=$ifIndex; Source=$sourceAddress; NextHop=$next }
    } finally {
        [Runtime.InteropServices.Marshal]::FreeHGlobal($destination)
        [Runtime.InteropServices.Marshal]::FreeHGlobal($route)
        [Runtime.InteropServices.Marshal]::FreeHGlobal($source)
    }
}

$principal = [Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()
"Windows=$([Environment]::OSVersion.VersionString) Elevated=$($principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator))"
$defaultRoute = Get-NetRoute -AddressFamily IPv4 -ErrorAction SilentlyContinue |
    Where-Object { $_.DestinationPrefix -eq '0.0.0.0/0' -and $_.NextHop -ne '0.0.0.0' } |
    Sort-Object RouteMetric,InterfaceMetric | Select-Object -First 1
$gateway = $defaultRoute.NextHop
Invoke-Icmp '127.0.0.1' 64
Invoke-Icmp '::1' 64
Invoke-Icmp '::1' 1
if ($gateway) { Invoke-Icmp $gateway 64 }
Invoke-Icmp '1.1.1.1' 1
Invoke-Icmp '1.1.1.1' 64
Invoke-Icmp '192.0.2.1' 64
Invoke-BestRoute '127.0.0.1'
if ($gateway) { Invoke-BestRoute $gateway }
Invoke-BestRoute '1.1.1.1'
Invoke-BestRoute '::1'

# Literal-IP TCP dial with hostname-based TLS authentication and HTTP Host.
$tcp = [Net.Sockets.TcpClient]::new()
try {
    $connect = $tcp.ConnectAsync([Net.IPAddress]::Parse('1.1.1.1'), 443)
    if (-not $connect.Wait(2000)) { throw 'TCP connect exceeded 2 seconds' }
    $tls = [Net.Security.SslStream]::new($tcp.GetStream(), $false)
    try {
        $tls.ReadTimeout = 3000
        $tls.WriteTimeout = 3000
        $tls.AuthenticateAsClient('one.one.one.one')
        $writer = [IO.StreamWriter]::new($tls, [Text.Encoding]::ASCII, 1024, $true)
        $writer.NewLine = "`r`n"
        $writer.Write("HEAD / HTTP/1.1`r`nHost: one.one.one.one`r`nConnection: close`r`n`r`n")
        $writer.Flush()
        $reader = [IO.StreamReader]::new($tls, [Text.Encoding]::ASCII, $false, 1024, $true)
        [pscustomobject]@{ Dialed='1.1.1.1:443'; TlsName='one.one.one.one'; TlsAuthenticated=$tls.IsAuthenticated; HttpHost='one.one.one.one'; StatusLine=$reader.ReadLine() }
    } finally { $tls.Dispose() }
} catch {
    [pscustomobject]@{ Dialed='1.1.1.1:443'; TlsName='one.one.one.one'; Error=$_.Exception.Message }
} finally { $tcp.Dispose() }
