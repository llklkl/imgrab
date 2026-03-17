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

	optionStyle = lipgloss.NewStyle().
			Padding(0, 1).
			MarginRight(1)

	optionSelectedStyle = optionStyle.Copy().
				Foreground(lipgloss.Color("229")).
				Background(lipgloss.Color("63"))
)

const (
	actionDownloadOnly = iota
	actionImportDocker
)

type confirmModel struct {
	image       SelectedImage
	actionIndex int
	confirmed   bool
	back        bool
}

var actionList = []string{
	"Download Only",
	"Import to Docker",
}

func newConfirmModel() confirmModel {
	return confirmModel{
		actionIndex: actionImportDocker,
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
			if m.actionIndex > 0 {
				m.actionIndex--
			}
		case "right", "l":
			if m.actionIndex < len(actionList)-1 {
				m.actionIndex++
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

func (m confirmModel) contentView() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Confirm Action") + "\n\n")
	b.WriteString(fmt.Sprintf("Image: %s\n", highlightStyle.Render(m.image.Name)))
	b.WriteString(fmt.Sprintf("Tag: %s\n", highlightStyle.Render(m.image.Tag)))
	b.WriteString(fmt.Sprintf("Architecture: %s\n\n", highlightStyle.Render(m.image.Arch)))

	b.WriteString("Select Action (←/→ to switch):\n\n")
	for i, action := range actionList {
		if i == m.actionIndex {
			b.WriteString(optionSelectedStyle.Render(action))
		} else {
			b.WriteString(optionStyle.Render(action))
		}
	}

	b.WriteString("\n\n")
	b.WriteString("Press y/Enter to confirm, n/Esc to return\n")

	return b.String()
}

func (m confirmModel) View() string {
	return confirmStyle.Render(m.contentView())
}

func (m confirmModel) action() int {
	return m.actionIndex
}
