package main

import (
	"cmp"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"slices"
	"strings"

	"github.com/manifoldco/promptui"
)

var version string

const GOARCH string = runtime.GOARCH
const GOOS string = runtime.GOOS

type cloud struct {
	IsActive bool   `json:"isActive"`
	Name     string `json:"name"`
}

func getClouds() []cloud {
	clouds := []cloud{}
	out, err := exec.Command("bash", "-c", "az cloud list").Output()
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(out, &clouds)
	if err != nil {
		log.Fatal(err)
	}
	slices.SortFunc(clouds,
		func(a, b cloud) int {
			return cmp.Compare(b.Name, a.Name)
		},
	)
	return clouds
}

func select_clouds(clouds []cloud) cloud {
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}?",
		Active:   "\U00002714 {{ .Name | blue }}",
		Inactive: "  {{ .Name | cyan }}",
		Selected: "\U00002714 {{ .Name | blue }}",
		Details: `
------------------------ Details -------------------------
{{ "Name:" | faint }}	{{ .Name }}
`,
	}

	searcher := func(input string, index int) bool {
		c := clouds[index]
		name := strings.Replace(strings.ToLower(c.Name), " ", "", -1)
		input = strings.Replace(strings.ToLower(input), " ", "", -1)

		return strings.Contains(name, input)
	}

	pos := 0

	for i, c := range clouds {
		if c.IsActive {
			pos = i
			break
		}
	}

	prompt := promptui.Select{
		Label:     "Select Cloud",
		Items:     clouds,
		Templates: templates,
		CursorPos: pos,
		Size:      10,
		Searcher:  searcher,
	}
	i, _, err := prompt.Run()
	if err != nil {
		log.Fatalf("Prompt failed %v\n", err)
	}
	return clouds[i]
}

func set_cloud(c cloud) {
	cmd := fmt.Sprintf("az cloud set --name  %s", c.Name)
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		log.Fatal(err)
	}
	if len(out) > 0 {
		fmt.Println(string(out))
	}
}

type subscription struct {
	CloudName    string `json:"cloudName"`
	ID           string `json:"id"`
	IsDefault    bool   `json:"isDefault"`
	Name         string `json:"name"`
	OriginalName string
	State        string `json:"state"`
	TenantID     string `json:"tenantId"`
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
	slices.SortFunc(subscriptions,
		func(a, b subscription) int {
			return cmp.Compare(fmt.Sprintf("%s-%s", b.TenantID, b.Name), fmt.Sprintf("%s-%s", a.TenantID, a.Name))
		},
	)
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
	switchCloud := false
	editSubscription := false
	flag.BoolVar(&showVersion, "v", false, "Show version")
	flag.BoolVar(&update, "u", false, "Update azs")
	flag.BoolVar(&dryupd, "U", false, "Dry-run update azs")
	flag.BoolVar(&switchCloud, "c", false, "Switch Cloud")
	flag.BoolVar(&editSubscription, "e", false, "Edit Subscription")
	flag.Parse()

	switch {
	case switchCloud:
		clouds := getClouds()
		c := select_clouds(clouds)
		set_cloud(c)
		return
	case editSubscription:
		subscriptions := getSubscriptions()
		s := select_subscriptions(subscriptions)
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
