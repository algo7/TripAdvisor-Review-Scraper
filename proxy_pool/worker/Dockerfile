FROM alpine:latest

WORKDIR /VPN

RUN apk add openvpn openresolv bash && \
    wget "https://raw.githubusercontent.com/ProtonVPN/scripts/master/update-resolv-conf.sh" -O "/etc/openvpn/update-resolv-conf" && \
    chmod +x "/etc/openvpn/update-resolv-conf"

CMD ["openvpn","config.ovpn"]