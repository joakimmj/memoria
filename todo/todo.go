package todo

import (
	"errors"
	"time"
)

type todo struct {
	title       string
	createdAt   time.Time
	completed   bool
	completedAt *time.Time
}

type todos []todo

func (t *todos) add(title string) {
	todo := todo{
		title:       title,
		createdAt:   time.Now(),
		completed:   false,
		completedAt: nil,
	}

	updatedTodo := todos{}
	updatedTodo = append(updatedTodo, todo)
	updatedTodo = append(updatedTodo, *t...)
	*t = updatedTodo
}

func (todos *todos) validateIndex(index int) error {
	if index < 0 || index >= len(*todos) {
		err := errors.New("Invalid index")
		return err
	}
	return nil
}

func (todos *todos) delete(index int) error {
	t := *todos

	if err := t.validateIndex(index); err != nil {
		return err
	}

	*todos = append(t[:index], t[index+1:]...)
	return nil
}

func (todos *todos) toggle(index int) error {
	t := *todos

	if err := t.validateIndex(index); err != nil {
		return err
	}

	isCompleted := !t[index].completed

	if isCompleted {
		completionTime := time.Now()
		t[index].completedAt = &completionTime
	} else {
		t[index].completedAt = nil
	}

	t[index].completed = isCompleted
	return nil
}

func (todos *todos) edit(index int, title string) error {
	t := *todos

	if err := t.validateIndex(index); err != nil {
		return err
	}

	t[index].title = title
	return nil
}
