# Independent diagnostic comparison only; C# does not satisfy the Go gate.
param([Parameter(Mandatory)][string]$OutputDirectory)
$ErrorActionPreference = 'Stop'
$principal = [Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()
if ($principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) { throw 'Run unelevated' }
Add-Type -TypeDefinition @'
using System;
using System.Diagnostics;
using System.Net;
using System.Runtime.InteropServices;
public static class NativeTimeoutComparison {
    [DllImport("iphlpapi.dll", SetLastError=true)] static extern IntPtr IcmpCreateFile();
    [DllImport("iphlpapi.dll", SetLastError=true)] static extern bool IcmpCloseHandle(IntPtr handle);
    [DllImport("iphlpapi.dll", SetLastError=true)] static extern uint IcmpSendEcho2(
        IntPtr handle, IntPtr evt, IntPtr apc, IntPtr context, uint destination,
        byte[] request, ushort size, IntPtr options, IntPtr reply, uint replySize, uint timeout);
    public static object Probe(string destination, uint timeout) {
        var handle = IcmpCreateFile();
        if (handle == new IntPtr(-1) || handle == IntPtr.Zero) throw new Exception("create: " + Marshal.GetLastWin32Error());
        var reply = Marshal.AllocHGlobal(512);
        var request = new byte[] {104,118,45,99,104,101,99,107};
        try {
            var clock = Stopwatch.StartNew();
            var count = IcmpSendEcho2(handle, IntPtr.Zero, IntPtr.Zero, IntPtr.Zero,
                BitConverter.ToUInt32(IPAddress.Parse(destination).GetAddressBytes(),0),
                request, (ushort)request.Length, IntPtr.Zero, reply, 512, timeout);
            var error = count == 0 ? Marshal.GetLastWin32Error() : 0;
            clock.Stop();
            return new { destination, timeout_ms=timeout, elapsed_ms=clock.Elapsed.TotalMilliseconds, reply_count=count,
                native_error=error, reply_status=count>0?(int?)Marshal.ReadInt32(reply,4):null };
        } finally {
            Marshal.FreeHGlobal(reply);
            if (!IcmpCloseHandle(handle)) throw new Exception("close: " + Marshal.GetLastWin32Error());
        }
    }
}
'@
$observations = @()
foreach ($timeout in @(50,150,300,500,750,1000,1500)) {
    $observations += [NativeTimeoutComparison]::Probe('192.0.2.1', $timeout)
}
foreach ($destination in @('198.51.100.1','203.0.113.1')) {
    $observations += [NativeTimeoutComparison]::Probe($destination, 150)
}
$observations | ConvertTo-Json -Depth 4 | Set-Content -LiteralPath (Join-Path $OutputDirectory 'native-timeout-comparison.json') -Encoding utf8
$observations | Format-Table
