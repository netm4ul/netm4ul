
# yes : 0
# no : 1
ask_yes_no(){
    # The first argument is the question
    read -n 1 -r -p "$1" input
    if [[ ! $input =~ ^[Yy]$ ]]; then
        return 0
    fi
    return 1
}

# └ ┘ ┤ ├ │ ┌ ┐ ┴ ┬ ┼ ─
# ╚ ╝ ╠ ╣ ║ ╔ ╗ ╩ ╦ ╬ ═

banner(){
    VERSION=$(git describe --abbrev=8 --always --tags)
    echo -e " ╔╦╦═─═════───════───═══─═════════════════─═════════──══════───═══╦╦╗";
    echo -e " ╠╩╝                                                              ╚╩╣";
    echo -e " │  ███╗   ██╗███████╗████████╗███╗   ███╗██╗  ██╗██╗   ██╗██╗      ║";
    echo -e " ║  ████╗  ██║██╔════╝╚══██╔══╝████╗ ████║██║  ██║██║   ██║██║      │";
    echo -e " ║  ██╔██╗ ██║█████╗     ██║   ██╔████╔██║███████║██║   ██║██║      ║";
    echo -e " ║  ██║╚██╗██║██╔══╝     ██║   ██║╚██╔╝██║╚═╦══██║██║   ██║██║      │";
    echo -e " │  ██║ ╚████║███████╗   ██║   ██║ ╚╦╝ ██║  │  ██║╚██████╔╝███████╗ ║";
    echo -e " ║  ╚╦╝  ╠═══╝╚═┬═══─╣   ╠═╝   ╚═╬══╣  ╚═╣  ╠══╩═╝ ╚══╦══╝ ╚══════╩─╣";
    echo -e " ║   ║   └───┬──┘    ╚══╦╩──────┬╝  ╚═══╦╩──┘        ┌┴──═╗         ║";
    echo -e " ╚═══╩═══════┴════──═╦══╩═══════╩═════──┴─═══════════╩════╩══════╦══╝";
    echo -e " ╔═══╦═══════════════╩══════╦══════════════════════╦═════════════╩══╗"
    echo -e " ╠ Author : Edznux                           ${GREEN}Git version : $VERSION ${COLOR_RESET}╣"
    echo -e " ╚═══╩══════════════════════╩══════════════════════╩════════════════╝"
}

usage(){
    banner
    cat <<USAGE

Usage : $ME [options]

Valid options are:
   --install /path/to/bin
        If you give: --install /opt
        netm4ul will be installed as /opt/netm4ul
        Default : $INSTALL_PATH

   --config /path/to/config/
        If you give: --config /home/user/configs
        netm4ul will use the /home/user/configs/netm4ul/netm4ul.conf file
        (Note the new "netm4ul" folder inside /home/user/configs)
        Default : $CONFIG_DIR

   --force-root
        This option enable you to run this script as root (not recommended)

This installer is for the git version.
You can find the pre-built package at https://github.com/netm4ul/netm4ul

USAGE
}

print_new_version_available(){
    echo -e "${RED} ╔══════════════════════════════════════════════════════════════════╗"
    echo -e " ╠         A New version is available on the master branch          ╣"
    echo -e " ╚══════════════════════════════════════════════════════════════════╝${COLOR_RESET}"
}


is_up_to_date(){
    count_diff_commit=$(git rev-list origin/master...HEAD --count)
    if [[ $count_diff_commit -ne 0 ]]; then
        return 1 # not up to date
    fi
    return exit 0
}

progress(){
    echo -e "${WHITE}----[ " $* " ]${COLOR_RESET}"
}

subprogress(){
    echo -e "${WHITE}------------[ " $2 $* "${WHITE} ]${COLOR_RESET}"
}