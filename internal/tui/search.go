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

func (i searchItem) Title() string { return i.name }
func (i searchItem) Description() string {
	desc := i.description
	parts := []string{}
	if i.owner != "" {
		parts = append(parts, fmt.Sprintf("Owner: %s", i.owner))
	}
	parts = append(parts, fmt.Sprintf("Stars: %d | Pulls: %s", i.stars, registry.FormatNumber(i.pullCount)))
	if desc != "" {
		parts = append(parts, desc)
	}
	return strings.Join(parts, " | ")
}
func (i searchItem) FilterValue() string { return i.name }

type searchModel struct {
	searchInput  textinput.Model
	list         list.Model
	searching    bool
	selected     string
	selectedDesc string
	err          error
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
		// Add support for focusing list with tab or arrow keys from input
		case "down", "j":
			if m.searchInput.Focused() {
				m.searchInput.Blur()
			}
		case "up", "k":
			if m.searchInput.Focused() {
				// Stay in input mode if up key pressed while focused
			}
		case "tab":
			if m.searchInput.Focused() {
				m.searchInput.Blur()
			} else {
				m.searchInput.Focus()
				m.list.ResetSelected()
			}
		case "ctrl+c":
			return m, tea.Quit
		case "q":
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
		m.searchInput.Blur()
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

	b.WriteString("imgrab - Docker Image Pull Tool\n\n")
	b.WriteString(m.searchInput.View())
	b.WriteString("\n\n")

	if m.searching {
		b.WriteString("Searching...\n")
	} else if m.err != nil {
		b.WriteString(fmt.Sprintf("Error: %v\n", m.err))
	} else if !m.searchInput.Focused() {
		b.WriteString(docStyle.Render(m.list.View()))
		b.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#666")).Render("Press Enter to select, Esc to return"))
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
	return m
}
