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
	pullCount   int64
	isOfficial  bool
	owner       string
}

func (i searchItem) Title() string {
	if i.owner != "" {
		return fmt.Sprintf("%s/%s", i.owner, i.name)
	}
	return i.name
}

func (i searchItem) Description() string {
	parts := []string{}
	parts = append(parts, fmt.Sprintf("Stars: %d | Pulls: %s", i.stars, registry.FormatNumber(i.pullCount)))
	if i.description != "" {
		parts = append(parts, i.description)
	}
	return strings.Join(parts, " | ")
}

func (i searchItem) FilterValue() string {
	if i.owner != "" {
		return fmt.Sprintf("%s/%s", i.owner, i.name)
	}
	return i.name
}

type searchModel struct {
	searchInput  textinput.Model
	list         list.Model
	searching    bool
	selected     string
	selectedDesc string
	err          error
	back         bool
}

func newSearchModel(initialQuery string) searchModel {
	ti := textinput.New()
	ti.Placeholder = "Search image (e.g. nginx)..."
	ti.CharLimit = 100
	ti.Width = 50

	if initialQuery != "" {
		ti.SetValue(initialQuery)
		ti.Blur()
	} else {
		ti.Focus()
	}

	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Docker Images"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.SetShowPagination(false)

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
		key := msg.String()

		switch key {
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
					if i.owner != "" {
						m.selected = fmt.Sprintf("%s/%s", i.owner, i.name)
					}
					m.selectedDesc = i.description
				}
			}
		case "esc", "q":
			if !m.searchInput.Focused() && !m.searching {
				m.back = true
				m.searchInput.Focus()
				m.searchInput.CursorEnd()
				m.searchInput.SetCursor(len(m.searchInput.Value()))
				return m, nil
			}
		case "down", "j":
			if m.searchInput.Focused() {
				m.searchInput.Blur()
			}
		case "up", "k":
		case "tab":
			if m.searchInput.Focused() {
				m.searchInput.Blur()
			} else {
				m.searchInput.Focus()
				m.list.ResetSelected()
			}
		case "ctrl+c":
			if m.searchInput.Focused() {
				return m, tea.Quit
			}
			return m, nil
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
		m.searchInput.Blur()
		return m, nil

	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v-5)
	}

	if m.searchInput.Focused() {
		m.searchInput, cmd = m.searchInput.Update(msg)
	} else {
		if km, ok := msg.(tea.KeyMsg); ok && (km.String() == "esc" || km.String() == "q") {
			return m, nil
		}
		m.list, cmd = m.list.Update(msg)
	}

	return m, cmd
}

func (m searchModel) View() string {
	var b strings.Builder

	b.WriteString("imgrab - Docker Image Pull Tool\n\n")
	b.WriteString(m.searchInput.View())
	b.WriteString("\n\n")

	if m.searching {
		b.WriteString("Searching...\n")
	} else if m.err != nil {
		b.WriteString(fmt.Sprintf("Error: %v\n", m.err))
	} else if !m.searchInput.Focused() {
		b.WriteString(docStyle.Render(m.list.View()))
		b.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#666")).Render("Press Enter to select, Esc/q to return"))
	} else {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666")).Render("Press Enter to search, Ctrl+C to quit"))
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
				pullCount:   r.PullCount,
				isOfficial:  r.IsOfficial,
				owner:       r.RepoOwner,
			})
		}

		return searchResultMsg{results: items}
	}
}

func (m searchModel) resetToInput() searchModel {
	m.searchInput.Focus()
	m.list.ResetSelected()
	m.selected = ""
	m.selectedDesc = ""
	m.back = false
	return m
}
