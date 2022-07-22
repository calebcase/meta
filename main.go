package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

//TODO: Make these configurable.
var (
	// SEP is the separator between subcommand names.
	SEP = "_"

	// SUB_MAX is the maximum number of levels to consider when outputing
	// subcommands. It defaults to 1 (or just the immediate subcommands).
	SUB_MAX = 1
)

type Name string

func (name Name) Encode() string {
	return strings.ReplaceAll(string(name), " ", SEP)
}

func (name Name) Decode() string {
	return strings.ReplaceAll(string(name), SEP, " ")
}

func (name Name) Parts() []string {
	return strings.Split(name.Decode(), " ")
}

func GetSubCommands(prefix string) map[string]string {
	prefix = Name(prefix).Encode()
	commands := map[string]string{}

	paths := filepath.SplitList(os.Getenv("PATH"))
	for _, path := range paths {
		matches, err := filepath.Glob(filepath.Join(path, prefix+SEP+"*"))
		if err != nil {
			continue
		}

		for _, match := range matches {
			base := filepath.Base(match)
			subbase := base[len(prefix+SEP):]
			subparts := Name(subbase).Parts()
			if len(subparts) <= SUB_MAX {
				if _, ok := commands[subbase]; !ok {
					commands[subbase] = match
				}
			}
		}
	}

	return commands
}

func GenUsage(prefix string) (usage []string) {
	usage = []string{
		fmt.Sprintf("Usage: %s COMMAND", Name(prefix).Decode()),
	}

	commands := GetSubCommands(prefix)

	if len(commands) > 0 {
		usage = append(usage, "", "Commands:")

		for command := range commands {
			usage = append(usage, fmt.Sprintf("  %s %s", Name(prefix).Decode(), Name(command).Decode()))
		}
	}

	return usage
}

func Execute() (err error) {
	name := filepath.Base(os.Args[0])

	if len(os.Args) == 1 {
		usage := GenUsage(name)
		fmt.Fprintln(os.Stderr, strings.Join(usage, "\n"))
		return errors.New("missing command")
	}

	if len(os.Args) >= 2 {
		args := []string{}
		copy(args, os.Args)

		// Find the first N non-optional arguments as the subcommand.
		found := 0
		subcommand := []string{}
		position := -1
		for i, arg := range os.Args[1:] {
			if arg == "--" {
				break
			}

			if strings.HasPrefix(arg, "-") {
				continue
			}

			subcommand = append(subcommand, arg)
			position = i
			found++

			args = append(os.Args[:position], os.Args[position+1:]...)

			if found >= SUB_MAX {
				break
			}
		}

		// If there weren't any subcommands, but there were flags...
		if position == -1 && len(os.Args) > 0 {
			switch os.Args[1] {
			case "--help", "-h":
				usage := GenUsage(name)
				fmt.Fprintln(os.Stdout, strings.Join(usage, "\n"))

				return nil
			default:
				return errors.New("flag not found")
			}
		}

		subcommands := GetSubCommands(name)
		executable, ok := subcommands[strings.Join(subcommand, SEP)]
		if !ok {
			return errors.New("command not found")
		}

		return syscall.Exec(executable, args, os.Environ())
	}

	return nil
}

func main() {
	err := Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n%s: %s\n", os.Args[0], err)
		os.Exit(1)
	}
}
