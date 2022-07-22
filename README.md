# meta

`meta` is a meta command line interface (CLI) that creates subcommands
automatically from the executables found in the PATH environment variable.

The user experience is focused on:

* simple initial setup
* ease of adding subcommands
* allowing any exectuable

The top level command can be setup by simply symlinking to `/usr/bin/meta`. For
example, this creates a new top level `hello` command:

```
$ go install github.com/calebcase/meta
$ ln -s "$(which meta)" ~/bin/hello
```

Executing `hello` without arguments will output help and exit non-zero:

```
$ hello
Usage: hello COMMAND
```

Subcommands are created through a simple naming convention
`command_subcommand`. For example, this creates a `world` subcommand:

```
$ cat <<EOF >~/bin/hello_world
#!/bin/bash
set -euo pipefail

printf 'Hello World!\n'
EOF
$ chmod u+x ~/bin/hello_world
```

NOTE: Subcommands can be any executable file. We've used bash here as an
example, but this could also be a python script, a binary, etc.

Now when `hello` outputs usage it has the `world` subcommand automatically:

```
$ hello
Usage: hello COMMAND

Commands:
  world
```

Running the subcommand:

```
$ hello world
Hello World!
```
