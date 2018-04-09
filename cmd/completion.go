package cmd

import (
	"flag"
	"fmt"
)

func generateBashCompletion() string {

	var formatedListArgs string
	flag.VisitAll(func(f *flag.Flag) {
		formatedListArgs += fmt.Sprintf("-%+v ", f.Name)
	})

	return `
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
        
        # list of all args
        type_list="` + formatedListArgs + `"
    
        COMPREPLY=( $(compgen -W "${type_list}" -- ${cur_word}) )
        return 0
    }
    
    # Register _netmaul_complete to provide completion for the following commands
    complete -F _netmaul_complete netmaul netm4ul ./netm4ul
    `
}
