FROM alpine:latest


RUN apk add openvpn openresolv supervisor squid bash && \
    wget "https://raw.githubusercontent.com/ProtonVPN/scripts/master/update-resolv-conf.sh" -O "/etc/openvpn/update-resolv-conf" && \
    chmod +x "/etc/openvpn/update-resolv-conf" && mkdir -p /etc/VPN

COPY supervisord.conf /etc/supervisord.conf
COPY squid.conf /etc/squid/squid.conf


EXPOSE 8888

CMD ["/usr/bin/supervisord", "-c", "/etc/supervisord.conf"]