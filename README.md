# NextMN-SRv6
NextMN-SRv6 is an experimental implementation for some [SRv6 MUP SIDs](https://datatracker.ietf.org/doc/draft-ietf-dmm-srv6-mobile-uplane/).
This project is still at the early stages of development and contains bugs and will crash in unexpected manners.
Please do not use it for anything other than experimentation. Expect breaking changes until v1.0.0


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
If you plan using NextMN-UPF with Docker:
- The container required the `NET_ADMIN` capability;
- The container required the forwarding to be enabled (not enabled by the UPF itself);
- The tun interface (`/dev/net/tun`) must be available in the container.

This can be done in `docker-compose.yaml` by defining the following for the service:

```yaml
cap_add:
    - NET_ADMIN
devices:
    - "/dev/net/tun"
sysctls:
    - net.ipv4.ip_forward=1
```

## Author
Louis Royer

## License
MIT
