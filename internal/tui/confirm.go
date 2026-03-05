package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	confirmStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1, 2).
			Width(60)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("63"))

	highlightStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212"))

	archOptionStyle = lipgloss.NewStyle().
			Padding(0, 1).
			MarginRight(1)

	archSelectedStyle = archOptionStyle.Copy().
				Foreground(lipgloss.Color("229")).
				Background(lipgloss.Color("63"))
)

type confirmModel struct {
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

func newConfirmModel() confirmModel {
	return confirmModel{
		archIndex: 0,
	}
}

func (m confirmModel) Init() tea.Cmd {
	return nil
}

func (m confirmModel) Update(msg tea.Msg) (confirmModel, tea.Cmd) {
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
		case "down", "j":
		case "up", "k":
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

func (m confirmModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Confirm Download") + "\n\n")
	b.WriteString(fmt.Sprintf("Image: %s\n", highlightStyle.Render(m.image.Name)))
	b.WriteString(fmt.Sprintf("Tag: %s\n\n", highlightStyle.Render(m.image.Tag)))

	b.WriteString("Select Architecture (←/→ to switch):\n\n")
	for i, arch := range archList {
		if i == m.archIndex {
			b.WriteString(archSelectedStyle.Render(arch))
		} else {
			b.WriteString(archOptionStyle.Render(arch))
		}
	}

	b.WriteString("\n\n")
	b.WriteString("Press y/Enter to confirm, n/Esc to return\n")

	return confirmStyle.Render(b.String())
}

func (m confirmModel) arch() string {
	return archList[m.archIndex]
}
