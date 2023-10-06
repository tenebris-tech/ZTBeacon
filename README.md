# ZTBeacon

TLS Beacon for Cloudflare ZeroTrust

Copyright (c) 2023 Tenebris Technologies Inc. See LICENSE for details.

This is untested alpha code. Use at your own risk.

To get started, run `ztbeacon -newcert -console` to create a self signed certificate and key in the current directory
and print the log, which includes the SHA256 hash (fingerprint) required for the Cloudflare ZeroTrust configuration
to the console.

See `ztbeacon -help` for additional command line options.

Notes: 

(1) The Cloudflare WARP (ZeroTrust) client does not appear to complete the HTTP request after obtaining the
certificate, so it is normal not to see any activity in the log.

(2) If you have a Cloudflare tunnel connected to the network, ensure that hosts connecting via the tunnel
can not reach the ZTBeacon port or they may think they are on the internal network.
