package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/schollz/progressbar/v3"

	"github.com/llklkl/imgrab/internal/registry"
)

var (
	progressStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1, 2).
			Width(70)
)

type progressModel struct {
	image    SelectedImage
	downloading bool
	done     bool
	err      error
	bar      *progressbar.ProgressBar
	progress int64
	total    int64
}

type progressMsg struct {
	progress int64
	total    int64
}

type downloadDoneMsg struct {
	err error
}

func (m progressModel) startDownload() tea.Cmd {
	return func() tea.Msg {
		imageRef := fmt.Sprintf("%s:%s", m.image.Name, m.image.Tag)
		err := registry.NewClient().PullAndSave(imageRef, ".")
		if err != nil {
			return downloadDoneMsg{err: err}
		}
		return downloadDoneMsg{}
	}
}

func (m progressModel) Init() tea.Cmd {
	return nil
}

func (m progressModel) Update(msg tea.Msg) (progressModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			if m.done {
				return m, tea.Quit
			}
		}
	case downloadDoneMsg:
		m.downloading = false
		m.done = true
		m.err = msg.err
	}
	return m, nil
}

func (m progressModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("下载中") + "\n\n")
	b.WriteString(fmt.Sprintf("镜像: %s\n", highlightStyle.Render(m.image.Name)))
	b.WriteString(fmt.Sprintf("版本: %s\n", highlightStyle.Render(m.image.Tag)))
	b.WriteString(fmt.Sprintf("架构: %s\n\n", highlightStyle.Render(m.image.Arch)))

	if m.done {
		if m.err != nil {
			b.WriteString(fmt.Sprintf("下载失败: %v\n", m.err))
		} else {
			b.WriteString("下载完成！\n")
		}
		b.WriteString("\n按 q 或 Ctrl+C 退出\n")
	} else {
		b.WriteString("正在下载镜像...\n")
	}

	return progressStyle.Render(b.String())
}
