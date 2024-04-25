package main

import (
	"bufio"
	"cmp"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/manifoldco/promptui"
)

type subscription struct {
	CloudName   string `json:"cloudName"`
	ID          string `json:"id"`
	IsDefault   bool   `json:"isDefault"`
	Name        string
	OrginalName string `json:"name"`
	State       string `json:"state"`
	TenantID    string `json:"tenantId"`
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

	for i, s := range subscriptions {
		subscriptions[i] = aliases.get(s)
	}

	slices.SortFunc(subscriptions,
		func(a, b subscription) int {
			return cmp.Compare(a.Name, b.Name)
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
{{ "Name:" | faint }}	{{if eq .Name .OrginalName}}{{ .Name }}{{else}}{{ .OrginalName }}{{end}}
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

func stringInput(desc string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(desc)
	outstr, _ := reader.ReadString('\n')
	return strings.Replace(outstr, "\n", "", -1)
}

func edit_subscription(s subscription) {
	project := stringInput("Project : ")
	name_str := stringInput("Name    : ")
	name := fmt.Sprintf("%s - %s", project, name_str)
	fmt.Printf("New Name   : %s\n", name)
	fmt.Print("\nOK (y/n)?")
	reader := bufio.NewReader(os.Stdin)
	char, _, _ := reader.ReadRune()
	switch char {
	case 'y':
		s.Name = name
		aliases.store(s)
	default:
		fmt.Println("Exit without change!")
	}
}
