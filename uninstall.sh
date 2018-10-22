#!/usr/bin/env bash

set -e

source ./scripts/function.sh
# See this file for the differents variables value (installation path...)
source ./scripts/global_var.sh


ME=$0

while [ ! -z "${1}" ]; do
    if [ "$1" == "--install" ]; then
        INSTALL_PATH="${2}"
        shift 2
    elif [ "$1" == "--config" ]; then
        CONFIG_DIR="${2}/netm4ul"
        shift 2
    elif [ "$1" == "--force-root" ]; then
        FORCE_ROOT=1
        shift 1
    else
        usage
        echo -e "${RED}ERROR: Unknown option '$1'.${COLOR_RESET}"
        exit 1
    fi
done


if [[ $FORCE_ROOT -ne 1 ]]; then  # If the FORCE_ROOT flag is not set, warn the user that they shouldn't run this as root.
    if [[ $EUID -eq 0 ]]; then
        echo "This uninstaller should *NOT* run as root. It will use sudo only on the required parts."
        exit 1
    fi
fi

# remove executable
if sudo test -f "$INSTALL_PATH$NAME"; then
	sudo rm -fv $INSTALL_PATH$NAME
fi

if sudo test -d "$CONFIG_DIR"; then
	sudo rm -rfv $CONFIG_DIR
fi


