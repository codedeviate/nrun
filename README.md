# nrun - The npm script runner

**Current version is v0.19.0**

nrun is a utility to make **npm run** a bit easier, and it has some nice features. It is written in Go which I find easier to use when creating portable executable code.

There is no requirement to have **npm** installed to use nrun. The scripts in package.json are parsed and executed directly.

Even though the goal is to make it portable, nrun will still need a shell to run. So Linux users and Mac users can probably run it smoothly whilst users stuck in Windows will have to run Cygwin or something like that. Initially this tool will support bash and zsh. Other shells and environments might be added at a later stage.

nrun will attempt to find the closest package.json (hierarchically upwards) used by the current project and execute the named script found in the scripts section of that package.json.

The major reason for using it is the number of keystrokes that can be saved. When typing the same commands over and over again it can be annoying, even though you might use the up arrow in the terminal and repeat an earlier command.

It's also much easier to map your most used command to shorter ones that are easier to remember and faster to type.

So if you have to type something like
```console
foo@bar:~$ npm run test:coverage:localhost
```
every time you want to run your tests. Then it would be easier to use nrun and type
```console
foo@bar:~$ nrun test:coverage:localhost
```
And by using shortcuts in .nrun.json you might shorten this even more to
```console
foo@bar:~$ nrun test
```
or something else that is easier to remember and faster to type.

And if you're not sure on which scripts there are available you can easily type
```console
foo@bar:~$ nrun
```
and all the available scripts will be printed out in a list.

If you want to know what a certain script does you can write
```console
foo@bar:~$ nrun -s test:coverage:localhost
```
and the command that this script runs will be printed out.

## Usage:
```console
  nrun <scriptname> [args]               Run the script by name
  nrun -a [alias]                        Execute aliases defined in the global .nrun.json file (separate multiple aliases with a space)
  nrun -l                                Shows all available scripts
  nrun                                   Shows all available scripts (same as the -l flag)
  nrun -p <project>                      Run the script in the specified project path
  nrun -s <scriptname>                   Show the script that will be executed without running it
  nrun -h                                Shows help section
  nrun -pl                               Shows all available projects
  nrun -pa <project> <path>              Add a project to the list of projects
  nrun -pr <project>                     Remove a project from the list of projects
  nrun -L ([license name]) (names)       Shows the licenses for the project
  nrun -V                                Shows all environment variables set by nrun
  nrun -e <command>                      Execute a command in the current project
  nrun -ep <command>                     Execute a command in all defined projects
  nrun -x  <script>                      Execute a defined nrun script in the current project
  nrun -xl                               List all defined nrun scripts and the commands they run
  nrun -xm <script> [<script>...]        Execute multiple defined nrun scripts
  nrun -xp <script>                      Execute a defined nrun script in all defined projects
  nrun -xat <token>                      Add the X_AUTH_TOKEN environment variable to the script environment
  nrun -T                                Measure the time it takes to run a script
  nrun -w <url>                          Get the content of the url and print it to the terminal
  nrun -wt <template>                    Get the content of the url and its parameters defined in the template and print it to the terminal
  nrun -wi                               Get the content of the url and print information about the response and the headers
  nrun -wh                               Get the content of the url and print the headers
  nrun -wnb                              Get the content of the url but don't print the body
  nrun -who                              Get the content of the url but only print the headers
  nrun -jwt <token>                      Read a token and parse it as a JWT token and print the header and payload
  nrun -jwt-sign <secret> < <json data>  Read payload from STDIN and sign it with the secret and print the token 
```
*Please note that the examples of the listed flags may require a combination with other flags and might not work stand-alone.*

## Flags

### -a
Execute aliases defined in the global .nrun.json file.

Several aliases can be executed after each other by separating them with a space.

```console
foo@bar:~$ nrun -a master pull
```
This will execute the *master* alias and then the *pull* alias.
If the *master* alias is defined as *git checkout master* and the *pull* alias is defined as *git pull* then the above command will be equivalent to the following command:
```console
foo@bar:~$ git checkout master && git pull
```

But aliases can also be combined with other flags. For example:
```console
foo@bar:~$ nrun -p project1 -a master pull
```
Which will execute the *master* alias in the *project1* project and then the *pull* alias in the *project1* project.

Please note that this flag is a boolean flag and does not take any arguments. The aliases to be executed are given as arguments. So the aliases doesn't have to be directly after the -a flag.

### -l
Shows all available scripts. This is the same as just typing nrun. It will show all scripts in the current project.

The equivalent of this in npm is to type *npm run*.

### -p
Run the script in the specified project path. The project-name given is first checked against all registered projects in the global .nrun.json file. If no match is found then the project-name is assumed to be a path and the script will be run in that path.

### -s
Show the script that will be executed without running it. This is useful if you want to see what a script does before running it. 

### -h
Shows help section. 

### -pl
Shows all available projects defined in the global .nrun.json file.

### -ap
Add a project to the list of projects in the global .nrun.json file. The project-name given is first checked against all registered projects in the global .nrun.json file.

### -pr
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

### -xm
Execute multiple defined nrun scripts in parallel.

This is useful if you want to execute multiple commands to be run in parallel. An example of this is if you want to run spin up both a NodeJS backend and a React frontend at the same time with one command.

```console
foo@bar:~$ nrun -xm start:backend start:frontend
```

### -xp
Execute a defined nrun script in all defined projects.

This is useful if you want to execute multiple commands in all projects.

### -xl
List all defined nrun scripts.

### -xat
Add the X_AUTH_TOKEN environment variable to the script.

The string given as a parameter can either be an index to a token in the .nrun.json file or a string that will be used as the token.

(This flag can also be used with the -w flag.)

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


### -jwt &lt;token&gt;
Reads the token and parses it.

The function prints out the header and payload of the token as prettified JSON.

```console
foo@bar:~$ nrun -jwt eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiTmlzc2UgSHVsdCIsInVzZXJuYW1lIjoibmlzc2UiLCJlbWFpbCI6Im5pc3NlQGV4YW1wbGUuY29tIn0.VA16nPv9fdcVArzz2xEHAsyHVbPwSIVqE7f9iSJ119A
Header: {
    "alg": "HS256",
    "typ": "JWT"
}
Payload: {
    "email": "nisse@example.com",
    "name": "Nisse Hult",
    "username": "nisse"
}
```
### -jwt-sign &lt;secret&gt; &lt; &lt;payload&gt;
Read the payload from STDIN and sign it with the given secret.

This also add the header to the token.

The token is signed with the HS256 algorithm.
```console
foo@bar:~$ cat payload.json
{
    "email": "nisse@example.com",
    "name": "Nisse Hult",
    "username": "nisse"
}
foo@bar:~$ nrun -jwt-sign mysecret < payload.json
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyAgICAiZW1haWwiOiAibmlzc2VAZXhhbXBsZS5jb20iLCAibmFtZSI6ICJOaXNzZSBIdWx0IiwgInVzZXJuYW1lIjogIm5pc3NlIn0.7cIQmQs9Sgf2z0kmU2460BFlfsyQqGpwsegs00_w3b8
```

1. **Note:** The payload must be valid JSON.
2. **Note:** The difference in the examples depends on the formatting of the payload (different number of whitespace characters).

## Pre- and post-scripts
It is possible to define scripts that will be run before and after the main script.
This is useful if you want to do some setup before running the main script and some cleanup after the main script has finished.

The naming convention for pre- and post-scripts is straightforward. Add "pre" before the name of the main script to make it execute before the main script and add "post" before the name of the main script to make it execute after the main script.

For example, if you have a script called "test" then you can define a script called "pretest" that will be run before the "test" script and a script called "posttest" that will be run after the "test" script.

```json
{
  "scripts": {
    "test": "echo \"Running tests\"",
    "pretest": "echo \"Setting up test environment\"",
    "posttest": "echo \"Cleaning up test environment\""
  }
}
```

The default behavior in npm is that it will look for a pre-script and a post-script for the script that is run.

In nrun however there is a slight difference.
When the pre-script executes it will also look for a pre- and post-script which makes it possible to add a pre-pre-script if needed.
Please note that this is not the behavior in npm.



## Installation
```console
foo@bar:~$ git clone git@github.com:codedeviate/nrun.git
foo@bar:~$ cd nrun
foo@bar:~$ go install
foo@bar:~$ go build -o nrun main.go
```

### Dependencies
There is currently one dependency for this tool (other than the need for [GoLang](https://go.dev/) to build it).
* [github.com/google/shlex](https://github.com/google/shlex)

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

Scripts are defined under a key called "scripts" and the key name is the name of the script and the value is the command to execute.
Scripts can only be defined in the global .nrun.json file.

The environment variables is not connected to the keys in the same directory but rather to the full script name.

### Example of global .nrun.json
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
  },
  "alias": {
    "pull": "git pull",
    "merge": "git merge origin/master",
    "checkout": "git checkout",
    "master": "git checkout master"
  },
  "scripts": {
    "status": [
      "@@hasfile: .git",
      "echo -n \"Branch in \" && echo -n $NRUN_CURRENT_PATH && echo -n \" is in branch \" && git rev-parse --abbrev-ref HEAD",
      "git status --short",
      "echo \"\""
    ],
    "test": [
      "echo \"Running tests\"",
      "echo \"Running tests\"",
      "echo \"Running tests\""
    ],
    "start:frontend":[
      "nrun -x -p frontend start"
    ],
    "start:backend": [
      "nrun -x -p backend start"
    ]
  },
  "xauthtokens":{
    "0": "1234567890",
    "1": "0987654321"
  },
  "webget": {
    "test": {
      "method": "GET",
      "url": "http://localhost:3007/invoices?limit=100",
      "format": "auto",
      "body": "",
      "headers": {
        "Content-Type": "application/json"
      },
      "XAuthToken": "14576cg9dg64752cv7cvb92"
    }
  },
  "personalflags": {
    "morning": [
      "say --voice Daniel \"Good morning! The status of your computer is as follows.\" &",
      "nrun -pl",
      "nrun -xp status"
    ]
  }
}
```
### Example of local .nrun.json
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
  }
}
```

If you are in **/Users/codedeviate/Development/nruntest** and execute 
```console
nrun start
```
then that will be the same as executing
```console
PORT=3007 npm run start:localhost
```
which is saving some keystrokes.

### Global mapping and environment
Global section names are "\*" for mapping values and "\*" for environment values. These values will be overridden by values defined in the specific directory.

If a value is already defined in the environment for the current shell then that value will be used instead of the value defined in the .nrun.json file.

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
  },
  "scripts": {
    "test": [
      "echo \"Running tests\"",
      "echo \"Running tests\"",
      "echo \"Running tests\""
    ]
  }
}
```
Now you can be in any project directory and type
```console
foo@bar:~$ nrun test
```
which is the equivalent to
```console
foo@bar:~$ PORT=3009 npm run test:coverage:localhost
```

If you are in a different path than you project then you can use the -p flag to specify the path to the project.
```console
foo@bar:~$ nrun -p nruntest test
```
This is the same as
```console
foo@bar:~$ cd /Users/codedeviate/Development/nruntest
foo@bar:~$ nrun test
```


```console
foo@bar:~$ nrun -p /Users/codedeviate/Development/nruntest test
```


## Different ways to use nrun
### You want to run a script that is located in another project
```console
foo@bar:~$ nrun -p proj1 test
```

### You have a project that is your main project that you want to use as your default project
Set the environment variable NRUNPROJECT to the name of the project.
```console
foo@bar:~$ nrun test
```
This will list all scripts in your default project.

If you have a default project defined but want to use nrun in the local directory you'll have to use the -p flag.
```console
foo@bar:~$ nrun -p .
```
This will list all scripts in your local directory.

If you want to execute a command in another project you can use the -e flag and the -p flag.
```console
foo@bar:~$ nrun -p proj1 -e ls
```
This will list all files in the proj1 directory.

Or a more realistic example.
```console
foo@bar:~$ nrun -p proj1 -e -- git commit -am "Some commit message"
```
This will commit all changes in the proj1 directory.


## nrun scripts

In the .nrun.json file you can define scripts that can be executed with nrun.

They are defined under the key "scripts" and the key name is the name of the script and the value is an array with the commands to execute.

If the value in the command array begins with "@@" then this is regarded an internal command.

### Environment variables

#### NRUN_CURRENT_PATH
This environment variable will be set to the current path.

```json
{
  "scripts": {
    "test": [
      "echo $NRUN_CURRENT_PATH",
      "echo $NRUN_CURRENT_SCRIPT"
    ]
  }
}
```
This will print the current path and the current script name.

#### NRUN_CURRENT_SCRIPT
This environment variable will be set to the name of the script that is executed.

#### NRUN_CURRENT_SCRIPT_CODE
This environment variable will be set to the script that is executed.

This might be somewhat useless since it only contains the current script value that is executed and not the entire code array.

### Internal commands
Internal commands are commands that are executed by nrun and not by the shell.

The definition of the internal commands are as follows.
```json
{
  "scripts": {
    "test": [
      "@@internalcommand: argument1,argument2"
    ]
  }
}
```

Commands that returns a boolean value that either lets the script continue or not can be negated by adding an exclamation mark in front of the command.
```json
{
  "scripts": {
    "test": [
      "@@!internalcommand: argument1,argument2"
    ]
  }
}
```

#### hasfile, hasfiles
These commands will both check if a file exists in the current directory.
If the file exists then the command will continue.
If the file doesn't exist then the command will return without completing the rest of the command array.

Multiple filenames can be specified by separating them with a comma.

Filenames can be relative to the projects path or absolute.

The difference between **hasfile** and **hasfiles** is that **hasfile** will continue if one of the files exists while **hasfiles** will only continue if all the files exists.

```json
{
  "scripts": {
    "test": [
      "@@hasfile: package.json",
      "npm run test"
    ]
  }
}
```

```json
{
  "scripts": {
    "test": [
      "@@hasfiles: package.json,package-lock.json",
      "npm run test"
    ]
  }
}
```

#### cd
This command will change the current directory to the specified directory.

#### set
Set an environment variable

#### env
This is the same as **set**

#### unset
Unset an environment variable

#### unenv
This is the same as @@unset

#### echo
Print a message to the stdout

#### @@isfile
Check if a file exists in the current directory.

#### isdir
Check if a directory exists in the current directory.

## Doing web requests with nrun
nrun has a built-in web request function that can be used to do web requests.

Web requests can be done with the flag **-w**

```console
foo@bar:~$ nrun -w https://www.google.com
```

### -w flag
Perform a web request.

### -wi flag (Information)
Print the information and headers for a web request.

```console
foo@bar:~$ nrun -w -wi https://www.google.com
URL: https://www.google.com
Method: GET
Status: 200 OK
Size: 14377 bytes
Time: 253 ms
X-Auth-Token:

Headers:
  X-Frame-Options: SAMEORIGIN
  Content-Type: text/html; charset=ISO-8859-1
  P3p: CP="This is not a P3P policy! See g.co/p3phelp for more info."
  Server: gws
  Alt-Svc: h3=":443"; ma=2592000,h3-29=":443"; ma=2592000
  Expires: -1
  Set-Cookie: SOCS=CAAaBgiA75CfBg; expires=Sun, 10-Mar-2024 19:37:54 GMT; path=/; domain=.google.com; Secure; SameSite=lax, AEC=ARSKqsI2T2AbLdMVdP0j8HonQZPuoyY3UGeWnrcs3c5UBboXdqhUtVH6ayA; expires=Tue, 08-Aug-2023 19:37:54 GMT; path=/; domain=.google.com; Secure; HttpOnly; SameSite=lax, __Secure-ENID=10.SE=Kg5dkXE-AXZ7Jj_Ce2kPXxVnOlX-7F05NDWN5nzXUo3EFKPHqAuhWbaFYCpFHY5a_r1Yu8QeGr8iBacHGyD-Rntknn7Sj-jG_R2M5Y0kPjjwGQAGuWmIsf6UYsWUcHUn8feDYyfBpULa-mJ-cCnUE-WJrZGYcugaAmXTUfQUo4E; expires=Mon, 11-Mar-2024 11:56:12 GMT; path=/; domain=.google.com; Secure; HttpOnly; SameSite=lax, CONSENT=PENDING+682; expires=Sat, 08-Feb-2025 19:37:54 GMT; path=/; domain=.google.com; Secure
  Cache-Control: private, max-age=0
  X-Xss-Protection: 0
  Date: Thu, 09 Feb 2023 19:37:54 GMT
Cookies:
  SOCS: CAAaBgiA75CfBg
  AEC: ARSKqsI2T2AbLdMVdP0j8HonQZPuoyY3UGeWnrcs3c5UBboXdqhUtVH6ayA
  __Secure-ENID: 10.SE=Kg5dkXE-AXZ7Jj_Ce2kPXxVnOlX-7F05NDWN5nzXUo3EFKPHqAuhWbaFYCpFHY5a_r1Yu8QeGr8iBacHGyD-Rntknn7Sj-jG_R2M5Y0kPjjwGQAGuWmIsf6UYsWUcHUn8feDYyfBpULa-mJ-cCnUE-WJrZGYcugaAmXTUfQUo4E
  CONSENT: PENDING+682
```
### -wt flag (Template)

### -wh flag (Headers)

### -who flag (Headers only)

### -wm flag (Method)

### -wf flag (Format)

### -wnb flag (No body)

### -xat flag (X-AUTH-TOKEN)

## Makefile
There are some predefined targets in the Makefile that can be used to build and install the tool.

```console
foo@bar:~$ make local
```
This will build the tool and place it in the current directory. It will also cross compile the tool for windows, linux and darwin. These files will be found in the bin directory.

```console
foo@bar:~$ make util
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

```console
foo@bar:~$ make go
```
This will build the tool and move it to the ~/go/bin directory. If you don't have a ~/go/bin directory or you have your go binaries somewhere else you'll either have to create it or skip using this target.

```console
foo@bar:~$ make all
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
```console
foo@bar:~$ nrun -L
```
This will show the packages licenses for the project.

```console
foo@bar:~$ nrun -L MIT
```
This will show the packages that are under the MIT license.

```console
foo@bar:~$ nrun -L MIT ISC
```
This will show the packages that are under the MIT or ISC license.

```console
foo@bar:~$ nrun -L names
```
This will show the license names that are used by packages in the project.

```console
foo@bar:~$ nrun -L bs
```
The search is case-insensitive and searches for substrings so this will show the packages that are under the different BSD license.

Such as
* 0BSD
* BSD-2-Clause
* BSD-3-Clause

## Personalized flags
You can add your own flags to the tool by adding a section called personalflags in the global .nrun.json file.

The section should be a JSON object where the key is the name of the flag and the value is an array of strings.
```json
{
  "projects": {
    "project1": "/Users/foobar/Development/project1",
    "project2": "/Users/foobar/Development/project2",
    "project3": "/Users/foobar/Development/project3",
    "project4": "/Users/foobar/Development/project4",
    "project5": "/Users/foobar/Development/project5",
    "project6": "/Users/foobar/Development/project6",
    "project7": "/Users/foobar/Development/project7",
    "project8": "/Users/foobar/Development/project8",
    "project9": "/Users/foobar/Development/project9"
  },
  "scripts": {
    "status": [
      "@@hasfile: .git",
      "echo -n \"Branch in \" && echo -n $NRUN_CURRENT_PATH && echo -n \" is in branch \" && git rev-parse --abbrev-ref HEAD",
      "git status --short",
      "echo \"\""
    ]
  },
  "personalflags": {
    "morning": [
      "say --voice Daniel \"Good morning! The status of your computer is as follows.\" &",
      "nrun -pl",
      "nrun -xp status"
    ]
  }
}
```
The flag **morning** will now be available in the tool.
```console
foo@bar:~$ nrun -morning
```
And when run it will execute the commands in the array. The commands are executed in the order they are in the array. So the first the voice saying "Good morning" will be played and then the commands will be executed. In this particular example the say commands will be executed in the background (due to the ampersand at the end of the line) and then all projects will be listed followed by a status for all registered projects that run git.

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
