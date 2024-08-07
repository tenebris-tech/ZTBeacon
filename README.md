# ZTBeacon

TLS Beacon for Cloudflare ZeroTrust

Copyright (c) 2023-2024 Tenebris Technologies Inc. See LICENSE for details.

This is untested alpha code. Use at your own risk.

To get started, run `ztbeacon -newcert -console` to create a self signed certificate and key in the current directory
and print the log, which includes the SHA256 hash (fingerprint) required for the Cloudflare ZeroTrust configuration,
to the console.

By default, connection information is not logged to reduce unnecessary log volume. To enable connection logging,
use the `-debug` option.

If you have a Cloudflare tunnel connected to the network, you may use the `-deny <ip>` option to block access from the
server handling the tunnel. This will prevent off-network access to the beacon.

See `ztbeacon -help` for additional command line options.
