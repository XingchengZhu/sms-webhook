// sms/form_sender.go
package sms

import (
    "fmt"
    "io"
    "net/http"
    "net/url"
    "time"

    "github.com/sirupsen/logrus"
)

type FormSender struct {
    URL       string
    CodeField string // 比如 "code"
    PhoneField string // 比如 "phone" or "target"
    ContentField string // 比如 "msg" or "content"
    CodeValue string
    Client    *http.Client
}

func (s *FormSender) Send(target, content string) error {
    form := url.Values{}
    if s.CodeField != "" && s.CodeValue != "" {
        form.Set(s.CodeField, s.CodeValue)
    }
    form.Set(s.PhoneField, target)
    form.Set(s.ContentField, content)

    req, err := http.NewRequest(http.MethodPost, s.URL, strings.NewReader(form.Encode()))
    if err != nil {
        return err
    }
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

    client := s.Client
    if client == nil {
        client = &http.Client{Timeout: 5 * time.Second}
    }

    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    b, _ := io.ReadAll(resp.Body)

    logrus.WithFields(logrus.Fields{
        "status": resp.StatusCode,
        "resp":   string(b),
    }).Info("form sms response")

    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        return fmt.Errorf("form sms: unexpected status %d", resp.StatusCode)
    }
    return nil
}
