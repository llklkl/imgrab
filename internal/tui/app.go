package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/llklkl/imgrab/internal/registry"
)

type state int

const (
	stateSearchInput state = iota
	stateSearchResults
	stateTags
	stateArchSelect
	stateDownload
	stateDone
)

func (s state) String() string {
	switch s {
	case stateSearchInput:
		return "SearchInput"
	case stateSearchResults:
		return "SearchResults"
	case stateTags:
		return "Tags"
	case stateArchSelect:
		return "ArchSelect"
	case stateDownload:
		return "Download"
	case stateDone:
		return "Done"
	default:
		return fmt.Sprintf("Unknown(%d)", s)
	}
}

type Model struct {
	state      state
	search     searchModel
	tags       tagsModel
	arch       archModel
	confirm    confirmModel
	download   progressModel
	selected   SelectedImage
	action     int
	windowSize tea.WindowSizeMsg
}

type SelectedImage struct {
	Name        string
	Description string
	Tag         string
	Arch        string
}

func NewModel(initialQuery string) Model {
	return Model{
		state:   stateSearchInput,
		search:  newSearchModel(initialQuery),
		tags:    newTagsModel(),
		arch:    newArchModel(),
		confirm: newConfirmModel(),
	}
}

func (m Model) Init() tea.Cmd {
	if m.search.searchInput.Value() != "" {
		return func() tea.Msg {
			resp, err := registry.SearchImages(m.search.searchInput.Value(), 1, 20)
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
	return m.search.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		m.windowSize = ws
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "ctrl+c" {
		return m, tea.Quit
	}

	switch m.state {
	case stateSearchInput:
		var cmd tea.Cmd
		m.search, cmd = m.search.Update(msg)

		if m.search.selected != "" {
			selectedRepo := m.search.selected
			m.search.selected = ""
			m.search.selectedDesc = ""

			m.selected.Name = selectedRepo
			m.state = stateTags
			m.tags.repository = selectedRepo

			if m.windowSize.Width > 0 && m.windowSize.Height > 0 {
				h, v := 2, 4
				m.tags.list.SetSize(m.windowSize.Width-h, m.windowSize.Height-v-5)
			}

			return m, m.tags.Init()
		}

		if (m.search.searching && !m.search.searchInput.Focused()) ||
			(!m.search.searchInput.Focused() && len(m.search.list.Items()) > 0) {
			m.state = stateSearchResults
			return m, cmd
		}
		return m, cmd

	case stateSearchResults:
		var cmd tea.Cmd
		m.search, cmd = m.search.Update(msg)

		if m.search.back {
			m.search.back = false
			m.state = stateSearchInput
			return m, cmd
		}

		if m.search.selected != "" {
			selectedRepo := m.search.selected
			m.search.selected = ""
			m.search.selectedDesc = ""

			m.selected.Name = selectedRepo
			m.state = stateTags
			m.tags.repository = selectedRepo

			if m.windowSize.Width > 0 && m.windowSize.Height > 0 {
				h, v := 2, 4
				m.tags.list.SetSize(m.windowSize.Width-h, m.windowSize.Height-v-5)
			}

			return m, m.tags.Init()
		}
		return m, cmd

	case stateTags:
		var cmd tea.Cmd

		m.tags, cmd = m.tags.Update(msg)

		if m.tags.back {
			m.state = stateSearchResults
			m.tags.back = false
			m.tags.selected = ""
			m.tags.repository = ""
			m.tags.list.ResetSelected()
			m.tags.list.SetItems(nil)
			m.tags.err = nil
			m.tags.loading = false
			m.search.list.ResetSelected()
			m.search.selected = ""
			m.search.selectedDesc = ""
			m.search.searching = false
			return m, nil
		}

		if m.tags.selected != "" {
			m.selected.Tag = m.tags.selected
			m.state = stateArchSelect
			m.arch.image = m.selected
			return m, nil
		}

		return m, cmd

	case stateArchSelect:
		var cmd tea.Cmd
		m.arch, cmd = m.arch.Update(msg)
		if m.arch.confirmed {
			m.selected.Arch = m.arch.arch()
			m.state = stateDownload
			m.confirm.image = m.selected
			return m, nil
		}
		if m.arch.back {
			m.state = stateTags
			m.arch.back = false
			m.arch.confirmed = false
			m.tags.selected = ""
			m.tags.list.ResetSelected()
			return m, nil
		}
		return m, cmd

	case stateDownload:
		var cmd tea.Cmd
		m.confirm, cmd = m.confirm.Update(msg)
		if m.confirm.confirmed {
			m.action = m.confirm.action()
			m.state = stateDone
			m.download.image = m.selected
			m.download.action = m.action
			m.download.downloading = true
			return m, m.download.startDownload()
		}
		if m.confirm.back {
			m.state = stateArchSelect
			m.confirm.back = false
			m.confirm.confirmed = false
			m.arch.confirmed = false
			m.arch.back = false
			return m, nil
		}
		return m, cmd

	case stateDone:
		var cmd tea.Cmd
		m.download, cmd = m.download.Update(msg)
		if msg, ok := msg.(tea.KeyMsg); ok {
			if msg.String() == "q" || msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
		}
		return m, cmd
	}
	return m, nil
}

func (m Model) View() string {
	switch m.state {
	case stateSearchInput:
		return m.search.View()
	case stateSearchResults:
		return m.search.View()
	case stateTags:
		return m.tags.View()
	case stateArchSelect:
		return m.arch.View()
	case stateDownload:
		return m.confirm.View()
	case stateDone:
		return m.download.View()
	}
	return ""
}
