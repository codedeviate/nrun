package helper

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type Joke struct {
	Error    bool   `json:"error"`
	Category string `json:"category"`
	Type     string `json:"type"`
	Joke     string `json:"joke"`
	Flags    struct {
		Nsfw      bool `json:"nsfw"`
		Religious bool `json:"religious"`
		Political bool `json:"political"`
		Racist    bool `json:"racist"`
		Sexist    bool `json:"sexist"`
		Explicit  bool `json:"explicit"`
	} `json:"flags"`
	Id   int    `json:"id"`
	Safe bool   `json:"safe"`
	Lang string `json:"lang"`
}

func TellAJoke() {
	// Call https://v2.jokeapi.dev/joke/Any?type=single to get a random joke
	// and print it to the console
	cmd := exec.Command("curl", "-s", "https://v2.jokeapi.dev/joke/Any?type=single")
	stdout, _ := cmd.Output()
	var joke Joke
	_ = json.Unmarshal(stdout, &joke)
	fmt.Println(joke.Joke)
	Notify(joke.Joke)
}

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
