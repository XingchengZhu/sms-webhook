package sms

// 所有短信通道必须实现这个接口
type Sender interface {
	Name() string
	Send(target, content string) error
}
