package cli

import "errors"

type mockCliService struct {
	addCalledWith string
	shouldFail    bool
}

func (m *mockCliService) Add(title string) error {
	m.addCalledWith = title
	if m.shouldFail {
		return errors.New("mock storage failure")
	}
	return nil
}

// func (m *mockCliService) List() error { return nil }
