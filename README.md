# nrun - The npm script runner
nrun is a simple wrapper for **npm run** with some nice features. It is written in Go which I find easier to use when creating portable executable code.

Even though the goal is to make it portable nrun will need a shell to run. So Linux users and Mac users can probably run it smoothly whilst users stuck in Windows will have to run Cygwin or something like that. Initially this tool will support bash and zsh. Other shells and environments might be added at a later stage.

nrun will lookup the closest package.json used by the current project and execute the named script found in the scripts section of that package.json.

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
There is currently one dependency
* [gopkg.in/ini.v1](https://pkg.go.dev/gopkg.in/ini.v1)

## .nrun.ini
Often used scriptnames can be mapped to other and shorter names in a file called .nrun.ini.

This file should be placed in either the users home directory or in the same directory as the package.json.

The format is more or less a standard ini-file. But there are some difference.
* Key names can't contain colons and are therefore replaced with underscores.
* The section name is the full pathname of the directory that contains the package.json file.
* The section name must be a full path without any trailing slash.

Environment variables can be defined by adding "ENV:" as a prefix to the sections name.

These environment variables is not connected to the keys in the same directory but rather to the full script name.

### Example .nrun.ini
```ini
[/Users/codedeviate/Development/nruntest]
start=start:localhost

[ENV:/Users/codedeviate/Development/nruntest]
start_localhost=PORT=3007
```

If you are in **/Users/codedeviate/Development/nruntest** and execute 
```bash
nrun start
```
then that will be the same as executing
```bash
PORT=3007 nrun start:localhost
```
which is saving some keystrokes.

### Global mapping and environment
Global section names are "\*" for mapping values and "ENV:\*" for environment values. These values will be overridden by values defined in the specific directory.

```ini
[*]
test=test:coverage:localhost

[ENV:*]
test_coverage_localhost=PORT=3007

[/Users/codedeviate/Development/nruntest]
start=start:localhost

[ENV:/Users/codedeviate/Development/nruntest]
start_localhost=PORT=3007
```
Now you can be in any project directory and type
``` bash
# > nrun test
```
which is the equivalent to
``` bash
# > PORT=3007 npm run test:coverage:localhost
```
