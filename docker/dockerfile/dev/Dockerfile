FROM wjqserver/caddy:daily-alpine AS builder

ARG USER=WJQSERVER-STUDIO
ARG REPO=ghproxy
ARG APPLICATION=ghproxy
ARG TARGETOS
ARG TARGETARCH
ARG TARGETPLATFORM

# 创建文件夹
RUN mkdir -p /data/www
RUN mkdir -p /data/${APPLICATION}/config 
RUN mkdir -p /data/${APPLICATION}/log 

# 安装依赖
RUN apk add --no-cache curl wget

# 前端
RUN wget -O /data/www/index.html https://raw.githubusercontent.com/${USER}/${REPO}/main/pages/index.html
RUN wget -O /data/www/favicon.ico https://raw.githubusercontent.com/${USER}/${REPO}/main/pages/favicon.ico

# 后端
RUN VERSION=$(curl -s https://raw.githubusercontent.com/${USER}/${REPO}/main/DEV-VERSION) && \
    wget -O /data/${APPLICATION}/${APPLICATION} https://github.com/${USER}/${REPO}/releases/download/$VERSION/${APPLICATION}-${TARGETOS}-${TARGETARCH}
RUN wget -O /usr/local/bin/init.sh https://raw.githubusercontent.com/${USER}/${REPO}/main/docker/dockerfile/dev/init.sh

# 拉取配置
RUN wget -O /data/caddy/Caddyfile https://raw.githubusercontent.com/${USER}/${REPO}/main/caddyfile/dev/Caddyfile
RUN wget -O /data/${APPLICATION}/config.yaml https://raw.githubusercontent.com/${USER}/${REPO}/main/config/config.yaml
RUN wget -O /data/${APPLICATION}/blacklist.json https://raw.githubusercontent.com/${USER}/${REPO}/main/config/blacklist.json
RUN wget -O /data/${APPLICATION}/whitelist.json https://raw.githubusercontent.com/${USER}/${REPO}/main/config/whitelist.json

# 权限
RUN chmod +x /data/${APPLICATION}/${APPLICATION}
RUN chmod +x /usr/local/bin/init.sh

FROM wjqserver/caddy:daily-alpine

COPY --from=builder /data/www /data/www
COPY --from=builder /data/caddy /data/caddy
COPY --from=builder /data/${APPLICATION} /data/${APPLICATION}
COPY --from=builder /usr/local/bin/init.sh /usr/local/bin/init.sh

# 权限
RUN chmod +x /data/${APPLICATION}/${APPLICATION}
RUN chmod +x /usr/local/bin/init.sh

CMD ["/usr/local/bin/init.sh"]

