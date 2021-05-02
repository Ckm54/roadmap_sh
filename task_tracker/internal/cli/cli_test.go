package cli

import (
	"errors"

	"github.com/ckm54/task_tracker/internal/task"
)

type mockCliService struct {
	calledAddTitle       string
	calledUpdateID       int
	calledUpdateTitle    string
	calledDeleteID       int
	calledUpdateStatusID int
	calledUpdateStatus   string
	calledListFilter     string
	shouldFail           bool
}

func (m *mockCliService) Add(title string) error {
	m.calledAddTitle = title
	if m.shouldFail {
		return errors.New("mock storage failure")
	}
	return nil
}

func (m *mockCliService) Update(id int, title string) error {
	m.calledUpdateID = id
	m.calledUpdateTitle = title
	if m.shouldFail {
		return errors.New("backend service failure")
	}
	return nil
}

func (m *mockCliService) UpdateStatus(id int, status string) error {
	m.calledUpdateStatusID = id
	m.calledUpdateStatus = status
	if m.shouldFail {
		return errors.New("backend service failure")
	}
	return nil
}

func (m *mockCliService) List(filter string) ([]task.TaskEntity, error) {
	m.calledListFilter = filter
	if m.shouldFail {
		return nil, errors.New("backend service failure")
	}
	return []task.TaskEntity{}, nil
}

func (m *mockCliService) Delete(id int) error {
	m.calledDeleteID = id
	if m.shouldFail {
		return errors.New("backend service failure")
	}
	return nil
}
