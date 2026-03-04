package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
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
}

type SelectedImage struct {
	Name        string
	Description string
	Tag         string
	Arch        string
}

func NewModel() Model {
	return Model{
		state:   stateSearch,
		search:  newSearchModel(),
		tags:    newTagsModel(),
		confirm: newConfirmModel(),
		progress: progressModel{},
	}
}

func (m Model) Init() tea.Cmd {
	return m.search.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.state {
	case stateSearch:
		var cmd tea.Cmd
		m.search, cmd = m.search.Update(msg)
		if m.search.selected != "" {
			m.selected.Name = m.search.selected
			m.selected.Description = m.search.selectedDesc
			m.state = stateTags
			m.tags.repository = m.search.selected
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
			m.search.selected = ""
			m.tags.back = false
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
		return fmt.Sprintf("下载完成！\n\n按 q 或 Ctrl+C 退出\n")
	}
	return ""
}
