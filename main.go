package main

import (
	"errors"
	"fmt"
	"okapp/widget/mp3"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"okapp/libs/filepicker"
)

type model struct {
	filePicker   filepicker.Model
	selectedFile string
	quitting     bool
	err          error
	msg          string
}

type clearErrorMsg struct{}

func clearErrorAfter(t time.Duration) tea.Cmd {
	return tea.Tick(t, func(_ time.Time) tea.Msg {
		return clearErrorMsg{}
	})
}

func (m model) Init() tea.Cmd {
	return m.filePicker.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.err != nil {
		m.err = nil
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		m.msg = ""
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "s", " ":
			entry := m.filePicker.GetSelectedFilename()
			if entry.IsDir() {
				m.selectedFile = filepath.Join(m.filePicker.CurrentDirectory, entry.Name())
			} else {
				m.selectedFile = m.filePicker.CurrentDirectory
			}
		case "alt+x":
			tidyResult := mp3.BatchRemoveMp3Tag(m.selectedFile)
			if tidyResult.Error != nil {
				m.err = tidyResult.Error
			} else {
				m.msg = fmt.Sprintf(
					"总文件数：%d；清理音频文件数：%d；错误数：%d",
					tidyResult.FileTotal,
					tidyResult.AudioTotal,
					tidyResult.ErrorTotal,
				)
			}
		}
	case clearErrorMsg:
		m.err = nil
	}
	var cmd tea.Cmd
	m.filePicker, cmd = m.filePicker.Update(msg)
	// Did the user select a file?
	if didSelect, path := m.filePicker.DidSelectFile(msg); didSelect {
		// Get the path of the selected file.
		m.selectedFile = path
	}
	// Did the user select a disabled file?
	// This is only necessary to display an error to the user.
	if didSelect, path := m.filePicker.DidSelectDisabledFile(msg); didSelect {
		// Let's clear the selectedFile and display an error.
		m.err = errors.New(path + " is not valid.")
		m.selectedFile = ""
		return m, tea.Batch(cmd, clearErrorAfter(2*time.Second))
	}
	return m, cmd
}

func (m model) View() string {
	if m.quitting {
		return ""
	}
	var s strings.Builder
	s.WriteString("\n")
	if m.err != nil {
		s.WriteString(m.filePicker.Styles.DisabledFile.Render(m.err.Error()))
	} else if m.selectedFile == "" {
		s.WriteString("请选择需要整理的文件夹:")
	} else if m.msg != "" {
		s.WriteString(m.msg)
	} else {
		s.WriteString("已选择文件夹: " + m.filePicker.Styles.Selected.Render(m.selectedFile))
	}
	s.WriteString("\n\n" + m.filePicker.View() + "\n")
	return s.String()
}

func main() {
	fp := filepicker.New()
	fp.CurrentDirectory, _ = os.UserHomeDir()
	fp.DirAllowed = true
	fp.FileAllowed = false
	fp.ShowHidden = false
	m := model{
		filePicker: fp,
	}
	tm, _ := tea.NewProgram(&m, tea.WithOutput(os.Stderr)).Run()
	mm := tm.(model)
	fmt.Println("\n  You selected: " + m.filePicker.Styles.Selected.Render(mm.selectedFile) + "\n")
}
