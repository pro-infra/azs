package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"

	"github.com/manifoldco/promptui"
)

var version string

const GOARCH string = runtime.GOARCH
const GOOS string = runtime.GOOS

type subscription struct {
	CloudName string `json:"cloudName"`
	ID        string `json:"id"`
	IsDefault bool   `json:"isDefault"`
	Name      string `json:"name"`
	State     string `json:"state"`
	TenantID  string `json:"tenantId"`
}

func getSubscriptions() []subscription {
	subscriptions := []subscription{}
	out, err := exec.Command("bash", "-c", "az account list").Output()
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(out, &subscriptions)
	if err != nil {
		log.Fatal(err)
	}
	return subscriptions
}

func select_subscriptions(subscriptions []subscription) subscription {
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}?",
		Active:   "\U00002714 {{ .Name | blue }}",
		Inactive: "  {{ .Name | cyan }}",
		Selected: "\U00002714 {{ .Name | blue }}",
		Details: `
------------------------ Details -------------------------
{{ "Name:" | faint }}	{{ .Name }}
{{ "ID:" | faint }}	{{ .ID }}
{{ "Tenant ID:" | faint }}	{{ .TenantID }}
`,
	}

	searcher := func(input string, index int) bool {
		s := subscriptions[index]
		name := strings.Replace(strings.ToLower(s.Name), " ", "", -1)
		input = strings.Replace(strings.ToLower(input), " ", "", -1)

		return strings.Contains(name, input)
	}

	pos := 0

	for i, s := range subscriptions {
		if s.IsDefault {
			pos = i
			break
		}
	}

	prompt := promptui.Select{
		Label:     "Select Subscription",
		Items:     subscriptions,
		Templates: templates,
		CursorPos: pos,
		Size:      10,
		Searcher:  searcher,
	}
	i, _, err := prompt.Run()
	if err != nil {
		log.Fatalf("Prompt failed %v\n", err)
	}
	return subscriptions[i]
}

func set_subscription(s subscription) {
	cmd := fmt.Sprintf("az account set --subscription %s", s.ID)
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		log.Fatal(err)
	}
	if len(out) > 0 {
		fmt.Println(string(out))
	}
}

func main() {
	showVersion := false
	update := false
	dryupd := false
	flag.BoolVar(&showVersion, "v", false, "Show version")
	flag.BoolVar(&update, "u", false, "Update azs")
	flag.BoolVar(&dryupd, "U", false, "Dry-run update azs")
	flag.Parse()

	switch {
	case showVersion:
		fmt.Println(version, GOOS, GOARCH)
		return
	case update || dryupd:
		updateazs(dryupd)
		return
	}

	subscriptions := getSubscriptions()
	s := select_subscriptions(subscriptions)
	set_subscription(s)
}
