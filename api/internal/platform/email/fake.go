package email

import "context"

type FakeSender struct {
	Messages []Message
	Err      error
}

func NewFakeSender() *FakeSender {
	return &FakeSender{}
}

func (f *FakeSender) Send(_ context.Context, msg Message) error {
	if f.Err != nil {
		return f.Err
	}
	f.Messages = append(f.Messages, msg)
	return nil
}
