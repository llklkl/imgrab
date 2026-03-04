package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/llklkl/imgrab/internal/registry"
)

type tagItem string

func (i tagItem) Title() string       { return string(i) }
func (i tagItem) Description() string { return "" }
func (i tagItem) FilterValue() string { return string(i) }

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
	l.Title = "镜像版本"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	return tagsModel{
		list: l,
	}
}

type tagsResultMsg struct {
	tags []string
	err  error
}

func (m tagsModel) Init() tea.Cmd {
	return m.fetchTags()
}

func (m tagsModel) Update(msg tea.Msg) (tagsModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.loading {
				return m, nil
			}
			if i, ok := m.list.SelectedItem().(tagItem); ok {
				m.selected = string(i)
			}
		case "esc":
			m.back = true
		case "q", "ctrl+c":
			return m, tea.Quit
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
			items[i] = tagItem(t)
		}
		m.list.SetItems(items)
		return m, nil

	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v-3)
	}

	m.list, cmd = m.list.Update(msg)

	return m, cmd
}

func (m tagsModel) View() string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("镜像: %s\n\n", m.repository))

	if m.loading {
		b.WriteString("加载版本列表中...\n")
	} else if m.err != nil {
		b.WriteString(fmt.Sprintf("错误: %v\n", m.err))
	} else {
		b.WriteString(docStyle.Render(m.list.View()))
		b.WriteString("\n按 Enter 选择版本, Esc 返回\n")
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
