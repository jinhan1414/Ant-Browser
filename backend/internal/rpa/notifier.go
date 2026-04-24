package rpa

import "fmt"

type Notifier interface {
	Notify(title string, body string) error
}

type unsupportedNotifier struct{}

func NewUnsupportedNotifier() Notifier {
	return &unsupportedNotifier{}
}

func (n *unsupportedNotifier) Notify(title string, body string) error {
	return fmt.Errorf("rpa notifier not configured")
}
