package tui

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type menu struct {
	title       string
	description string
	action      string
}

type menuModel struct {
	choices  []menu
	cursor   int
	quitting bool
	selected menu
	spinner  spinner.Model
	err      error
}

func newMenu(title, description, action string) menu {
	return menu{
		title:       title,
		description: description,
		action:      action,
	}
}

func newMenuModel() menuModel {
	sp := spinner.New()
	sp.Spinner = spinner.Ellipsis
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("200"))
	items := []menu{
		newMenu("Run evaluation", "Execute a prompt suite", "eval"),
		newMenu("Run benchmark", "Measure model latency", "bench"),
		newMenu("YAML format check", "Validate YAML file", "lint"),
		newMenu("Quit", "Exit promptctl", "q"),
	}

	return menuModel{
		choices: items,
		spinner: sp,
	}
}

func RunMenu(defaultSuitePath string) (string, string, error) {
	program := tea.NewProgram(newMenuModel())
	finalModel, err := program.Run()
	if err != nil {
		return "", "", err
	}

	model := finalModel.(menuModel)
	if model.selected.action != "eval" {
		return model.selected.action, "", nil
	}

	suitePath, err := promptForSuitePath(defaultSuitePath)
	if err != nil {
		return "", "", err
	}
	return model.selected.action, suitePath, nil
}

func promptForSuitePath(defaultPath string) (string, error) {
	fmt.Printf("YAML suite path [%s]: ", defaultPath)
	input, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && len(input) == 0 {
		return "", err
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return defaultPath, nil
	}
	return input, nil
}

func (m menuModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m menuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter":
			m.selected = m.choices[m.cursor]
			if m.selected.action == "q" {
				return m, tea.Quit
			}
			return m, tea.Quit
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m menuModel) View() string {
	if m.err != nil {
		return m.err.Error()
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%s promptctl\n\n", m.spinner.View())

	for i, c := range m.choices {
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}
		fmt.Fprintf(&b, "%s %s\n", cursor, c.title)
	}
	return b.String()

}
