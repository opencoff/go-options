// options.go - Self documenting CLI options parser for Go
//
// Copyright (c) 2012 Simon Menke <simon.menke@gmail.com>
// Changes/Enhancements Copyright (c) 2015-2016 Sudhi Herle <sudhi@herle.net>
//
// This software does not come with any express or implied
// warranty; it is provided "as is". No claim  is made to its
// suitability for any purpose.

// Package options implements a self-documenting command line
// options parsing framework.
//
// The list of options a program wishes to use is specified as a
// lengthy, multi-line string. This string also serves as the
// help-string for the options.
//
// Here is an example option specification:
//
//     usage: example-tool
//     A short description of the command
//     --
//     flag        --flag,-f,FLAG           A description for this flag
//     option=     --option=,-o=,OPTION=    A description for this option
//                                          the description continues here
//     !required=  --required,-r=,REQUIRED= A required option
//     --
//     env_var=    ENV_VAR=                 An environment variable
//     --
//     help        help,h                   Show this help message
//     run         run                      Run some function
//     --
//     Additional help for options or defaults etc. go here.
package options

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Representation of a parsed option specification.
type Spec struct {
	usage string

	allow_unknown_args bool

	options     map[string]string
	defaults    map[string]string
	flags       map[string]bool
	required    map[string]bool
	environment map[string]string
	commands    map[string]string
}

// Representation of parsed command line arguments according to a
// given option specification
type Options struct {
	options map[string]string

	// If the options are repeated on the command line, second and subsequent
	// values are in optionv
	optionv map[string][]string

	defaults map[string]string
	Command  string
	Args     []string
}

// Parse a spec string and return a Spec object
func Parse(desc string) (spec *Spec, err error) {
	spec = new(Spec)
	spec.options = make(map[string]string, 0)
	spec.defaults = make(map[string]string, 0)
	spec.flags = make(map[string]bool, 0)
	spec.required = make(map[string]bool, 0)
	spec.commands = make(map[string]string, 0)
	spec.environment = make(map[string]string, 0)
	spec.allow_unknown_args = false

	g_indent := -1
	indent := -1
	section := 0
	lines := []string{}

	for _, line := range strings.Split(desc, "\n") {
		if g_indent == -1 {
			clean_line := strings.TrimLeft(line, " \t")
			if clean_line != "" {
				g_indent = len(line) - len(clean_line)
			}
		} else {
			line = line[g_indent:]
		}

		line := strings.TrimRight(line, " \t")

		if line == "" {
			if section != 1 && section != 2 && section != 3 {
				lines = append(lines, line)
			}
			continue
		}

		if section == 1 || section == 2 || section == 3 {
			if strings.HasPrefix(line, "#") {
				if indent == -1 {
					indent = len(line) - len(strings.TrimLeft(line[1:], " \t"))
				}

				if line == "#" {
					lines = append(lines, "")
				} else {
					line = line[indent:]
					lines = append(lines, line)
				}
				continue
			}
		}

		switch section {

		case 0: // usage
			if line == "--" {
				if len(lines) > 0 && lines[len(lines)-1] != "" {
					lines = append(lines, "")
				}
				section += 1
				continue
			}

			lines = append(lines, line)

		case 1: // options
			if line == "--" {
				if len(lines) > 0 && lines[len(lines)-1] != "" {
					lines = append(lines, "")
				}
				section += 1
				continue
			}

			parts := strings.SplitN(line, " ", 2)
			if len(parts) == 1 {
				err = fmt.Errorf("Invalid option spec: %s", line)
				return
			}
			if indent == -1 {
				indent = len(line) - len(strings.TrimLeft(parts[1], " \t"))
			}
			option := parts[0]
			line = strings.Trim(parts[1], " \t")

			required := false
			flag := true

			if strings.HasPrefix(option, "!") {
				option = option[1:]
				required = true
			}

			if strings.Contains(option, "=") {
				ks := strings.Split(option, "=")
				option = ks[0]
				if defval := ks[1]; len(defval) > 0 {
					spec.defaults[option] = defval
				}
				flag = false
			}

			spec.flags[option] = flag
			spec.required[option] = required

			parts = strings.SplitN(line, " ", 2)
			if len(parts) == 1 {
				parts = append(parts, "-")
			}
			parts[1] = strings.Trim(parts[1], " \t")

			if parts[1] != "-" {
				lines = append(lines, "  "+line)
			}

			parts = strings.Split(parts[0], ",")

			for _, part := range parts {
				pieces := strings.SplitN(part, "=", 2)
				part = pieces[0]

				if strings.HasPrefix(part, "--") || strings.HasPrefix(part, "-") {
					spec.options[part] = option
					continue
				}

				spec.environment[part] = option
			}

		case 2: // environment variables
			if line == "--" {
				if len(lines) > 0 && lines[len(lines)-1] != "" {
					lines = append(lines, "")
				}
				section += 1
				continue
			}

			parts := strings.SplitN(line, " ", 2)
			if len(parts) == 1 {
				err = fmt.Errorf("Invalid env spec: %s", line)
				return
			}
			if indent == -1 {
				indent = len(line) - len(strings.TrimLeft(parts[1], " \t"))
			}
			env := parts[0]
			line = strings.Trim(parts[1], " \t")

			required := false
			flag := true

			if strings.HasPrefix(env, "!") {
				env = env[1:]
				required = true
			}

			if strings.HasSuffix(env, "=") {
				env = env[0 : len(env)-1]
				flag = false
			}

			spec.flags[env] = flag
			spec.required[env] = required

			parts = strings.SplitN(line, " ", 2)
			if len(parts) == 1 {
				parts = append(parts, "-")
			}
			parts[1] = strings.Trim(parts[1], " \t")

			if parts[1] != "-" {
				lines = append(lines, "  "+line)
			}

			parts = strings.Split(parts[0], ",")

			for _, part := range parts {
				part = strings.SplitN(part, "=", 2)[0]
				spec.environment[part] = env
			}

		case 3: // commands
			if line == "--" {
				if len(lines) > 0 && lines[len(lines)-1] != "" {
					lines = append(lines, "")
				}
				section += 1
				continue
			}

			if line == "*" {
				spec.allow_unknown_args = true
				continue
			}

			parts := strings.SplitN(line, " ", 2)
			if len(parts) == 1 {
				err = fmt.Errorf("Invalid command spec: %s", line)
				return
			}
			if indent == -1 {
				indent = len(line) - len(strings.TrimLeft(parts[1], " \t"))
			}
			command := parts[0]
			line = strings.Trim(parts[1], " \t")

			parts = strings.SplitN(line, " ", 2)
			if len(parts) == 1 {
				parts = append(parts, "-")
			}
			parts[1] = strings.Trim(parts[1], " \t")

			if parts[1] != "-" {
				lines = append(lines, "  "+line)
			}

			parts = strings.Split(parts[0], ",")
			for _, part := range parts {
				spec.commands[part] = command
			}

		case 4: // appendix
			if line == "--" {
				if len(lines) > 0 && lines[len(lines)-1] != "" {
					lines = append(lines, "")
				}
				section += 1
				continue
			}

			lines = append(lines, line)

		}
	}

	spec.usage = strings.Join(lines, "\n") + "\n"
	spec.usage = strings.Trim(spec.usage, " \t\n")
	//fmt.Printf("Parsed data:\n%+v\n", spec)
	return
}

// Parse a spec string and die if it fails
func MustParse(desc string) *Spec {
	var p *Spec
	var err error

	if p, err = Parse(desc); err != nil {
		fmt.Fprintf(os.Stderr, "Spec parse error for\n'%.80s' ..\n%s\n", desc, err)
		os.Exit(1)
	}

	return p
}

// Parse the command line arguments in 'args' and the environment
// variables in 'environ'. This expects the parsing to succeed and
// exits with usage string and error if the parsing fails.
func (this *Spec) MustInterpret(args []string, environ []string) *Options {
	opts, err := this.Interpret(args, environ)
	if err != nil {
		this.PrintUsageWithError(err)
	}

	return opts
}

// Parse the command line arguments in 'args' and the environment
// variables in 'environ'. Return the resulting, parsed options in
// 'o' and any error in 'err'.
func (spec *Spec) Interpret(args []string, environ []string) (o *Options, err error) {
	opts := new(Options)
	opts.options = make(map[string]string, 0)
	opts.optionv = make(map[string][]string, 0)
	opts.defaults = spec.defaults
	opts.Args = []string{}

	for _, env := range environ {
		parts := strings.SplitN(env, "=", 2)
		if option, present := spec.environment[parts[0]]; present {
			opts.options[option] = parts[1]
		}
	}

	//fmt.Printf("Options: %+v\n", spec.options)

	for i := 1; i < len(args); i++ {
		arg := args[i]

		// A lone "--" terminates option parsing
		if arg == "--" {
			if i+1 < len(args) {
				opts.Args = append(opts.Args, args[i+1:]...)
			}
			break
		}

		if strings.HasPrefix(arg, "--") || strings.HasPrefix(arg, "-") {
			option := "-"
			value := "true"

			parts := strings.SplitN(arg, "=", 2)

			//fmt.Printf("<< arg %d: %s >>: parts = %v\n", i, arg, parts)

			if len(parts) == 2 {
				option = parts[0]
			} else {
				option = arg
			}

			if opt, present := spec.options[option]; present {
				option = opt
			} else {
				err = fmt.Errorf("Invalid option: %s was not recognized", arg)
				return
			}

			if spec.flags[option] {
				if len(parts) == 2 {
					err = fmt.Errorf("Invalid option: %s was not recognized (doesn't take a value)", arg)
					return
				}
			} else {
				if len(parts) == 2 {
					value = parts[1]
				} else if len(args) > i+1 {
					value = args[i+1]
					i++
				} else {
					err = fmt.Errorf("Invalid option: %s was not recognized (requires a value)", arg)
					return
				}
			}

			// second and subsequent options go in optionv
			if _, ok := opts.options[option]; ok {
				opts.optionv[option] = append(opts.optionv[option], value)
			} else {
				opts.options[option] = value
			}
			continue
		}

		if command, present := spec.commands[arg]; present {
			opts.Command = command
			opts.Args = args[i:]
			opts.Args[0] = opts.Command
			break
		}

		if spec.allow_unknown_args {
			opts.Args = append(opts.Args, arg)
			continue
		}

		err = fmt.Errorf("Invalid argument: %s was not recognized", arg)
		return
	}

	for option, required := range spec.required {
		if _, present := opts.options[option]; required && !present {
			err = fmt.Errorf("Missing option: %s", option)
			return
		}
	}

	for env, option := range spec.environment {
		if value, present := opts.options[option]; present {
			os.Setenv(env, value)
		}
	}

	o = opts
	return
}

// Print the usage string to STDOUT
func (spec *Spec) PrintUsage() {
	fmt.Fprintf(os.Stdout, "%s\n", spec.usage)
}

// Print the usage string to STDOUT and exit with a non-zero code.
func (spec *Spec) PrintUsageAndExit() {
	spec.PrintUsage()
	os.Exit(1)
}

// Print the error string corresponding to 'err' and then show the
// usage string. Both are sent to STDERR. Exit with a non-zero code.
func (spec *Spec) PrintUsageWithError(err error) {
	fmt.Fprintf(os.Stderr, "error: %s\n%s\n", err, spec.usage)
	os.Exit(1)
}

// Return the option corresponding to 'nm'. If the option is not set
// (provided on the command line), the bool retval will be False.
func (opts *Options) Get(nm string) (string, bool) {
	if v, ok := opts.options[nm]; ok {
		return v, true
	}

	if v, ok := opts.defaults[nm]; ok {
		return v, true
	}

	return "", false
}

// For options that are providd multiple times, return all of them in a
// slice. A nil slice implies the option was not set on the command line.
func (opts *Options) GetMulti(nm string) []string {
	var rv []string
	var v string
	var ok bool

	if v, ok = opts.options[nm]; !ok {
		return nil
	}

	rv = append(rv, v)
	rv = append(rv, opts.optionv[nm]...)
	return rv
}

// Interpret the option corresponding to the key 'nm' as
// a Bool and parse it. A failed parse defaults to False.
func (opts *Options) GetBool(nm string) bool {
	if v, ok := opts.Get(nm); ok {
		switch strings.ToLower(v) {
		case "true", "ok", "1", "yes", "on":
			return true

		default:
			return false
		}
	}

	return false
}

// Interpret the option corresponding to the key 'nm' as a signed
// integer (auto-detected base). The second retval will be false if
// the parse fails or the key is not found.
func (opts *Options) GetInt(nm string) (int64, bool) {
	if v, ok := opts.Get(nm); ok {
		if i, err := strconv.ParseInt(v, 0, 64); err == nil {
			return i, true
		}
	}
	return 0, false
}

// Interpret the option corresponding to the key 'nm' as an unsigned
// integer (auto-detected base). The second retval will be false if
// the parse fails or the key is not found.
func (opts *Options) GetUint(nm string) (uint64, bool) {
	if v, ok := opts.Get(nm); ok {
		if i, err := strconv.ParseUint(v, 0, 64); err == nil {
			return i, true
		}
	}
	return 0, false
}

// Return true if the option with the key 'nm' is set (i.e., provided
// on the command line).
func (opts *Options) IsSet(nm string) bool {
	_, ok := opts.options[nm]
	return ok
}

// vim: ft=go:sw=4:ts=4:tw=78:expandtab:
