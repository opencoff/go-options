# Options - self documenting go-lang command line parsing

The list of options a program wishes to use is specified as a
lengthy, multi-line string. This string also serves as the
help-string for the options.

Here is an example option specification:

    Usage: example-tool
    A short description of the command
    --
    flag        --flag,-f,FLAG           A description for this flag
    option=     --option=,-o=,OPTION=    A description for this option
                                         the description continues here
    !required=  --required,-r=,REQUIRED= A required option
    --
    env_var=    ENV_VAR=                 An environment variable
    --
    help        help,h                   Show this help message
    run         run                      Run some function
    --
    Additional help for options or defaults etc. go here.

Read the documentation

## Notes
Forked and modified from github.com/fd/options. I've retained the
original license. Sadly, I didn't do a proper fork; so, my list of
changes list is lost.

The godoc is [here](http://go.pkgdoc.org/github.com/opencoff/go-options).
