package helper

import (
	"fmt"
	"os/exec"
	"strings"
)

func DummyCode() {
	fmt.Println("Dummy code that outputs the path to different shells if they are found")

	cmd := exec.Command("which", "sh")
	stdout, _ := cmd.Output()
	fmt.Println("Bourne Shell (sh)                 :", strings.Trim(string(stdout), " \n"))

	cmd = exec.Command("which", "bash")
	stdout, _ = cmd.Output()
	fmt.Println("GNU Bourne-Again Shell (bash)     :", strings.Trim(string(stdout), " \n"))

	cmd = exec.Command("which", "csh")
	stdout, _ = cmd.Output()
	fmt.Println("C Shell (csh)                     :", strings.Trim(string(stdout), " \n"))

	cmd = exec.Command("which", "ksh")
	stdout, _ = cmd.Output()
	fmt.Println("Korn Shell (ksh)                  :", strings.Trim(string(stdout), " \n"))

	cmd = exec.Command("which", "zsh")
	stdout, _ = cmd.Output()
	fmt.Println("Z Shell (zsh)                     :", strings.Trim(string(stdout), " \n"))

	cmd = exec.Command("which", "dash")
	stdout, _ = cmd.Output()
	fmt.Println("Debian Almquist Shell (dash)      :", strings.Trim(string(stdout), " \n"))

	cmd = exec.Command("which", "fish")
	stdout, _ = cmd.Output()
	fmt.Println("Friendly Interactive Shell (fish) :", strings.Trim(string(stdout), " \n"))

	cmd = exec.Command("which", "ash")
	stdout, _ = cmd.Output()
	fmt.Println("Almquist Shell (ash)              :", strings.Trim(string(stdout), " \n"))
}
