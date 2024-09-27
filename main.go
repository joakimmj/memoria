package main

import (
	"fmt"
	"log"
	"mia/todo"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// You generally won't need this unless you're processing stuff with
// complicated ANSI escape sequences. Turn it on if you notice flickering.
//
// Also keep in mind that high performance rendering only works for programs
// that use the full size of the terminal. We're enabling that below with
// tea.EnterAltScreen().
const useHighPerformanceRenderer = false

const (
	todoTitle = "TODOs"
	noteTitle = "Notes"
)

// viewState is used to track which model is focused
type viewState uint

const (
	defaultTime           = time.Minute
	todoView    viewState = iota
	noteView
)

var (
	infoBoxStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		b.Left = "┤"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()

	activeStyle = func() lipgloss.Style {
		// TODO: how does this work on lighter terminal windows..?
		return lipgloss.NewStyle().BorderForeground(lipgloss.Color("#A6E3A1")).Foreground(lipgloss.Color("#A6E3A1"))
	}()
)

type mainModel struct {
	view        viewState
	todoView    todo.TodoView
	noteView    viewport.Model
	noteContent string
	ready       bool
}

func (m mainModel) Init() tea.Cmd {
	return tea.Batch()
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab":
			if m.view == todoView {
				m.view = noteView
			} else {
				m.view = todoView
			}
		}
	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.headerView(todoTitle))
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + headerHeight + footerHeight

		contentHeight := msg.Height - verticalMarginHeight
		todoHeight := contentHeight / 4
		noteHeight := (contentHeight / 4) * 3

		if !m.ready {
			// Initiate views
			m.noteView = viewport.New(msg.Width, noteHeight)
			m.noteView.HighPerformanceRendering = useHighPerformanceRenderer
			m.noteView.SetContent(m.noteContent)

			m.todoView.SetWidth(msg.Width)
			m.todoView.SetHeight(todoHeight)

			m.ready = true
		} else {
			// Update views
			m.noteView.Width = msg.Width
			m.noteView.Height = noteHeight

			m.todoView.SetWidth(msg.Width)
			m.todoView.SetHeight(todoHeight)
		}

		if useHighPerformanceRenderer {
			// Render (or re-render) the whole viewport. Necessary both to
			// initialize the viewport and when the window is resized.
			//
			// This is needed for high-performance rendering only.
			cmds = append(cmds, viewport.Sync(m.noteView))
		}
	}

	// Handle keyboard and mouse events in the viewport
	if m.view == todoView {
		// m.todoView.todoTable, cmd = m.todoView.todoTable.Update(msg)
		m.todoView, cmd = m.todoView.Update(msg)
	} else {
		m.noteView, cmd = m.noteView.Update(msg)
	}
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m mainModel) headerView(titleStr string) string {
	activeTitle := m.activeViewTitle()
	title := infoBoxStyle.Render(titleStr)
	line := strings.Repeat("─", max(0, m.noteView.Width-lipgloss.Width(title)))
	header := lipgloss.JoinHorizontal(lipgloss.Center, title, line)

	if activeTitle == titleStr {
		return activeStyle.Render(header)
	}

	return header
}

func (m mainModel) footerView() string {
	info := infoBoxStyle.Render(m.footerText())
	line := strings.Repeat("─", max(0, m.noteView.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

func (m mainModel) footerText() string {
	s := "<tab>: next window • q: exit"
	if m.view == noteView {
		s += " • ↑/k: up • ↓/j: down • e: edit"
	}
	return s
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (m mainModel) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	return fmt.Sprintf("%s\n%s\n%s\n%s\n%s",
		m.headerView(todoTitle),
		m.todoView.View(),
		m.headerView(noteTitle),
		m.noteView.View(),
		m.footerView(),
	)
}

func (m mainModel) activeViewTitle() string {
	if m.view == todoView {
		return todoTitle
	}
	return noteTitle
}

func readFile(filename string) []byte {
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Could not read file (%s): %s\n", filename, err)
		os.Exit(1)
	}
	return content
}

func main() {
	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		fmt.Println("fatal:", err)
		os.Exit(1)
	}
	defer f.Close()

	noteContent := readFile("note.md")

	p := tea.NewProgram(
		mainModel{
			view:        todoView,
			todoView:    todo.NewTodoView(),
			noteContent: string(noteContent),
		},
		tea.WithAltScreen(),       // use the full size of the terminal in its "alternate screen buffer"
		tea.WithMouseCellMotion(), // turn on mouse support so we can track the mouse wheel
	)

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
