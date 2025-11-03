// sms/sender.go
package sms

// Sender 是所有短信实现要满足的接口
type Sender interface {
    Send(target, content string) error
}
