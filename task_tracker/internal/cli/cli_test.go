package cli

import "errors"

type mockCliService struct {
	calledAddTitle    string
	calledUpdateID    int
	calledUpdateTitle string
	shouldFail        bool
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
