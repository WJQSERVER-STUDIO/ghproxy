#!/bin/bash

APPLICATON=ghproxy

if [ ! -f /data/caddy/config/Caddyfile ]; then
    cp /data/caddy/Caddyfile /data/caddy/config/Caddyfile
fi

if [ ! -f /data/${APPLICATON}/config/blacklist.yaml ]; then
    cp /data/${APPLICATON}/blacklist.yaml /data/${APPLICATON}/config/blacklist.yaml
fi

if [ ! -f /data/${APPLICATON}/config/config.yaml ]; then
    cp /data/${APPLICATON}/config.yaml /data/${APPLICATON}/config/config.yaml
fi

/data/caddy/caddy run --config /data/caddy/config/Caddyfile > /data/${APPLICATON}/log/caddy.log 2>&1 &

/data/${APPLICATON}/${APPLICATON} > /data/ghproxy/log/run.log 2>&1 &

while [[ true ]]; do
    sleep 1
done    

