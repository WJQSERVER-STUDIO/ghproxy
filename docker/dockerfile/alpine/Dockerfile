FROM wjqserver/caddy:daily-alpine

ARG USER=WJQSERVER-STUDIO
ARG REPO=ghproxy
ARG APPLICATION=ghproxy

RUN mkdir -p /data/www
RUN mkdir -p /data/${APPLICATION}/config 
RUN mkdir -p /data/${APPLICATION}/log 
RUN wget -O /data/www/index.html https://raw.githubusercontent.com/${USER}/${REPO}/main/pages/index.html
RUN wget -O /data/www/favicon.ico https://raw.githubusercontent.com/${USER}/${REPO}/main/pages/favicon.ico
RUN wget -O /data/caddy/Caddyfile https://raw.githubusercontent.com/${USER}/${REPO}/main/caddyfile/release/Caddyfile
RUN VERSION=$(curl -s https://raw.githubusercontent.com/${USER}/${REPO}/main/DEV-VERSION) && \
    wget -O /data/${APPLICATION}/${APPLICATION} https://github.com/${USER}/${REPO}/releases/download/$VERSION/ghproxy-alpine
RUN wget -O /data/${APPLICATION}/config.yaml https://raw.githubusercontent.com/${USER}/${REPO}/main/config/config.yaml
RUN wget -O /data/${APPLICATION}/blacklist.json https://raw.githubusercontent.com/${USER}/${REPO}/main/config/blacklist.json
RUN wget -O /data/${APPLICATION}/whitelist.json https://raw.githubusercontent.com/${USER}/${REPO}/main/config/whitelist.json
RUN wget -O /usr/local/bin/init.sh https://raw.githubusercontent.com/${USER}/${REPO}/main/docker/dockerfile/alpine/init.sh
RUN chmod +x /data/${APPLICATION}/${APPLICATION}
RUN chmod +x /usr/local/bin/init.sh

CMD ["/usr/local/bin/init.sh"]

