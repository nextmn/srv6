# NextMN-SRv6
NextMN-SRv6 is an experimental implementation for some [SRv6 MUP Endpoint Behaviors](https://datatracker.ietf.org/doc/draft-ietf-dmm-srv6-mobile-uplane/).

> [!WARNING]
> This project is still at the early stages of development and contains bugs and will crash in unexpected manners.
> Please do not use it for anything other than experimentation. Expect breaking changes until v1.0.0

## Roadmap
Provider | Behavior | Implemented? | Todo
---|---|---|---
NextMN | End.MAP | no | -
NextMN | End.M.GTP6.D | no | -
NextMN | End.M.GTP6.D.Di | no | -
NextMN | End.M.GTP6.E | yes | send ICMP when errors
NextMN | End.M.GTP4.E | yes | send ICMP when errors
NextMN | H.M.GTP4.D | yes | send ICMP when errors, optional: respond to GTP Echo Req
NextMN | End.Limit | no | -
NextMN | [End.M.GTP6.E.Red](https://datatracker.ietf.org/doc/draft-kawakami-dmm-srv6-gtp6e-reduced/) | no | requires a map of gnbs addr; [order of bit field considerations](https://datatracker.ietf.org/meeting/118/materials/slides-118-dmm-draft-kawakami-dmm-srv6-gtp6e-reduced-01)
NextMNWithCtrl | H.M.GTP4.D | partial | -
NextMNWithCtrl | H.Encaps | partial | src port number should not be hardcoded
Linux  | End | yes | -
Linux  | End.DX4 | yes | -
Linux  | H.Encaps | yes | -
Linux  | H.Inline | untested | -

PDU Session Type | Supported?
---|---
IPv4 | yes
IPv6 | no
IPv4v6 | no
Ethernet | no
Unstructured | no


## Incoming packet flow
![incoming packet flow](./doc/img/incoming-packet-flow.svg)

## Getting started
### Build dependencies
- golang
- make (optional)

### Runtime dependencies
- iproute2
- iptables

### Build and install
Simply run `make build` and `make install`.

### Docker
If you plan using NextMN-SRv6 with Docker:
- The container requires the `NET_ADMIN` capability;
- The container should enable IPv6, and Segment Routing
- The container requires the forwarding to be enabled (not enabled by the container itself);
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
