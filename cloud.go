package main

import (
	"cmp"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"slices"
	"strings"

	"github.com/manifoldco/promptui"
)

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
