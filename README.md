# ZTBeacon

TLS Beacon for Cloudflare ZeroTrust

Copyright (c) 2023 Tenebris Technologies Inc. See LICENSE for details.

This is untested alpha code. Use at your own risk.

To get started, run `ztbeacon -newcert -console` to create a self signed certificate and key in the current directory
and print the log, which includes the SHA256 hash (fingerprint) required for the Cloudflare ZeroTrust configuration
to the console.

See `ztbeacon -help` for additional command line options.
