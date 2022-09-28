# nrun

# nrun - The npm script runner
nrun is a simple wrapper for **npm run** with some nice features

nrun will lookup the package.json used by the current project and execute the named script found in the scripts section of the package.json.

## Usage:
```bash
  nrun <scriptname> [args]   Run the script by name
  nrun -l                    Shows all available scripts
  nrun                       Shows all available scripts (same as the -l flag)
  nrun -s <scriptname>       Show the script without running it
  nrun -h                    Shows this help
```
## .nrun.ini
Often used scriptnames can be mapped to other and shorter names in a file called .nrun.ini.

This file should be placed in either the users home directory or in the same directory as the package.json.

The format is more or less a standard ini-file. But there are some difference.
* Key names can't contain colons and are therefore replaced with underscores.
* The section name is the full pathname of the directory that contains the package.json file.
* The section name must be a full path without any trailing slash.

Environment variables can be defined by adding "ENV:" as a prefix to the sections name.

These environment variables is not connected to the keys in the same directory but rather to the full script name.

Global section names are "\*" for mapping values and "ENV:\*" for environment values. These values will be overridden by values defined in the specific directory.

### Example .nrun.ini
```ini
[/Users/codedeviate/Development/nruntest]
start=start:localhost
[ENV:/Users/codedeviate/Development/nruntest]
start_localhost=PORT=3007
```

If you are in **/Users/codedeviate/Development/nruntest** and execute **nrun start** then that will be the same as executing **PORT=3007 nrun start:localhost** which is saving some keystrokes.
