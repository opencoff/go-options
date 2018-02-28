package options

import (
	"fmt"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	_, err := Parse(`
    usage: haraway <flags>... <command> <args>...
    more summary
    --
    #         Options
    root=     -r,--root=,HARAWAY_ROOT     Path to the haraway data root
    prefix=   -p,--prefix,HARAWAY_PREFIX  Path to the haraway install prefix.
    verbose   -v,--verbose                Show more info
    debug     -d,--debug,HARAWAY_DEBUG    Show debug info
    --
    --
    #         Commands
    exec      exec                        Execute a command within the haraway sanbox
    shell     sh,shell                    Open a shell within the haraway sanbox
    --
    `)

	if err != nil {
		t.Error(err)
	}
}

func TestMulti(t *testing.T) {
	spec, err := Parse(`
    usage: multi <flags>... <command> <args>...
    --
    #         Options
    include=, -I,--include=,    Add dir to include search path
    exclude=, -X,--exclude=     Add dir to exclude search path
    --
    --
    --
    `)

	if err != nil {
		t.Error(err)
	}

	argv := []string{"multi", "-I", "/usr/local", "--include=/usr/include", "-I=/foo"}
	oo, err := spec.Interpret(argv, []string{})
	if err != nil {
		t.Error(err)
	}

	z, ok := oo.Get("include")
	if !ok {
		t.Error("expected to find at least one -I; found none!")
	}

	if z != "/usr/local" {
		t.Errorf("expected to see /usr/local, saw %s", z)
	}

	zv := oo.GetMulti("include")
	if len(zv) != 3 {
		t.Errorf("expected to see 3 entries, saw %d", len(zv))
	}
}

func TestParse(t *testing.T) {
	spec, err := Parse(`
    usage: haraway <flags>... <command> <args>...
    --
    root=     -r,--root=,HARAWAY_ROOT     Path to the haraway data root
    prefix=   -p,--prefix,HARAWAY_PREFIX  Path to the haraway install prefix.
    verbose   -v,--verbose                Show more info
    debug     -d,--debug,HARAWAY_DEBUG    Show debug info
    --
    --
    exec      c,exec                      Execute a command within the haraway sanbox
    shell     sh,shell                    Open a shell within the haraway sanbox
    --
    `)
	if err != nil {
		t.Error(err)
	}

	opts, err := spec.Interpret([]string{"haraway", "-p", "/usr/local", "-r=hello", "-v", "c", "ls"}, []string{})

	if err != nil {
		t.Fatal(err)
	}

	if v, ok := opts.Get("root"); ok && v != "hello" {
		t.Error("--root != hello")
	}

	if v, ok := opts.Get("verbose"); ok && v != "true" {
		t.Error("--verbose != true")
	}

	if opts.GetBool("verbose") != true {
		t.Error("--verbose != true (bool)")
	}

	if strings.Join(opts.Args, " ") != "exec ls" {
		t.Errorf(".Args != [`exec`, `ls`] (was: %+v)", opts.Args)
	}
}

func ExampleParse() {
	spec, err := Parse(`
    usage: example-tool
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
    More freestyle text
    `)
	if err != nil {
		spec.PrintUsageWithError(err)
	}

	opts, err := spec.Interpret([]string{"example-tool", "--required", "hello world"}, []string{})
	if err != nil {
		spec.PrintUsageWithError(err)
	}

	v, _ := opts.Get("required")
	fmt.Printf("required: %s", v)

	// Output:
	// required: hello world
}

func TestDefaults(t *testing.T) {
	spec, err := Parse(`
    usage: haraway <flags>... <command> <args>...
    --
    root=XYZ  -r,--root=,HARAWAY_ROOT     Path to the haraway data root
    num=2     -n=                         Path to the haraway install prefix.
    --
    --
    exec      c,exec                      Execute a command within the haraway sanbox
    shell     sh,shell                    Open a shell within the haraway sanbox
    --
    `)
	if err != nil {
		t.Error(err)
	}

	opts, err := spec.Interpret([]string{"haraway"}, []string{})

	if err != nil {
		t.Fatal(err)
	}

	if v, ok := opts.Get("root"); ok && v != "XYZ" {
		t.Error("--root != XYZ")
	}

	if v, ok := opts.GetInt("num"); ok && v != 2 {
		t.Errorf("--num != 2; saw %v", v)
	}

	opts, err = spec.Interpret([]string{"haraway", "-r", "hello", "-n=5"}, []string{})

	if err != nil {
		t.Fatal(err)
	}

	if v, ok := opts.Get("root"); ok && v != "hello" {
		t.Error("--root != hello")
	}

	if v, ok := opts.GetInt("num"); ok && v != 5 {
		t.Error("-n != 5")
	}
}
