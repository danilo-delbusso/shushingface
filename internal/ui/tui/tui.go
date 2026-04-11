package tui

import (
	"context"
	"fmt"

	"codeberg.org/dbus/shushingface/internal/core"
	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	statusStyle = lipgloss.NewStyle().
			Italic(true).
			Foreground(lipgloss.Color("#626262"))

	resultStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(1, 2).
			MarginTop(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			MarginTop(1)
)

type state int

const (
	stateIdle state = iota
	stateRecording
	stateProcessing
	stateDone
	stateError
)

type model struct {
	state      state
	err        error
	transcript string
	refined    string
	spinner    spinner.Model
	viewport   viewport.Model
	engine     *core.Engine
	ctx        context.Context
}

type resultMsg struct {
	transcript string
	refined    string
}

type errMsg struct{ err error }

func NewModel(engine *core.Engine, ctx context.Context) *model {
	return &model{
		engine:   engine,
		ctx:      ctx,
		spinner:  spinner.New(spinner.WithSpinner(spinner.Dot), spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("205")))),
		viewport: viewport.New(0, 0),
	}
}

func (m *model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case " ":
			if m.state == stateIdle || m.state == stateDone || m.state == stateError {
				m.state = stateRecording
				m.transcript = ""
				m.refined = ""
				m.err = nil
				if err := m.engine.StartRecording(); err != nil {
					m.state = stateError
					m.err = err
					return m, nil
				}
				return m, nil
			} else if m.state == stateRecording {
				m.state = stateProcessing
				return m, func() tea.Msg {
					t, r, err := m.engine.StopAndProcess(m.ctx)
					if err != nil {
						return errMsg{err}
					}
					return resultMsg{t, r}
				}
			}
		case "c":
			if m.state == stateDone && m.refined != "" {
				clipboard.WriteAll(m.refined)
			}
		}

	case resultMsg:
		m.state = stateDone
		m.transcript = msg.transcript
		m.refined = msg.refined
		m.viewport.SetContent(m.refined)
		return m, nil

	case errMsg:
		m.state = stateError
		m.err = msg.err
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width - 6
		m.viewport.Height = msg.Height - 12
	}

	return m, nil
}

func (m *model) View() string {
	var s string

	s += titleStyle.Render("SUSSURRO SPEECH TRANSPILER") + "\n\n"

	switch m.state {
	case stateIdle:
		s += statusStyle.Render("Ready to record...")
	case stateRecording:
		s += lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Render("● Recording...")
	case stateProcessing:
		s += m.spinner.View() + statusStyle.Render(" Transcribing and refining...")
	case stateDone:
		s += statusStyle.Render("Success!")
		s += "\n" + resultStyle.Render(m.viewport.View())
	case stateError:
		s += lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Render(fmt.Sprintf("Error: %v", m.err))
	}

	s += helpStyle.Render("\n[Space] Toggle Record • [C] Copy Refined • [Q] Quit")

	return lipgloss.NewStyle().Padding(1, 2).Render(s)
}

func Start(engine *core.Engine, ctx context.Context) error {
	p := tea.NewProgram(NewModel(engine, ctx), tea.WithAltScreen(), tea.WithContext(ctx))
	_, err := p.Run()
	return err
}
