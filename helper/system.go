package helper

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
)

func ReadMemoryStats() Memory {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return Memory{}
	}
	defer file.Close()
	bufio.NewScanner(file)
	scanner := bufio.NewScanner(file)
	res := Memory{}
	for scanner.Scan() {
		key, value := parseLine(scanner.Text())
		switch key {
		case "MemTotal":
			res.MemTotal = value
		case "MemFree":
			res.MemFree = value
		case "MemAvailable":
			res.MemAvailable = value
		}
	}
	return res
}

func GetVersionFromExecutable(executable string, args []string) string {
	cmd := exec.Command(executable, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out.String())
}

func SystemInfo(Version string) {
	fmt.Println("System information:")
	fmt.Println("  OS:", runtime.GOOS)
	if kernelVersion := GetVersionFromExecutable("uname", []string{"-r"}); Version != "" {
		fmt.Println("  Kernel:", kernelVersion)
	}
	fmt.Println("  Architecture:", runtime.GOARCH)

	memory := ReadMemoryStats()
	if memory.MemTotal > 0 {
		fmt.Println("  Total memory:", memory.MemTotal/1024, "MB")
	}
	if memory.MemFree > 0 {
		fmt.Println("  Free memory:", memory.MemFree/1024, "MB")
	}
	if memory.MemAvailable > 0 {
		fmt.Println("  Available memory:", memory.MemAvailable/1024, "MB")
	}

	var version string
	fmt.Println("  Versions of installed tools:")
	if version = GetVersionFromExecutable("git", []string{"--version"}); version != "" {
		cleanVersion := regexp.MustCompile(`git\s+version\s+([0-9\.]+)`).FindStringSubmatch(version)
		fmt.Println("    Git:", cleanVersion[1])
	}
	if version = GetVersionFromExecutable("nrun", []string{"-v"}); version != "" {
		fmt.Println("    nrun:", version)
	}
	if version = GetVersionFromExecutable("node", []string{"-v"}); version != "" {
		fmt.Println("    Node:", version)
	}
	if version = GetVersionFromExecutable("npm", []string{"-v"}); version != "" {
		fmt.Println("    NPM:", version)
	}
	if version = GetVersionFromExecutable("go", []string{"version"}); version != "" {
		cleanVersion := regexp.MustCompile(`go\s+version\s+go([0-9\.]+)`).FindStringSubmatch(version)
		fmt.Println("    Go:", cleanVersion[1])
	}
	if version = GetVersionFromExecutable("php", []string{"-v"}); version != "" {
		cleanVersion := regexp.MustCompile(`PHP\s+([0-9\.]+)\s+`).FindStringSubmatch(version)
		fmt.Println("    PHP:", cleanVersion[1])
	}
	if version = GetVersionFromExecutable("python3", []string{"--version"}); version != "" {
		cleanVersion := regexp.MustCompile(`Python\s+([0-9\.]+)`).FindStringSubmatch(version)
		fmt.Println("    Python3:", cleanVersion[1])
	}
	if version = GetVersionFromExecutable("ruby", []string{"-v"}); version != "" {
		cleanVersion := regexp.MustCompile(`ruby\s+([0-9\.]+)`).FindStringSubmatch(version)
		fmt.Println("    Ruby:", cleanVersion[1])
	}
	if version = GetVersionFromExecutable("gcc", []string{"--version"}); version != "" {
		cleanVersion := regexp.MustCompile(`clang\s+version\s+([0-9\.]+)`).FindStringSubmatch(version)
		fmt.Println("    GCC:", cleanVersion[1])
	}
	if version = GetVersionFromExecutable("make", []string{"-v"}); version != "" {
		cleanVersion := regexp.MustCompile(`Make\s+([0-9\.]+)`).FindStringSubmatch(version)
		fmt.Println("    Make:", cleanVersion[1])
	}
	if version = GetVersionFromExecutable("ldd", []string{"--version"}); version != "" {
		fmt.Println("    LDD:", version)
	}
	if version = GetVersionFromExecutable("zig", []string{"version"}); version != "" {
		fmt.Println("    Zig:", version)
	}
	if version = GetVersionFromExecutable("bun", []string{"--version"}); version != "" {
		cleanVersion := regexp.MustCompile(`bun\s+([0-9\.]+)`).FindStringSubmatch(version)
		fmt.Println("    Bun:", cleanVersion[1])
	}
	if version = GetVersionFromExecutable("deno", []string{"--version"}); version != "" {
		cleanVersion := regexp.MustCompile(`deno\s+([0-9\.]+)`).FindStringSubmatch(version)
		fmt.Println("    Deno:", cleanVersion[1])
	}
	if version = GetVersionFromExecutable("rustc", []string{"--version"}); version != "" {
		cleanVersion := regexp.MustCompile(`rustc\s+([0-9\.]+)`).FindStringSubmatch(version)
		fmt.Println("    Rust:", cleanVersion[1])
	}
	if version = GetVersionFromExecutable("cargo", []string{"--version"}); version != "" {
		cleanVersion := regexp.MustCompile(`cargo\s+([0-9\.]+)`).FindStringSubmatch(version)
		fmt.Println("    Cargo:", cleanVersion[1])
	}
	fmt.Println("\nVersion information brought to you by nrun.")
}
