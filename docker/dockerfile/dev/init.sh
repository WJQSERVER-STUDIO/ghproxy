#!/bin/sh

APPLICATION=ghproxy

if [ ! -f /data/caddy/config/Caddyfile ]; then
    cp /data/caddy/Caddyfile /data/caddy/config/Caddyfile
fi

if [ ! -f /data/${APPLICATION}/config/blacklist.json ]; then
    cp /data/${APPLICATION}/blacklist.json /data/${APPLICATION}/config/blacklist.json
fi

if [ ! -f /data/${APPLICATION}/config/whitelist.json ]; then
    cp /data/${APPLICATION}/whitelist.json /data/${APPLICATION}/config/whitelist.json
fi

if [ ! -f /data/${APPLICATION}/config/config.toml ]; then
    cp /data/${APPLICATION}/config.toml /data/${APPLICATION}/config/config.toml
fi

/data/caddy/caddy run --config /data/caddy/config/Caddyfile > /data/${APPLICATION}/log/caddy.log 2>&1 &

/data/${APPLICATION}/${APPLICATION} -cfg /data/${APPLICATION}/config/config.toml > /data/${APPLICATION}/log/run.log 2>&1 &

while true; do
    sleep 1
done