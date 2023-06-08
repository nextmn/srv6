# NextMN-SRv6
NextMN-SRv6 is an experimental implementation for some [SRv6 MUP Endpoint Behaviors](https://datatracker.ietf.org/doc/draft-ietf-dmm-srv6-mobile-uplane/).
This project is still at the early stages of development and contains bugs and will crash in unexpected manners.
Please do not use it for anything other than experimentation. Expect breaking changes until v1.0.0

## Roadmap
Behavior | Implemented? | Todo
---|---|---
End.MAP | no | -
End.M.GTP6.D | no | -
End.M.GTP6.D.Di | partial | handle GTP6 packets using go-packet instead of go-gtp, respect S04, send ICMP when errors, optional: respond to GTP Echo Req
End.M.GTP6.E | partial | create GTP6 packets using go-packet instead of go-gtp, send ICMP when errors
End.M.GTP4.E | partial | create GTP4 packets using go-packet instead of go-gtp, send ICMP when errors, IPv4 SA from SRv6 DA 
H.M.GTP4.D | partial | handle GTP4 packets using go-packets instead of go-gtp, send ArgsMobSession and IPv4 DA, send ICMP when errors, optional: respond to GTP Echo Req
End.Limit | no | -

PDU Session Type | Supported?
---|---
IPv4 | partial
IPv6 | partial
IPv4v6 | yes
Ethernet | no
Unstructured | no

TODO: SR Policy set by [nextmn-srv6-ctrl](https://github.com/nextmn-srv6-ctrl)

## Getting started
### Build dependencies
- golang
- make (optional)

### Runtime dependencies
- iproute2

### Build and install
Simply run `make build` and `make install`.

### Docker
If you plan using NextMN-UPF with Docker:
- The container requires the `NET_ADMIN` capability;
- The container should enable IPv6, and Segment Routing
- The container requires the forwarding to be enabled (not enabled by the UPF itself);
- The tun interface (`/dev/net/tun`) must be available in the container.

This can be done in `docker-compose.yaml` by defining the following for the service:

```yaml
cap_add:
    - NET_ADMIN
devices:
    - "/dev/net/tun"
sysctls:
    - net.ipv6.conf.all.disable_ipv6=0
    - net.ipv4.ip_forward=1
    - net.ipv6.conf.all.forwarding=1
    - net.ipv6.conf.all.seg6_enabled=1
    - net.ipv6.conf.default.seg6_enabled=1
```

## Author
Louis Royer

## License
MIT
