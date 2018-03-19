package cli

var template = `
_netmaul_complete()
{
    local cur_word prev_word type_list

    # COMP_WORDS is an array of words in the current command line.
    # COMP_CWORD is the index of the current word (the one the cursor is
    # in). So COMP_WORDS[COMP_CWORD] is the current word; we also record
    # the previous word here, although this specific script doesn't
    # use it yet.
    cur_word="${COMP_WORDS[COMP_CWORD]}"
    prev_word="${COMP_WORDS[COMP_CWORD-1]}"

    # Ask netm4ul to generate a list of types it supports
    type_list="-info -config -verbose -version -no-colors -targets -mode -modules -completion -server -client"

    # Only perform completion if the current word starts with a dash ('-'),
    # meaning that the user is trying to complete an option.
    if [[ ${cur_word} == -* ]] ; then
        # COMPREPLY is the array of possible completions, generated with
        # the compgen builtin.
        COMPREPLY=( $(compgen -W "${type_list}" -- ${cur_word}) )
    else
        COMPREPLY=()
    fi
    return 0
}

# Register _netmaul_complete to provide completion for the following commands
complete -F _netmaul_complete netmaul netm4ul ./netm4ul
`
