package sms

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "strings"

    "github.com/sirupsen/logrus"

    "github.com/XingchengZhu/sms-webhook/config"
)

type Sender struct {
    client *http.Client
    cfg    config.Config
    log    *logrus.Logger
}

func NewSender(cfg config.Config, log *logrus.Logger) *Sender {
    return &Sender{
        client: &http.Client{},
        cfg:    cfg,
        log:    log,
    }
}

// SendMany: 对一条内容，发给多个通道
func (s *Sender) SendMany(channels []string, content string) []error {
    var errs []error
    for _, chName := range channels {
        if err := s.Send(chName, content); err != nil {
            s.log.WithError(err).Errorf("send to channel %s failed", chName)
            errs = append(errs, fmt.Errorf("%s: %v", chName, err))
        }
    }
    return errs
}

// Send: 发给单个通道
func (s *Sender) Send(channelName string, content string) error {
    ch, ok := s.cfg.Channels[channelName]
    if !ok {
        return fmt.Errorf("channel %s not found", channelName)
    }

    method := strings.ToUpper(ch.Method)
    if method == "" {
        method = "POST"
    }

    var body io.Reader

    // 根据类型拼 body
    switch ch.Type {
    case "json":
        payload := map[string]string{
            ch.CodeKey:    s.cfg.DefaultCode,
            ch.TargetKey:  s.cfg.DefaultTarget,
            ch.ContentKey: content,
        }
        for k, v := range ch.Static {
            payload[k] = v
        }
        b, _ := json.Marshal(payload)
        body = bytes.NewReader(b)

    case "form":
        form := url.Values{}
        form.Set(ch.CodeKey, s.cfg.DefaultCode)
        form.Set(ch.TargetKey, s.cfg.DefaultTarget)
        form.Set(ch.ContentKey, content)
        for k, v := range ch.Static {
            form.Set(k, v)
        }
        body = strings.NewReader(form.Encode())

    case "header":
        // header 类型就把内容当文本发，真正的 code/target 放在 header 里
        body = bytes.NewReader([]byte(content))

    default:
        // 默认就是发纯文本
        body = bytes.NewReader([]byte(content))
    }

    req, err := http.NewRequest(method, ch.URL, body)
    if err != nil {
        return err
    }

    // 设置 Content-Type
    switch ch.Type {
    case "json":
        req.Header.Set("Content-Type", "application/json")
    case "form":
        req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    default:
        req.Header.Set("Content-Type", "text/plain")
    }

    // header 类型要额外塞 code/target
    if ch.Type == "header" {
        req.Header.Set("X-SMS-Code", s.cfg.DefaultCode)
        req.Header.Set("X-SMS-Target", s.cfg.DefaultTarget)
    }

    // 统一加上通道自己定义的 headers
    for k, v := range ch.Headers {
        req.Header.Set(k, v)
    }

    s.log.WithFields(logrus.Fields{
        "channel": channelName,
        "type":    ch.Type,
        "url":     ch.URL,
        "method":  method,
    }).Info("sending sms request")

    resp, err := s.client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    respBody, _ := io.ReadAll(resp.Body)

    s.log.WithFields(logrus.Fields{
        "channel":    channelName,
        "statusCode": resp.StatusCode,
        "response":   string(respBody),
    }).Info("sms response")

    if resp.StatusCode >= 300 {
        return fmt.Errorf("channel %s returned %d", channelName, resp.StatusCode)
    }

    return nil
}
