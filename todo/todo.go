package todo

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
	gap "github.com/muesli/go-app-paths"
)

type todo struct {
	id          uint
	task        string
	createdAt   time.Time
	completed   bool
	completedAt *time.Time
}

type todos struct {
	db *sql.DB
}

func openDB() (*todos, error) {
	path := setupPath()
	db, err := sql.Open("sqlite3", filepath.Join(path, "memoria-todos.db"))
	if err != nil {
		return nil, err
	}
	t := todos{db}
	return t.setup()
}

func (t *todos) getAll() ([]todo, error) {
	var todos []todo
	rows, err := t.db.Query("SELECT * FROM todos ORDER BY id DESC")
	if err != nil {
		return todos, fmt.Errorf("unable to get values: %w", err)
	}
	for rows.Next() {
		var todo todo
		err = rows.Scan(
			&todo.id,
			&todo.completed,
			&todo.task,
			&todo.createdAt,
			&todo.completedAt,
		)
		if err != nil {
			return todos, err
		}
		todos = append(todos, todo)
	}
	return todos, err
}

func (t *todos) get(id uint) (todo, error) {
	var todo todo
	err := t.db.QueryRow("SELECT * FROM todos WHERE id = ?", id).
		Scan(
			&todo.id,
			&todo.completed,
			&todo.task,
			&todo.createdAt,
			&todo.completedAt,
		)
	return todo, err
}

func (t *todos) add(task string) error {
	// We don't care about the returned values, so we're using Exec. If we
	// wanted to reuse these statements, it would be more efficient to use
	// prepared statements. Learn more:
	// https://go.dev/doc/database/prepared-statements
	_, err := t.db.Exec(
		"INSERT INTO todos(completed, task, createdAt) VALUES( ?, ?, ?)",
		false,
		task,
		time.Now(),
	)
	return err
}

func (t *todos) delete(id uint) error {
	_, err := t.db.Exec("DELETE FROM todos WHERE id = ?", id)
	return err
}

func (t *todos) edit(id uint, task string) error {
	_, err := t.db.Exec(
		"UPDATE todos SET task = ? WHERE id = ?",
		task,
		id,
	)
	return err
}

func (t *todos) toggle(id uint) error {
	todo, err := t.get(id)
	if err != nil {
		return err
	}

	isCompleted := !todo.completed
	var completedAt time.Time

	if isCompleted {
		completedAt = time.Now()
	}

	_, err = t.db.Exec(
		"UPDATE todos SET completed = ?, completedAt = ? WHERE id = ?",
		isCompleted,
		completedAt,
		id,
	)
	return err
}

// SETUP DB
func (t *todos) setup() (*todos, error) {
	if t.tableExists("todos") {
		return t, nil
	}
	if err := t.createTable(); err != nil {
		return nil, err
	}
	return t, nil
}

func (t *todos) tableExists(name string) bool {
	if _, err := t.db.Query("SELECT * FROM todos"); err == nil {
		log.Print("SD", err)
		return true
	}
	return false
}

func (t *todos) createTable() error {
	_, err := t.db.Exec(`CREATE TABLE "todos" ( "id" INTEGER, "completed" BOOLEAN NOT NULL CHECK (completed IN (0, 1)), "task" TEXT NOT NULL, "createdAt" DATETIME NOT NULL, "completedAt" DATETIME, PRIMARY KEY("id" AUTOINCREMENT))`)
	return err
}

// setupPath uses XDG to create the necessary data dirs for the program.
func setupPath() string {
	// get XDG paths
	scope := gap.NewScope(gap.User, "todos")
	dirs, err := scope.DataDirs()
	if err != nil {
		log.Fatal(err)
	}
	// create the app base dir, if it doesn't exist
	var taskDir string
	if len(dirs) > 0 {
		taskDir = dirs[0]
	} else {
		taskDir, _ = os.UserHomeDir()
	}
	if err := initTaskDir(taskDir); err != nil {
		log.Fatal(err)
	}
	return taskDir
}

func initTaskDir(path string) error {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return os.Mkdir(path, 0o770)
		}
		return err
	}
	return nil
}
