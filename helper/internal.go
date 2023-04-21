package helper

import (
	"fmt"
	probing "github.com/prometheus-community/pro-bing"
	"os/exec"
	"strings"
	"sync"
	"time"
)

func InternalCommands(packageJSON PackageJSON, script string, args []string, envs map[string]string, Version string) bool {
	if script == "nurse" {
		Nurse(packageJSON, script, args, envs, Version)
		return true
	}
	return false
}

func Nurse(packageJSON PackageJSON, script string, args []string, envs map[string]string, Version string) bool {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		// Ping npmjs.com to check if we have the latest version
		pinger, err := probing.NewPinger("npmjs.com")
		if err != nil {
			panic(err)
		}
		pinger.Interval = time.Millisecond * 100
		pinger.Count = 3
		pinger.Run()
		stats := pinger.Statistics()
		if stats.PacketsRecv > 0 {
			fmt.Println("Ping: You have a good connection to npmjs.com")
		} else {
			pinger.Interval = time.Millisecond * 1000
			pinger.Count = 3
			pinger.Run()
			stats := pinger.Statistics()
			if stats.PacketsRecv > 0 {
				fmt.Println("Ping: You have a good connection to npmjs.com (with longer pings)")
			} else {
				fmt.Println("Ping: You don't have a good connection to npmjs.com")
			}
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		// Check npm version
		cmd := exec.Command("npm", "--version")
		output, err := cmd.Output()
		if err != nil {
			fmt.Println("Error running npm --version", err)
		} else {
			fmt.Println("npm version:", strings.TrimSpace(string(output)))
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		// Check node version
		cmd := exec.Command("node", "--version")
		output, err := cmd.Output()
		if err != nil {
			fmt.Println("Error running node --version", err)
		} else {
			fmt.Println("node version:", strings.TrimSpace(string(output)))
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		// Check node version
		cmd := exec.Command("npm", "config", "get", "registry")
		output, err := cmd.Output()
		if err != nil {
			fmt.Println("Error running npm config get registry", err)
		} else {
			fmt.Println("Registry:", strings.TrimSpace(string(output)))
		}
		wg.Done()
	}()

	wg.Wait()
	return false
}
