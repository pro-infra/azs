package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/manifoldco/promptui"
)

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
		Active:   "\U00002714 {{ .Name | cyan }}",
		Inactive: "  {{ .Name | cyan }}",
		Selected: "\U00002714 {{ .Name | red | cyan }}",
		Details: `
--------- Subscription ----------
{{ "Name:" | faint }}	{{ .Name }}
{{ "ID:" | faint }}	{{ .ID }}`,
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
	subscriptions := getSubscriptions()
	s := select_subscriptions(subscriptions)
	fmt.Printf("Select NAME: %s\n", s.Name)
	set_subscription(s)
}
