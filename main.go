package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			MarginBottom(1)

	itemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	selectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(0).
				Foreground(lipgloss.Color("205")).
				Bold(true)

	quitStyle = lipgloss.NewStyle().
			MarginTop(1).
			Foreground(lipgloss.Color("241"))

	catStyle = lipgloss.NewStyle().
			MarginLeft(4).
			Foreground(lipgloss.Color("#FFB6C1")) // Light pink for the kitty

	errStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			MarginTop(1)
)

var catFrames = []string{
	`
   /\_/\
  ( o.o )
   > ^ <
`,
	`
   /\_/\
  ( -.- )
   > ^ <
`,
	`
   /\_/\
  ( o.o )
   > ^ <
`,
	`
    /\_/\
   ( o.o )
    > ^ <
`,
	`
  /\_/\
 ( o.o )
  > ^ <
`,
	`
   /\_/\
  ( o.o )
   > ^ <
   oo
`,
	`
   /\_/\
  ( -.- )
   > p <
   oo
`,
	`
   /\_/\
  ( o.o )
   > p <
   oo
`,
}

type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(time.Millisecond*200, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type projectOpenedMsg struct {
	err error
}

type model struct {
	projects []string
	cursor   int
	frame    int
	err      error
}

func initialModel() model {
	// Path to the repository containing projects
	repoPath := "/home/brazlucas/repo"

	entries, err := os.ReadDir(repoPath)
	if err != nil {
		log.Fatal(err)
	}

	var projects []string
	for _, e := range entries {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			projects = append(projects, e.Name())
		}
	}

	return model{
		projects: projects,
	}
}

func (m model) Init() tea.Cmd {
	return tick()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.projects)-1 {
				m.cursor++
			}
		case "enter":
			project := m.projects[m.cursor]
			cmd := getTmuxCmd(project)
			return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
				return projectOpenedMsg{err}
			})
		}
	case tickMsg:
		m.frame++
		return m, tick()
	case projectOpenedMsg:
		if msg.err != nil {
			m.err = msg.err
		}
		return m, nil
	}
	return m, nil
}

func (m model) View() string {
	// Project list view
	s := titleStyle.Render("📂 Seletor de Projetos") + "\n"

	for i, project := range m.projects {
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = "👉" // cursor!
			s += selectedItemStyle.Render(fmt.Sprintf("%s %s", cursor, project)) + "\n"
		} else {
			s += itemStyle.Render(project) + "\n"
		}
	}

	s += quitStyle.Render("\n(use ↑/↓ para navegar, enter para selecionar, q para sair)\n")

	if m.err != nil {
		s += errStyle.Render(fmt.Sprintf("\nErro: %v", m.err))
	}

	// Cat view
	currentFrame := catFrames[m.frame%len(catFrames)]
	cat := catStyle.Render(currentFrame)

	// Combine views side-by-side
	return lipgloss.JoinHorizontal(lipgloss.Top, s, cat)
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Ocorreu um erro: %v", err)
		os.Exit(1)
	}
}

func getTmuxCmd(project string) *exec.Cmd {
	// Check if inside tmux
	inTmux := os.Getenv("TMUX") != ""
	projectPath := filepath.Join("/home/brazlucas/repo", project)

	if inTmux {
		// Create a new window
		return exec.Command("tmux", "new-window", "-c", projectPath, "-n", project)
	}

	// Create a new session
	// We name the session after the project
	return exec.Command("tmux", "new-session", "-A", "-s", project, "-c", projectPath)
}
