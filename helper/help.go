package helper

import "fmt"

func ShowHelp(version string) {
	fmt.Println("nrun - The npm script runner")
	fmt.Println("============================")
	fmt.Println("nrun will lookup the package.json used by the current project and execute the named script found in the scripts section of the package.json.")
	fmt.Println("")
	fmt.Println("Version: ", version)
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  nrun <script name> [args]         Run the script by name")
	fmt.Println("  nrun -n                           Pass through to npm. Send everything to npm and let it handle it.")
	fmt.Println("  nrun -i                           Show information about the current project")
	fmt.Println("  nrun -l                           Shows all available scripts")
	fmt.Println("  nrun                              Shows all available scripts (same as the -l flag)")
	fmt.Println("  nrun -s <script name>             Show the script without running it")
	fmt.Println("  nrun -h                           Shows this help")
	fmt.Println("  nrun -v                           Shows current version")
	fmt.Println("  nrun -pl                          List all projects from the config")
	fmt.Println("  nrun -pa <project name> <path>    Add a project to the config")
	fmt.Println("  nrun -pr <project name>           Remove a project from the config")
	fmt.Println("  nrun -L ([license name]) (names)  Shows all licenses of dependencies")
	fmt.Println("  nrun -V                           Shows all environment variables set by nrun")
	fmt.Println("  nrun -e <command>                 Execute a command")
	fmt.Println("  nrun -ep <command>                Execute a command in all projects")
	fmt.Println("  nrun -x <script>                  Execute a nrun script")
	fmt.Println("  nrun -xp <script>                 Execute a nrun script in all projects")
	fmt.Println("  nrun -T                           Measure the time it takes to execute the script")
	fmt.Println("For more information, see README.md")
}
