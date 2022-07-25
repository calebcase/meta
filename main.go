package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Separator     string `env:"META_SEP" env-default:"_"`
	SubCommandMax int    `env:"META_SUBCMD_MAX" env-default:"1"`
}

var cfg Config

type Name string

func (name Name) Encode() string {
	return strings.ReplaceAll(string(name), " ", cfg.Separator)
}

func (name Name) Decode() string {
	return strings.ReplaceAll(string(name), cfg.Separator, " ")
}

func (name Name) Parts() []string {
	return strings.Split(name.Decode(), " ")
}

type Cmd struct {
	Path  string
	Name  string
	Blurb string
}

func GetBlurb(path string) string {
	output, err := exec.Command(path, "--help-blurb").Output()
	if err != nil {
		return ""
	}

	blurb, _, _ := strings.Cut(string(output), "\n")

	return blurb
}

func GetSubCommands(prefix string) map[string]Cmd {
	prefix = Name(prefix).Encode()
	commands := map[string]Cmd{}

	paths := filepath.SplitList(os.Getenv("PATH"))
	for _, path := range paths {
		matches, err := filepath.Glob(filepath.Join(path, prefix+cfg.Separator+"*"))
		if err != nil {
			continue
		}

		for _, match := range matches {
			base := filepath.Base(match)
			subbase := base[len(prefix+cfg.Separator):]
			subparts := Name(subbase).Parts()
			if len(subparts) <= cfg.SubCommandMax {
				if _, ok := commands[subbase]; !ok {
					commands[subbase] = Cmd{
						Path:  match,
						Name:  subbase,
						Blurb: GetBlurb(match),
					}
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

	cmds := GetSubCommands(prefix)

	if len(cmds) > 0 {
		usage = append(usage, "", "Commands:")

		for subbase, cmd := range cmds {
			usage = append(usage, fmt.Sprintf("  %s %s\t%s", Name(prefix).Decode(), Name(subbase).Decode(), cmd.Blurb))
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

			if found >= cfg.SubCommandMax {
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
		cmd, ok := subcommands[strings.Join(subcommand, cfg.Separator)]
		if !ok {
			return errors.New("command not found")
		}

		return syscall.Exec(cmd.Path, args, os.Environ())
	}

	return nil
}

func main() {
	var err error

	err = cleanenv.ReadEnv(&cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n%s: %s\n", os.Args[0], err)
		os.Exit(1)
	}

	err = Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n%s: %s\n", os.Args[0], err)
		os.Exit(1)
	}
}
