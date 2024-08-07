# ZTBeacon

TLS Beacon for Cloudflare ZeroTrust

Copyright (c) 2023-2024 Tenebris Technologies Inc. See LICENSE for details.

No warrantee expressed or implied. Use at your own risk.

This is intended as an easier and more reliable alternative to manual certificate creation and writing you own
Python server as suggested in https://blog.cloudflare.com/location-aware-warp

To get started, run `ztbeacon -newcert -console` to create a self signed certificate and key in the current directory
and print the log, which includes the SHA256 hash (fingerprint) required for the Cloudflare ZeroTrust configuration,
to the console.

By default, connection information is not logged to reduce unnecessary log volume. To enable connection logging,
use the `-debug` option.

If you have a Cloudflare tunnel connected to the network, you may use the `-deny <ip>` option to block access from the
server handling the tunnel. This will prevent off-network access to the beacon.

See `ztbeacon -help` for additional command line options.

ztbeacon.service is an example service file for Linux users.
