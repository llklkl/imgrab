package tui

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/llklkl/imgrab/internal/docker"
	"github.com/llklkl/imgrab/internal/registry"
)

var progressDebugLog *log.Logger

func init() {
	f, err := os.OpenFile("/tmp/imgrab_progress_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		progressDebugLog = log.New(f, "[PROGRESS] ", log.Ltime|log.Lmicroseconds)
	}
}

var (
	progressStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1, 2).
		Width(70)
)

type progressModel struct {
	image            SelectedImage
	action           int
	downloading      bool
	done             bool
	importing        bool
	err              error
	progress         int64
	total            int64
	tarPath          string
	startTime        time.Time
	speed            float64
	countdownSeconds int
	autoExitPending  bool
	progressChan     chan registry.ProgressUpdate
	doneChan         chan struct {
		tarPath string
		err     error
	}
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

type tickMsg struct {
	time time.Time
}

func (m *progressModel) startDownload() tea.Cmd {
	progressDebugLog.Printf("startDownload called for image=%s:%s arch=%s", m.image.Name, m.image.Tag, m.image.Arch)

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

	m.progressChan = make(chan registry.ProgressUpdate)
	m.doneChan = make(chan struct {
		tarPath string
		err     error
	})

	progressChan := m.progressChan
	doneChan := m.doneChan

	go func() {
		progressDebugLog.Printf("goroutine: parsing image ref=%s arch=%s", imageRef, arch)
		ref, err := registry.ParseImageRef(imageRef, arch, "")
		if err != nil {
			progressDebugLog.Printf("goroutine: parse image ref error: %v", err)
			doneChan <- struct {
				tarPath string
				err     error
			}{"", err}
			return
		}

		progressDebugLog.Printf("goroutine: getting credential for registry=%s", ref.Registry)
		auth, err := registry.GetCredential(ref.Registry)
		if err != nil {
			progressDebugLog.Printf("goroutine: get credential error: %v", err)
			doneChan <- struct {
				tarPath string
				err     error
			}{"", err}
			return
		}

		progressDebugLog.Printf("goroutine: pulling image")
		client := registry.NewClient().WithAuth(auth)
		img, err := client.PullImage(ref)
		if err != nil {
			progressDebugLog.Printf("goroutine: pull image error: %v", err)
			doneChan <- struct {
				tarPath string
				err     error
			}{"", err}
			return
		}

		progressDebugLog.Printf("goroutine: saving image to tar")
		opts := &registry.PullOptions{
			OutputDir:    ".",
			ShowProgress: false,
			ProgressChan: progressChan,
		}

		outputPath, err := registry.SaveImageToTar(img, ref, opts)
		progressDebugLog.Printf("goroutine: save image done, outputPath=%s, err=%v", outputPath, err)
		doneChan <- struct {
			tarPath string
			err     error
		}{outputPath, err}
	}()

	return m.continueDownload(progressChan, doneChan)
}

func (m progressModel) continueDownload(progressChan chan registry.ProgressUpdate, doneChan chan struct {
	tarPath string
	err     error
}) tea.Cmd {
	return func() tea.Msg {
		for {
			select {
			case update := <-progressChan:
				progressDebugLog.Printf("progress update: progress=%d total=%d", update.Progress, update.Total)
				return progressMsg{progress: update.Progress, total: update.Total}
			case result := <-doneChan:
				progressDebugLog.Printf("download done: tarPath=%s err=%v", result.tarPath, result.err)
				return downloadDoneMsg{tarPath: result.tarPath, err: result.err}
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
	progressDebugLog.Printf("Init: starting tick timer")
	return tea.Every(500*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg{time: t}
	})
}

func (m progressModel) Update(msg tea.Msg) (progressModel, tea.Cmd) {
	progressDebugLog.Printf("Update: received msg type=%T, done=%v, importing=%v", msg, m.done, m.importing)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		progressDebugLog.Printf("Update: KeyMsg key=%s", key)
		switch key {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		}
	case progressMsg:
		progressDebugLog.Printf("Update: progressMsg progress=%d total=%d", msg.progress, msg.total)
		m.progress = msg.progress
		m.total = msg.total
		if !m.downloading {
			m.downloading = true
			m.startTime = time.Now()
		}
		elapsed := time.Since(m.startTime).Seconds()
		if elapsed > 0 {
			m.speed = float64(m.progress) / elapsed
		}
		return m, m.continueDownload(m.progressChan, m.doneChan)
	case downloadDoneMsg:
		progressDebugLog.Printf("Update: downloadDoneMsg tarPath=%s err=%v action=%d", msg.tarPath, msg.err, m.action)
		m.downloading = false
		m.err = msg.err
		m.tarPath = msg.tarPath
		if m.err != nil {
			m.done = true
			return m, nil
		} else if m.action == actionImportDocker {
			m.importing = true
			return m, m.startImport()
		} else {
			m.done = true
			progressDebugLog.Printf("Download complete, exiting immediately")
			return m, tea.Quit
		}
	case importDoneMsg:
		progressDebugLog.Printf("Update: importDoneMsg err=%v", msg.err)
		m.importing = false
		m.done = true
		if msg.err != nil && m.err == nil {
			m.err = msg.err
		}
		progressDebugLog.Printf("Import complete, exiting immediately")
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
	b.WriteString(fmt.Sprintf("Architecture: %s\n", highlightStyle.Render(m.image.Arch)))

	actionStr := "Download Only"
	if m.action == actionImportDocker {
		actionStr = "Import to Docker"
	}
	b.WriteString(fmt.Sprintf("Action: %s\n\n", highlightStyle.Render(actionStr)))

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
		if m.autoExitPending && m.countdownSeconds > 0 {
			b.WriteString(fmt.Sprintf("\nAuto-exiting in %d second(s)...\n", m.countdownSeconds))
		} else {
			b.WriteString("\nPress q or Ctrl+C to exit\n")
		}
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

func (m progressModel) progressViewWithoutBorder() string {
	var b strings.Builder

	actionStr := "Download Only"
	if m.action == actionImportDocker {
		actionStr = "Import to Docker"
	}

	if m.done {
		if m.err != nil {
			if m.importing || (m.action == actionImportDocker && !strings.Contains(m.err.Error(), "docker")) {
				b.WriteString(fmt.Sprintf("Download and/or import failed: %v\n", m.err))
			} else {
				b.WriteString(fmt.Sprintf("Download failed: %v\n", m.err))
			}
			b.WriteString("\nPress q or Ctrl+C to exit\n")
		} else {
			if m.action == actionImportDocker {
				b.WriteString("Download complete and imported to Docker!\n")
			} else {
				b.WriteString("Download complete!\n")
			}
			if m.tarPath != "" {
				b.WriteString(fmt.Sprintf("Image saved to: %s\n", m.tarPath))
			}
			if m.autoExitPending && m.countdownSeconds > 0 {
				b.WriteString(fmt.Sprintf("\nAuto-exiting in %d second(s)...\n", m.countdownSeconds))
			} else {
				b.WriteString("\nPress q or Ctrl+C to exit\n")
			}
		}
	} else if m.importing {
		b.WriteString(fmt.Sprintf("Importing image to Docker... [%s]\n", actionStr))
	} else {
		if m.total > 0 {
			percent := float64(m.progress) / float64(m.total) * 100

			var eta string
			if m.speed > 0 {
				remaining := float64(m.total - m.progress)
				etaSeconds := remaining / m.speed
				eta = formatDuration(time.Duration(etaSeconds) * time.Second)
			}

			b.WriteString(fmt.Sprintf("Progress: %.1f%% (%s / %s)\n",
				percent, formatBytes(m.progress), formatBytes(m.total)))

			speedStr := fmt.Sprintf("Speed: %s/s", formatBytes(int64(m.speed)))
			if eta != "" {
				speedStr += fmt.Sprintf(" | ETA: %s", eta)
			}
			b.WriteString(speedStr + "\n")

			barWidth := 50
			filled := int(float64(barWidth) * float64(m.progress) / float64(m.total))
			if filled > barWidth {
				filled = barWidth
			}
			bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
			b.WriteString(fmt.Sprintf("[%s]\n", bar))
		} else {
			b.WriteString(fmt.Sprintf("Downloading image... [%s]\n", actionStr))
		}
		if !m.downloading && m.progress == 0 && m.total == 0 {
			b.WriteString("\nPress q or Ctrl+C to cancel\n")
		}
	}

	return b.String()
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h > 0 {
		return fmt.Sprintf("%dh%dm%ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm%ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
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
