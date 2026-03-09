package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/llklkl/imgrab/internal/registry"
)

type tagItem struct {
	name string
}

func (i tagItem) Title() string       { return i.name }
func (i tagItem) Description() string { return "" }
func (i tagItem) FilterValue() string { return i.name }

type tagsModel struct {
	repository string
	list       list.Model
	loading    bool
	selected   string
	back       bool
	err        error
}

func newTagsModel() tagsModel {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Image Tags"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.SetShowPagination(false)

	return tagsModel{
		list: l,
	}
}

func (m tagsModel) reset() tagsModel {
	m.selected = ""
	m.back = false
	m.repository = ""
	m.err = nil
	m.loading = false
	m.list.ResetSelected()
	m.list.SetItems(nil)
	return m
}

type tagsResultMsg struct {
	tags []string
	err  error
}

func (m tagsModel) Init() tea.Cmd {
	return m.fetchTags()
}

func (m tagsModel) Update(msg tea.Msg) (tagsModel, tea.Cmd) {
	if msg == nil {
		return m, nil
	}

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()

		switch key {
		case "enter":
			if m.loading {
				return m, nil
			}
			if len(m.list.Items()) > 0 {
				if i, ok := m.list.SelectedItem().(tagItem); ok {
					m.selected = i.name
				}
			}
		case "esc", "q":
			m.back = true
			return m, nil
		case "ctrl+c":
			return m, tea.Quit
		case "up", "k", "down", "j":
			if m.loading {
				return m, nil
			}
		}

	case tagsResultMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.err = nil
		items := make([]list.Item, len(msg.tags))
		for i, t := range msg.tags {
			items[i] = tagItem{name: t}
		}
		m.list.SetItems(items)
		return m, nil
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v-5)
	}

	m.list, cmd = m.list.Update(msg)

	return m, cmd
}

func (m tagsModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Select Tag") + "\n\n")
	b.WriteString(fmt.Sprintf("Image: %s\n\n", highlightStyle.Render(m.repository)))

	if m.loading {
		b.WriteString("Loading tags...\n")
	} else if m.err != nil {
		b.WriteString(fmt.Sprintf("Error: %v\n", m.err))
	} else {
		b.WriteString(docStyle.Render(m.list.View()))
		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666")).Render("Press Enter to select, Esc/q to return"))
	}

	return b.String()
}

func (m tagsModel) fetchTags() tea.Cmd {
	m.loading = true
	return func() tea.Msg {
		resp, err := registry.ListTags(m.repository, 1, 50)
		if err != nil {
			return tagsResultMsg{err: err}
		}
		tags := make([]string, 0, len(resp.Results))
		for _, t := range resp.Results {
			tags = append(tags, t.Name)
		}
		return tagsResultMsg{tags: tags}
	}
}
