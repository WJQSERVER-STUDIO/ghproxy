#!/bin/sh

APPLICATION=ghproxy

if [ ! -f /data/${APPLICATION}/config/blacklist.json ]; then
    cp /data/${APPLICATION}/blacklist.json /data/${APPLICATION}/config/blacklist.json
fi

if [ ! -f /data/${APPLICATION}/config/whitelist.json ]; then
    cp /data/${APPLICATION}/whitelist.json /data/${APPLICATION}/config/whitelist.json
fi

if [ ! -f /data/${APPLICATION}/config/config.toml ]; then
    cp /data/${APPLICATION}/config.toml /data/${APPLICATION}/config/config.toml
fi

/data/${APPLICATION}/${APPLICATION} -cfg /data/${APPLICATION}/config/config.toml > /data/${APPLICATION}/log/run.log 2>&1 &

sleep 30

while [[ true ]]; do
    # Failure Circuit Breaker
    curl -f --max-time 5 -retry 3 http://127.0.0.1:8080/api/healthcheck || exit 1
    sleep 120
done    