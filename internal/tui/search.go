package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/llklkl/imgrab/internal/registry"
)

var (
	docStyle = lipgloss.NewStyle().Margin(1, 2)
)

type searchItem struct {
	name        string
	description string
	stars       int
	isOfficial  bool
}

func (i searchItem) Title() string       { return i.name }
func (i searchItem) Description() string { return i.description }
func (i searchItem) FilterValue() string { return i.name }

type searchModel struct {
	searchInput  textinput.Model
	list         list.Model
	searching    bool
	selected     string
	selectedDesc string
	err          error
}

func newSearchModel() searchModel {
	ti := textinput.New()
	ti.Placeholder = "搜索镜像 (例如: nginx)..."
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 50

	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Docker 镜像"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	return searchModel{
		searchInput: ti,
		list:        l,
	}
}

type searchResultMsg struct {
	results []searchItem
	err     error
}

func (m searchModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m searchModel) Update(msg tea.Msg) (searchModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.searching {
				return m, nil
			}
			if m.searchInput.Focused() {
				query := strings.TrimSpace(m.searchInput.Value())
				if query != "" {
					m.searching = true
					m.searchInput.Blur()
					return m, m.searchImages(query)
				}
			} else {
				if i, ok := m.list.SelectedItem().(searchItem); ok {
					m.selected = i.name
					m.selectedDesc = i.description
				}
			}
		case "esc":
			if !m.searchInput.Focused() {
				m.searchInput.Focus()
				m.list.ResetSelected()
			}
		case "q", "ctrl+c":
			if !m.searchInput.Focused() {
				return m, tea.Quit
			}
		}

	case searchResultMsg:
		m.searching = false
		if msg.err != nil {
			m.err = msg.err
			m.searchInput.Focus()
			return m, nil
		}
		m.err = nil
		items := make([]list.Item, len(msg.results))
		for i, r := range msg.results {
			items[i] = r
		}
		m.list.SetItems(items)
		return m, nil

	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v-5)
	}

	if m.searchInput.Focused() {
		m.searchInput, cmd = m.searchInput.Update(msg)
	} else {
		m.list, cmd = m.list.Update(msg)
	}

	return m, cmd
}

func (m searchModel) View() string {
	var b strings.Builder

	b.WriteString("imgrab - Docker 镜像拉取工具\n\n")
	b.WriteString(m.searchInput.View())
	b.WriteString("\n\n")

	if m.searching {
		b.WriteString("搜索中...\n")
	} else if m.err != nil {
		b.WriteString(fmt.Sprintf("错误: %v\n", m.err))
	} else if !m.searchInput.Focused() {
		b.WriteString(docStyle.Render(m.list.View()))
		b.WriteString("\n按 Enter 选择镜像, Esc 返回搜索\n")
	} else {
		b.WriteString("按 Enter 开始搜索\n")
	}

	return b.String()
}

func (m searchModel) searchImages(query string) tea.Cmd {
	return func() tea.Msg {
		resp, err := registry.SearchImages(query, 1, 20)
		if err != nil {
			return searchResultMsg{err: err}
		}

		items := make([]searchItem, 0, len(resp.Results))
		for _, r := range resp.Results {
			items = append(items, searchItem{
				name:        r.Name,
				description: r.Description,
				stars:       r.Stars,
				isOfficial:  r.IsOfficial,
			})
		}

		return searchResultMsg{results: items}
	}
}
