#!/usr/bin/env bash

# we must be in the root directory : docker can't build on upper directory tree
curr_dir=`basename "$PWD"`
if [ $curr_dir == "Dockerfiles" ]; then
    echo "You can't run this file from this directory. Please move to the root directory of netm4ul. (cd ..)"
    exit 1
fi

# main container (netm4ul/netm4ul:latest)
docker build -t netm4ul/netm4ul .

# "child" container (every tool into it's own docker)
list_of_container=("nmap" "masscan")
for app in ${list_of_container[@]}; do
    docker build -t netm4ul/netm4ul:$app . -f Dockerfiles/Dockerfile.$app
done