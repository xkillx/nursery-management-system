package email

import "context"

type Sender interface {
	Send(ctx context.Context, msg Message) error
}

type Message struct {
	To          string
	Subject     string
	Text        string
	HTML        string
	Attachments []Attachment
}

type Attachment struct {
	Filename    string
	ContentType string
	Data        []byte
}
