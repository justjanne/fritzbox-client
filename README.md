# Fritzbox CLI client

## fritzbox-sip

Allows connecting, disconnecting and reconnecting SIP numbers

```
Usage: fritzbox-cert --host HOST --user USER --pass PASS <connect|disconnect|reconnect> [ids]
```

## fritzbox-cert

Allows updating the TLS certificate automatically (e.g., as acme post-hook)

```
Usage: fritzbox-cert --host HOST --user USER --pass PASS --key KEY --cert CERT [--keypass KEYPASS]
```

Inspired by [wikrie]/[fritzbox-cert-update.sh]

[wikrie]: https://github.com/wikrie

[fritzbox-cert-update.sh]: https://gist.github.com/wikrie/f1d5747a714e0a34d0582981f7cb4cfb
