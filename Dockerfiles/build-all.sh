#!/usr/bin/env bash

list_of_container=("nmap" "masscan")
for app in ${list_of_container[@]}; do
    docker build -t netm4ul/netm4ul:$app . -f Dockerfile.$app
done