package sms

type Sender interface {
    Name() string
    Send(target, content string) error
}
