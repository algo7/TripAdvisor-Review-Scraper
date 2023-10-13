# Base image
FROM alpine:latest

RUN apk add openvpn \
    openresolv \
    supervisor \
    squid \
    dante-server bash && \
    # ProtonVPN DNS update scripts: https://protonvpn.com/support/linux-openvpn/
    wget "https://raw.githubusercontent.com/ProtonVPN/scripts/master/update-resolv-conf.sh" -O "/etc/openvpn/update-resolv-conf" && \
    chmod +x "/etc/openvpn/update-resolv-conf"

# Copy the config file of supervisord, Squid HTTP Proxy, and Dante SOCKS Proxy
COPY supervisord.conf sockd.conf /etc/
COPY squid.conf /etc/squid/squid.conf

# VPN config and pass will be bind mounted into this dir
WORKDIR /VPN

# Squid Proxy on 8888, and Dante on 8881
EXPOSE 8888 8881

#  Starts Dante, Squid and OpenVPN via supervisord
CMD ["/usr/bin/supervisord", "-c", "/etc/supervisord.conf"]