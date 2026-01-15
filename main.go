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

type model struct {
	projects []string
	cursor   int
	selected string
	frame    int
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
			m.selected = m.projects[m.cursor]
			return m, tea.Quit
		}
	case tickMsg:
		m.frame++
		return m, tick()
	}
	return m, nil
}

func (m model) View() string {
	if m.selected != "" {
		return fmt.Sprintf("\n✨ Ótima escolha! Abrindo o projeto: %s 🚀\n", m.selected)
	}

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

	// Cat view
	currentFrame := catFrames[m.frame%len(catFrames)]
	cat := catStyle.Render(currentFrame)

	// Combine views side-by-side
	return lipgloss.JoinHorizontal(lipgloss.Top, s, cat)
}

func main() {
	p := tea.NewProgram(initialModel())
	m, err := p.Run()
	if err != nil {
		fmt.Printf("Ocorreu um erro: %v", err)
		os.Exit(1)
	}

	if finalModel, ok := m.(model); ok && finalModel.selected != "" {
		openTmux(finalModel.selected)
	}
}

func openTmux(project string) {
	// Check if inside tmux
	inTmux := os.Getenv("TMUX") != ""
	projectPath := filepath.Join("/home/brazlucas/repo", project)

	var cmd *exec.Cmd
	if inTmux {
		// Create a new window
		cmd = exec.Command("tmux", "new-window", "-c", projectPath, "-n", project)
	} else {
		// Create a new session
		// We name the session after the project
		// If session exists, we should probably attach to it?
		// For simplicity, let's try new-session. If it fails, maybe user wants to attach.
		// But let's start with basic new-session.
		cmd = exec.Command("tmux", "new-session", "-A", "-s", project, "-c", projectPath)
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Erro ao abrir tmux: %v\n", err)
	}
}
