# ZTBeacon

TLS Beacon for Cloudflare ZeroTrust

Copyright (c) 2023-2024 Tenebris Technologies Inc. See LICENSE for details.

No warrantee expressed or implied. Use at your own risk.

This is intended as an easier and more reliable alternative to manual certificate creation and writing you own
Python server as suggested in https://blog.cloudflare.com/location-aware-warp

### Building
To build ZTBeacon, clone the repo, enter the directory, and use go build.

We recommend using the latest version of Go, available at https://go.dev/dl/

```
git clone https://github.com/tenebris-tech/ZTBeacon.git
cd ZTBeacon
go build
```

ztbeacon.service is an example service file for Linux users.
```
mkdir /opt/ztbeacon
cp ZTBeacon /opt/ztbeacon/ztbeacon
chmod 755 /opt/ztbeacon/ztbeacon
cp ztbeacon.service /etc/systemd/system
systemctl enable ztbeacon
```
Note: You must follow the instructions below the first time to generate a certificate and obtain the SHA256 hash required
to configure CloudFlare ZeroTrust. Then you may use `systemctl start ztbeacon`.

### Getting started
To get started, run `ztbeacon -newcert -console` to create a self signed certificate and key in the current directory
and print the log, which includes the SHA256 hash (fingerprint) required for the Cloudflare ZeroTrust configuration,
to the console.

By default, connection information is not logged to reduce unnecessary log volume. To enable connection logging,
use the `-debug` option.

If you have a Cloudflare tunnel connected to the network, you may use the `-deny <ip>` option to block access from the
server handling the tunnel. This will prevent off-network access to the beacon.

See `ztbeacon -help` for additional command line options.
