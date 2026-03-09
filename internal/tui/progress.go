package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/llklkl/imgrab/internal/docker"
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
	action      int
	downloading bool
	done        bool
	importing   bool
	err         error
	progress    int64
	total       int64
	tarPath     string
}

type progressMsg struct {
	progress int64
	total    int64
}

type downloadDoneMsg struct {
	tarPath string
	err     error
}

type importDoneMsg struct {
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

		progressChan := make(chan registry.ProgressUpdate)
		doneChan := make(chan struct {
			tarPath string
			err     error
		})

		go func() {
			ref, err := registry.ParseImageRef(imageRef, arch, "")
			if err != nil {
				doneChan <- struct {
					tarPath string
					err     error
				}{"", err}
				return
			}

			auth, err := registry.GetCredential(ref.Registry)
			if err != nil {
				doneChan <- struct {
					tarPath string
					err     error
				}{"", err}
				return
			}

			client := registry.NewClient().WithAuth(auth)
			img, err := client.PullImage(ref)
			if err != nil {
				doneChan <- struct {
					tarPath string
					err     error
				}{"", err}
				return
			}

			opts := &registry.PullOptions{
				OutputDir:    ".",
				ShowProgress: false,
				ProgressChan: progressChan,
			}

			outputPath, err := registry.SaveImageToTar(img, ref, opts)
			doneChan <- struct {
				tarPath string
				err     error
			}{outputPath, err}
		}()

		return func() tea.Msg {
			for {
				select {
				case update := <-progressChan:
					return progressMsg{progress: update.Progress, total: update.Total}
				case result := <-doneChan:
					return downloadDoneMsg{tarPath: result.tarPath, err: result.err}
				}
			}
		}
	}
}

func (m progressModel) startImport() tea.Cmd {
	return func() tea.Msg {
		if m.tarPath == "" {
			return importDoneMsg{err: nil}
		}
		err := docker.ImportTarToDocker(m.tarPath)
		return importDoneMsg{err: err}
	}
}

func (m progressModel) Init() tea.Cmd {
	return nil
}

func (m progressModel) Update(msg tea.Msg) (progressModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()

		switch key {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		}
	case progressMsg:
		m.progress = msg.progress
		m.total = msg.total
		return m, m.startDownload()
	case downloadDoneMsg:
		m.downloading = false
		m.err = msg.err
		m.tarPath = msg.tarPath
		if m.err != nil {
			m.done = true
			return m, tea.Quit
		} else if m.action == actionImportDocker {
			m.importing = true
			return m, m.startImport()
		} else {
			m.done = true
			return m, tea.Quit
		}
	case importDoneMsg:
		m.importing = false
		m.done = true
		if msg.err != nil && m.err == nil {
			m.err = msg.err
		}
		return m, tea.Quit
	}
	return m, nil
}

func (m progressModel) View() string {
	var b strings.Builder

	title := "Downloading"
	if m.importing {
		title = "Importing to Docker"
	}
	b.WriteString(titleStyle.Render(title) + "\n\n")
	b.WriteString(fmt.Sprintf("Image: %s\n", highlightStyle.Render(m.image.Name)))
	b.WriteString(fmt.Sprintf("Tag: %s\n", highlightStyle.Render(m.image.Tag)))
	b.WriteString(fmt.Sprintf("Architecture: %s\n\n", highlightStyle.Render(m.image.Arch)))

	if m.done {
		if m.err != nil {
			if m.importing || (m.action == actionImportDocker && !strings.Contains(m.err.Error(), "docker")) {
				b.WriteString(fmt.Sprintf("Download and/or import failed: %v\n", m.err))
			} else {
				b.WriteString(fmt.Sprintf("Download failed: %v\n", m.err))
			}
		} else {
			if m.action == actionImportDocker {
				b.WriteString("Download complete and imported to Docker!\n")
			} else {
				b.WriteString("Download complete!\n")
			}
			if m.tarPath != "" {
				b.WriteString(fmt.Sprintf("Image saved to: %s\n", m.tarPath))
			}
		}
		b.WriteString("\nPress q or Ctrl+C to exit\n")
	} else if m.importing {
		b.WriteString("Importing image to Docker...\n")
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
		if !m.downloading && m.progress == 0 && m.total == 0 {
			b.WriteString("\nPress q or Ctrl+C to cancel\n")
		}
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
