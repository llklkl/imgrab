package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/llklkl/imgrab/internal/registry"
)

type state int

const (
	stateSearch state = iota
	stateTags
	stateConfirm
	stateProgress
	stateDone
)

type Model struct {
	state      state
	search     searchModel
	tags       tagsModel
	confirm    confirmModel
	progress   progressModel
	selected   SelectedImage
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
		state:    stateSearch,
		search:   newSearchModel(initialQuery),
		tags:     newTagsModel(),
		confirm:  newConfirmModel(),
		progress: progressModel{},
	}
}

func (m Model) Init() tea.Cmd {
	// 如果有初始搜索参数，立即执行搜索
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

	switch m.state {
	case stateSearch:
		var cmd tea.Cmd
		m.search, cmd = m.search.Update(msg)
		if m.search.selected != "" {
			m.selected.Name = m.search.selected
			m.selected.Description = m.search.selectedDesc
			m.state = stateTags
			m.tags.repository = m.search.selected

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
		if m.tags.selected != "" {
			m.selected.Tag = m.tags.selected
			m.state = stateConfirm
			m.confirm.image = m.selected
			return m, nil
		}
		if m.tags.back {
			m.state = stateSearch
			// Directly reset to input mode instead of staying on results page
			m.search = m.search.resetToInput()
			m.tags.back = false
			m.tags.selected = ""
			m.tags.repository = ""
			m.tags.list.ResetSelected()
			return m, nil
		}
		return m, cmd
	case stateConfirm:
		var cmd tea.Cmd
		m.confirm, cmd = m.confirm.Update(msg)
		if m.confirm.confirmed {
			m.selected.Arch = m.confirm.arch()
			m.state = stateProgress
			m.progress.image = m.selected
			return m, m.progress.startDownload()
		}
		if m.confirm.back {
			m.state = stateTags
			m.confirm.back = false
			m.confirm.confirmed = false
			return m, nil
		}
		return m, cmd
	case stateProgress:
		var cmd tea.Cmd
		m.progress, cmd = m.progress.Update(msg)
		if m.progress.done {
			m.state = stateDone
			return m, tea.Quit
		}
		return m, cmd
	case stateDone:
		if msg, ok := msg.(tea.KeyMsg); ok {
			if msg.String() == "q" || msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
		}
		return m, nil
	}
	return m, nil
}

func (m Model) View() string {
	switch m.state {
	case stateSearch:
		return m.search.View()
	case stateTags:
		return m.tags.View()
	case stateConfirm:
		return m.confirm.View()
	case stateProgress:
		return m.progress.View()
	case stateDone:
		return fmt.Sprintf("Download complete!\n\nPress q or Ctrl+C to exit\n")
	}
	return ""
}
