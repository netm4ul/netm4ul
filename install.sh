#!/usr/bin/env bash

set -e

# See this file for the differents variables value (installation path...)
source ./scripts/global_var.sh

source ./scripts/function.sh

# This program will need sudo (and the password if required in the sudoers) to complete successfuly
# This privilege will be needed for:
# - Write permission to /etc/ ("/etc/netm4ul/", "/etc/netm4ul/netm4ul.conf" and "/etc/bash_completion.d/netm4ul")
# - Write permission to /etc/local/bin/netm4ul
ME=$0

while [ ! -z "${1}" ]; do
    if [ "$1" == "--install" ]; then
        INSTALL_PATH="${2}"
        shift 2
    elif [ "$1" == "--config" ]; then
        CONFIG_DIR="${2}/netm4ul/"
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
        echo "This installer should *NOT* run as root. It will use sudo only on the required parts."
        exit 1
    fi
fi

banner

if [[ is_up_to_date -eq 0 ]]; then
    print_new_version_available
    progress "Downloading new updates"
    git pull
fi

progress "Requierements"
# Check that we have golang installed. Other dependencie will be automatically downloaded & installed
if ! [ -x "$(command -v go)" ]; then
    echo "Golang is not installed ! Aborting"
    exit 1
fi
subprogress "Golang installed" ${GREEN}

# Ensuring that dep is installed.
if ! [ -x "$(command -v dep)" ]; then
    echo "Dep is not installed !"
    echo "Installing..."
    go get -u github.com/golang/dep/cmd/dep
fi

subprogress "Dep installed" ${GREEN}

progress "Compilation"
# This will compile the program
subprogress "Start compilation"
make || exit 1
subprogress "Compilation done" ${GREEN}

if [[ $SKIP_TEST -eq 0 ]]; then
    # We don't exit early (like above) because some tests might not be reliable and or could fail without impact.
    progress "Running tests"
    make test || TEST_FAIL=1

    if [[ $TEST_FAIL -eq 1 ]]; then
        subprogress "Tests failed, continuing anyway" ${RED}
    else
        subprogress "Tests OK" ${GREEN}
    fi
fi

# create config dir
progress "Installing globally"
subprogress "Creating $CONFIG_DIR folder"
[ -d $CONFIG_DIR ] || sudo mkdir $CONFIG_DIR

subprogress "Creating executable in $INSTALL_PATH"
# copy the executable into the install path
sudo cp ./netm4ul $INSTALL_PATH

# Generate & install the autocompletion
# We should add other cases like MacOS, and non-bash shell (zsh) here.
if [ "$SHELL" == "/bin/bash" ]; then
    subprogress "Adding autocompletion"
    [ -d $BASH_COMPLETION_DIR ] && ./netm4ul completion bash | sudo tee $BASH_COMPLETION_DIR$NAME  > /dev/null # ubuntu bash completion
fi

# Check if "netm4ul" is correctly installed
progress "Checking if installation is successful"
if ! [ -x "$(command -v netm4ul)" ]; then
    echo -e "${RED}Could not execute netm4ul command.${COLOR_RESET}Something went wrong ?"
    subprogress "Errored !" ${RED}
else
    subprogress "Done !" ${GREEN}
fi