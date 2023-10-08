FROM alpine:latest


RUN apk add openvpn openresolv supervisor squid dante-server bash && \
    wget "https://raw.githubusercontent.com/ProtonVPN/scripts/master/update-resolv-conf.sh" -O "/etc/openvpn/update-resolv-conf" && \
    chmod +x "/etc/openvpn/update-resolv-conf"

COPY supervisord.conf /etc/supervisord.conf
COPY squid.conf /etc/squid/squid.conf
COPY sockd.conf /etc/sockd.conf


EXPOSE 8888 8881

CMD ["/usr/bin/supervisord", "-c", "/etc/supervisord.conf"]