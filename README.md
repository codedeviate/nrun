# nrun - The npm script runner

**Current version is v0.12.1**

nrun is a utility to make **npm run** a bit easier, and it has some nice features. It is written in Go which I find easier to use when creating portable executable code.

There is no requirement to have **npm** installed to use nrun. The scripts in package.json are parsed and executed directly.

Even though the goal is to make it portable, nrun will still need a shell to run. So Linux users and Mac users can probably run it smoothly whilst users stuck in Windows will have to run Cygwin or something like that. Initially this tool will support bash and zsh. Other shells and environments might be added at a later stage.

nrun will attempt to find the closest package.json (hierarchically upwards) used by the current project and execute the named script found in the scripts section of that package.json.

The major reason for using it is the number of keystrokes that can be saved. When typing the same commands over and over again it can be annoying, even though you might use the up arrow in the terminal and repeat an earlier command.

It's also much easier to map your most used command to shorter ones that are easier to remember and faster to type.

So if you have to type something like
``` bash
# > npm run test:coverage:localhost
```
every time you want to run your tests. Then it would be easier to use nrun and type
``` bash
# > nrun test:coverage:localhost
```
And by using shortcuts in .nrun.json you might shorten this even more to
``` bash
# > nrun test
```
or something else that is easier to remember and faster to type.

And if you're not sure on which scripts there are available you can easily type
``` bash
# > nrun
```
and all the available scripts will be printed out in a list.

If you want to know what a certain script does you can write
``` bash
# > nrun -s test:coverage:localhost
```
and the command that this script runs will be printed out.

## Usage:
```bash
  nrun <scriptname> [args]          Run the script by name
  nrun -l                           Shows all available scripts
  nrun                              Shows all available scripts (same as the -l flag)
  nrun -p <project>                 Run the script in the specified project path
  nrun -s <scriptname>              Show the script that will be executed without running it
  nrun -h                           Shows help section
  nrun -lp                          Shows all available projects
  nrun -ap <project> <path>         Add a project to the list of projects
  nrun -rp <project>                Remove a project from the list of projects
  nrun -L ([license name]) (names)  Shows the licenses for the project
  nrun -V                           Shows all environment variables set by nrun
  nrun -e <command>                 Execute a command in the current shell
  nrun -x                           Execute a defined nrun script
  nrun -T                           Measure the time it takes to run a script
```

## Flags

### -l
Shows all available scripts. This is the same as just typing nrun. It will show all scripts in the current project.

The equivalent of this in npm is to type *npm run*.

### -p
Run the script in the specified project path. The project-name given is first checked against all registered projects in the global .nrun.json file. If no match is found then the project-name is assumed to be a path and the script will be run in that path.

### -s
Show the script that will be executed without running it. This is useful if you want to see what a script does before running it. 

### -h
Shows help section. 

### -lp
Shows all available projects defined in the global .nrun.json file.

### -ap
Add a project to the list of projects in the global .nrun.json file. The project-name given is first checked against all registered projects in the global .nrun.json file.

### -rp
Remove a project from the list of projects in the global .nrun.json file.

### -L
Shows the licenses for the project and its dependencies. If no arguments are given then the licenses for the project will be shown. If a license name is given then the licenses for the dependencies that have that license will be shown. If a list of names are given then the licenses for the dependencies that have one of those licenses will be shown.

### -V
Shows all environment variables set by nrun and their values. This is useful if you want to see what environment variables are set by nrun and what values they have.

### -e
Execute a command in the current directory.
This is useful if you want to execute a command.
The command will be executed in a shell and the output will be printed to the terminal.
The command will be executed in a subshell.

If the command requires flags then add -- before the command.

This flag can be used together with the -p flag making it possible to execute a command in a specific project.

### -ep
Execute a command in all defined projects.
This is useful if you want to execute a command all projects.
The command will be executed in a shell and the output will be printed to the terminal.
The command will be executed in a subshell.

If the command requires flags then add -- before the command.

### -x
Execute a defined nrun script.
This is useful if you want to execute multiple commands.
The commands don't have to be related to npm scripts.
Scripts will be executed in separate shells and the output will be printed to the terminal.

Flags and parameters can't be forwarded to the scripts.

### -lx
List all defined nrun scripts.

### -T
Measure the time it takes to run a script.

The time will be printed to the terminal when the script has finished.

| Time spent              | Format | Example |
|-------------------------|--------|---------|
| 1 minute ->             | Xm Ys  | 2m 13s  |
| 10 seconds - 1 minute   | X.Ys   | 36.1s   |
| 5 seconds - 10 seconds  | X.YYs  | 6.23s   |
| 1 second - 5 seconds    | X.YYYs | 2.152s  |
| 0.02 seconds - 1 second | Xms    | 657ms   |
| 20ms - 0.02 seconds     | Xus    | 1220us  |
| 0 -> 20ms               | Xns    | 16922ns |


## Installation
```bash
# > git clone git@github.com:codedeviate/nrun.git
# > cd nrun
# > go install
# > go build -o nrun main.go
```

### Dependencies
There is currently no dependencies for this tool (other than the need for [GoLang](https://go.dev/) to build it).
* ~~[gopkg.in/ini.v1](https://pkg.go.dev/gopkg.in/ini.v1)~~

## .nrun.json
Often used scriptnames can be mapped to other and shorter names in a file called .nrun.json.

This file should be placed in either the users home directory or in the same directory as the package.json.

The format is more or less a standard JSON-file. But there are some difference.
* ~~Key names can't contain colons and are therefore replaced with underscores.~~
* The section name is the full pathname of the directory that contains the package.json file.
* The section name must be a full path without any trailing slash.

Paths are defined under a key called "path" and environment variables are defined under a key called "env".

Projects are defined under a key called "projects" and the key name is the name of the project and the value is the path to the project.
Projects can only be defined in the global .nrun.json file.

The environment variables is not connected to the keys in the same directory but rather to the full script name.

### Example .nrun.json
```json
{
  "path": {
    "/Users/codedeviate/Development/nruntest": {
      "start": "start:localhost"
    }
  },
  "env": {
    "/Users/codedeviate/Development/nruntest": {
      "start:localhost": "PORT=3007"
    }
  },
  "projects": {
    "nruntest": "/Users/codedeviate/Development/nruntest"
  }
}
```

If you are in **/Users/codedeviate/Development/nruntest** and execute 
```bash
nrun start
```
then that will be the same as executing
```bash
PORT=3007 npm run start:localhost
```
which is saving some keystrokes.

### Global mapping and environment
Global section names are "\*" for mapping values and "\*" for environment values. These values will be overridden by values defined in the specific directory.

```json
{
  "path": {
    "/Users/codedeviate/Development/nruntest": {
      "start": "start:localhost"
    },
    "*": {
      "test": "test:coverage:localhost"
    }
  },
  "env": {
    "/Users/codedeviate/Development/nruntest": {
      "test:coverage:localhost": "PORT=3007"
    },
    "*": {
      "test:coverage:localhost": "PORT=3009"
    }
  },
  "projects": {
    "nruntest": "/Users/codedeviate/Development/nruntest"
  }
}
```
Now you can be in any project directory and type
``` bash
# > nrun test
```
which is the equivalent to
``` bash
# > PORT=3009 npm run test:coverage:localhost
```

If you are in a different path than you project then you can use the -p flag to specify the path to the project.
``` bash
# > nrun -p nruntest test
```
This is the same as
``` bash
# > cd /Users/codedeviate/Development/nruntest
# > nrun test
```


``` bash
# > nrun -p /Users/codedeviate/Development/nruntest test
```


## Different ways to use nrun
### You want to run a script that is located in another project
```bash
# > nrun -p proj1 test
```

### You have a project that is your main project that you want to use as your default project
Set the environment variable NRUNPROJECT to the name of the project.
```bash
# > nrun test
```
This will list all scripts in your default project.

If you have a default project defined but want to use nrun in the local directory you'll have to use the -p flag.
```bash
# > nrun -p .
```
This will list all scripts in your local directory.

If you want to execute a command in another project you can use the -e flag and the -p flag.
```bash
# > nrun -p proj1 -e ls
```
This will list all files in the proj1 directory.

Or a more realistic example.
```bash
# > nrun -p proj1 -e -- git commit -am "Some commit message"
```
This will commit all changes in the proj1 directory.

## Makefile
There are some predefined targets in the Makefile that can be used to build and install the tool.

```bash
# > make local
```
This will build the tool and place it in the current directory. It will also cross compile the tool for windows, linux and darwin. These files will be found in the bin directory.

```bash
# > make util
```
This will build the tool and move it to the ~/Utils directory. This is my preferred location for tools I use myself. So if you don't have a ~/Utils directory you'll either have to create it or skip using this target.

The cross compilation will be done for the following platforms:
* darwin/amd64
* darwin/arm64
* linux/amd64
* linux/arm64
* windows/amd64
* windows/arm64

This list is subject to change.

The cross compilations doesn't include any code signing or notarization. So if you want to use the cross compiled binaries on a mac you'll have to sign and notarize them yourself.

```bash
# > make go
```
This will build the tool and move it to the ~/go/bin directory. If you don't have a ~/go/bin directory or you have your go binaries somewhere else you'll either have to create it or skip using this target.

```bash
# > make all
```
This will create all the targets.

## Show licenses
To show the licenses for the project you can use the -L flag.

This will only work with NodeJS projects since the license is read from the package.json files in the node_modules directory.

Filtering can be done by adding the name, or a part of the name, of the license as a parameter.
You can use multiple parameters to filter the licenses.

The filter is case-insensitive and will match if the parameter is a substring of the license.

To show the name of all licenses for the project you can use the -L flag with "names" as the parameter
.
```bash
# > nrun -L
```
This will show the packages licenses for the project.

```bash
# > nrun -L MIT
```
This will show the packages that are under the MIT license.

```bash
# > nrun -L MIT ISC
```
This will show the packages that are under the MIT or ISC license.

```bash
# > nrun -L names
```
This will show the license names that are used by packages in the project.

```bash
# > nrun -L bs
```
The search is case-insensitive and searches for substrings so this will show the packages that are under the different BSD license.

Such as
* 0BSD
* BSD-2-Clause
* BSD-3-Clause

## Fallback to npm
If the script is not found in the package.json file then nrun will try a fallback to npm.

The following npm commands are passed along to npm:
>    access, adduser, audit, bin, bugs, cache, ci, completion,
>    config, dedupe, deprecate, diff, dist-tag, docs, doctor,
>    edit, exec, explain, explore, find-dupes, fund, get, help,
>    hook, init, install, install-ci-test, install-test, link,
>    ll, login, logout, ls, org, outdated, owner, pack, ping,
>    pkg, prefix, profile, prune, publish, rebuild, repo,
>    restart, root, run-script, search, set, set-script,
>    shrinkwrap, star, stars, start, stop, team, test, token,
>    uninstall, unpublish, unstar, update, version, view, whoami

Please note that not all of these commands are supported by nrun. This is because nrun is not a replacement for npm. It is a tool to make it easier to run scripts in your project.

Since nrun uses flags to specify the project and the script to run it is not possible to use flags such as the -h flag to get help for the npm commands.
