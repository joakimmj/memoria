package todo

import (
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TodoView interface {
	SetHeight(int)
	SetWidth(int)
	Update(tea.Msg) (TodoView, tea.Cmd)
	View() string
}

const controls = "↑/k: up • ↓/j: down • e: edit • d: delete todo • a: add todo • <space>: toggle done"

type inputState uint

const (
	hide inputState = iota
	add
	edit
	del
)

type viewState struct {
	todos         todos
	todoTable     table.Model
	height        int
	width         int
	hideCompleted bool
	// showInput     bool
	inputState inputState
	textInput  textinput.Model
}

var (
	controlsStyle = lipgloss.NewStyle().MarginTop(1)
)

func (view *viewState) updateTable(current int) {
	todoTable := newTable(&view.todos, view.hideCompleted)
	todoTable.SetCursor(current)
	view.todoTable = todoTable
}

func (view *viewState) SetHeight(height int) {
	view.height = height
}

func (view *viewState) SetWidth(width int) {
	if width > 80 {
		view.width = width
	}
}

func (view *viewState) toggleCompleted() {
	selectedRow := view.todoTable.SelectedRow()
	if len(selectedRow) <= 0 {
		return
	}

	indexField := selectedRow[0]
	index, err := strconv.Atoi(indexField)
	if err == nil {
		view.todos.toggle(index)
		view.updateTable(view.getRowIndex(indexField))
	}
}

// TODO: maybe add some form of validation...
func (view *viewState) deleteTask() {
	selectedRow := view.todoTable.SelectedRow()
	if len(selectedRow) <= 0 {
		return
	}

	indexField := selectedRow[0]
	index, err := strconv.Atoi(indexField)
	if err == nil {
		view.todos.delete(index)
		rowIndex := view.getRowIndex(indexField)
		if rowIndex > len(view.todos) {
			rowIndex = len(view.todos) - 1
		}
		if rowIndex <= 0 {
			rowIndex = 0
		}
		view.updateTable(rowIndex)
	}
}

func (view *viewState) getRowIndex(todoIndex string) int {
	for i, val := range view.todoTable.Rows() {
		if val[0] == todoIndex {
			return i
		}
	}
	return 0
}

func (view *viewState) updateInput(msg tea.Msg) (TodoView, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			view.inputState = hide
			view.todoTable.Focus()
			view.textInput.Blur()
		case "enter":
			switch view.inputState {
			case add:
				view.todos.add(view.textInput.Value())
			case edit:
				selectedRow := view.todoTable.SelectedRow()
				if len(selectedRow) > 0 {
					indexField := selectedRow[0]
					index, err := strconv.Atoi(indexField)
					if err == nil {
						view.todos.edit(index, view.textInput.Value())
					}
				}
			case del:
				if view.textInput.Value() == "y" {
					view.deleteTask()
				}
			}
			view.textInput.SetValue("")
			view.inputState = hide
			view.todoTable.Focus()
			view.textInput.Blur()
			view.updateTable(0) // TODO: add=0, edit=keep, del=ok (n -> keep)
		}
	}

	view.textInput, cmd = view.textInput.Update(msg)
	return view, cmd
}

func (view *viewState) Update(msg tea.Msg) (TodoView, tea.Cmd) {
	if view.inputState != hide {
		return view.updateInput(msg)
	}

	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case " ":
			view.toggleCompleted()
			return view, cmd
		case "e":
			selectedRow := view.todoTable.SelectedRow()
			if len(selectedRow) > 0 {
				view.textInput.SetValue(selectedRow[2])
			}
			view.inputState = edit
			view.todoTable.Blur()
			view.textInput.Placeholder = ""
			view.textInput.Prompt = "> "
			view.textInput.Focus()
			return view, cmd
		case "a":
			view.inputState = add
			view.todoTable.Blur()
			view.textInput.SetValue("")
			view.textInput.Placeholder = "task..."
			view.textInput.Prompt = "> "
			view.textInput.Focus()
			return view, cmd
		case "d":
			view.inputState = del
			view.todoTable.Blur()
			view.textInput.SetValue("")
			view.textInput.Placeholder = "Sure?"
			view.textInput.Prompt = "Are you sure you want to delete this task? (y/N) "
			view.textInput.Focus()
			return view, cmd
		case "h":
			view.hideCompleted = !view.hideCompleted
			view.updateTable(0)
			return view, cmd
		}
	}

	view.todoTable, cmd = view.todoTable.Update(msg)
	return view, cmd
}

func (view *viewState) View() string {
	helpMenu := controlsStyle.Render(controls)
	tableHeight := view.height - lipgloss.Height(helpMenu)
	view.todoTable.SetHeight(tableHeight)
	view.todoTable.SetWidth(view.width)

	var blocks []string

	blocks = append(blocks, view.todoTable.View())
	if view.inputState != hide {
		blocks = append(blocks, view.textInput.View())
	}
	blocks = append(blocks, helpMenu)

	return lipgloss.JoinVertical(lipgloss.Left, blocks...)
}

func NewTodoView() TodoView {
	todos := todos{}
	todos.add("Buy milk")
	todos.add("Buy bread and some other stuff\nCoffee")
	todos.add("Buy bread")
	todos.add("Buy bread")
	todos.add("Buy bread")
	todos.add("Buy bread")
	todos.toggle(3)

	ti := textinput.New()
	ti.Placeholder = "task..."
	ti.Prompt = "> "
	ti.CharLimit = 50
	ti.Width = 60

	return &viewState{
		todos:         todos,
		todoTable:     newTable(&todos, true),
		width:         0,
		height:        20,
		hideCompleted: true,
		inputState:    hide,
		textInput:     ti,
	}
}

func newTable(todos *todos, hideCompleted bool) table.Model {
	columns := []table.Column{
		{Title: "#", Width: 4},
		{Title: "Done", Width: 4},
		{Title: "Task", Width: 50},
		{Title: "CreatedAt", Width: 10},
		{Title: "CompletedAt", Width: 10},
	}

	rows := []table.Row{}

	t := *todos
	for idx, val := range t {
		if hideCompleted && val.completed {
			continue
		}

		completed := "❌"
		if val.completed {
			completed = "✅"
		}
		completedAt := "-"
		if val.completedAt != nil {
			completedAt = val.completedAt.Format(time.DateOnly)
		}

		rows = append(rows, table.Row{
			strconv.Itoa(idx),
			completed,
			val.title,
			val.createdAt.Format(time.DateOnly),
			completedAt,
		})
	}

	todoTable := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("#2E3434")).
		Background(lipgloss.Color("#A6E3A1")).
		Bold(false)
	todoTable.SetStyles(s)

	return todoTable
}
