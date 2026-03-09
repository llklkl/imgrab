package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type archModel struct {
	image     SelectedImage
	archIndex int
	confirmed bool
	back      bool
}

var archList = []string{
	"linux/amd64",
	"linux/arm64",
	"linux/arm/v7",
	"linux/386",
}

func newArchModel() archModel {
	return archModel{
		archIndex: 0,
	}
}

func (m archModel) Init() tea.Cmd {
	return nil
}

func (m archModel) Update(msg tea.Msg) (archModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()

		switch key {
		case "left", "h":
			if m.archIndex > 0 {
				m.archIndex--
			}
		case "right", "l":
			if m.archIndex < len(archList)-1 {
				m.archIndex++
			}
		case "y", "Y", "enter":
			m.confirmed = true
		case "n", "N", "esc":
			m.back = true
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m archModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Select Architecture") + "\n\n")
	b.WriteString(fmt.Sprintf("Image: %s\n", highlightStyle.Render(m.image.Name)))
	b.WriteString(fmt.Sprintf("Tag: %s\n\n", highlightStyle.Render(m.image.Tag)))

	b.WriteString("Select Architecture (←/→ to switch):\n\n")
	for i, arch := range archList {
		if i == m.archIndex {
			b.WriteString(optionSelectedStyle.Render(arch))
		} else {
			b.WriteString(optionStyle.Render(arch))
		}
	}

	b.WriteString("\n\n")
	b.WriteString("Press y/Enter to confirm, n/Esc to return\n")

	return confirmStyle.Render(b.String())
}

func (m archModel) arch() string {
	return archList[m.archIndex]
}
