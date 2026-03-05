package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

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
	image       SelectedImage
	downloading bool
	done        bool
	err         error
	progress    int64
	total       int64
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
		arch := m.image.Arch
		if arch == "" {
			arch = ""
		} else {
			parts := strings.Split(arch, "/")
			if len(parts) >= 2 {
				arch = parts[1]
			}
		}
		err := registry.NewClient().PullAndSave(imageRef, arch, ".", nil)
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
		case "q", "ctrl+c", "esc":
			// Always allow exiting from any state
			if m.done {
				return m, tea.Quit
			}
			// Handle download cancellation
			return m, tea.Quit
		}
	case progressMsg:
		m.progress = msg.progress
		m.total = msg.total
	case downloadDoneMsg:
		m.downloading = false
		m.done = true
		m.err = msg.err
	}
	return m, nil
}

func (m progressModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Downloading") + "\n\n")
	b.WriteString(fmt.Sprintf("Image: %s\n", highlightStyle.Render(m.image.Name)))
	b.WriteString(fmt.Sprintf("Tag: %s\n", highlightStyle.Render(m.image.Tag)))
	b.WriteString(fmt.Sprintf("Architecture: %s\n\n", highlightStyle.Render(m.image.Arch)))

	if m.done {
		if m.err != nil {
			b.WriteString(fmt.Sprintf("Download failed: %v\n", m.err))
		} else {
			b.WriteString("Download complete!\n")
		}
		b.WriteString("\nPress q or Ctrl+C to exit\n")
	} else {
		if m.total > 0 {
			percent := float64(m.progress) / float64(m.total) * 100
			b.WriteString(fmt.Sprintf("Progress: %.1f%% (%s / %s)\n",
				percent, formatBytes(m.progress), formatBytes(m.total)))

			barWidth := 50
			filled := int(float64(barWidth) * float64(m.progress) / float64(m.total))
			if filled > barWidth {
				filled = barWidth
			}
			bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
			b.WriteString(fmt.Sprintf("[%s]\n", bar))
		} else {
			b.WriteString("Downloading image...\n")
		}
		b.WriteString("\nPress q or Ctrl+C to cancel\n")
	}

	return progressStyle.Render(b.String())
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
