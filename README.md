# nrun - The npm script runner

**Current version is v0.9.0**

nrun is a simple wrapper for **npm run** with some nice features. It is written in Go which I find easier to use when creating portable executable code.

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
And by using shortcuts in .nrun.ini you might shorten this even more to
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
  nrun <scriptname> [args]   Run the script by name
  nrun -l                    Shows all available scripts
  nrun                       Shows all available scripts (same as the -l flag)
  nrun -p <project>          Run the script in the specified projects path
  nrun -s <scriptname>       Show the script that will be executed without running it
  nrun -h                    Shows help section
```

## Installation
```bash
# > git clone git@github.com:codedeviate/nrun.git
# > cd nrun
# > go install
# > go build -o nrun main.go
```

### Dependencies
There is currently no dependencies for this tool.
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