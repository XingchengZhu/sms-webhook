// sms/json_sender.go
package sms

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"

    "github.com/sirupsen/logrus"
)

type JSONSender struct {
    name   string
    URL    string
    Code   string
    Client *http.Client
}

func NewJSONSender(name, url, code string) *JSONSender {
    return &JSONSender{
        name: name,
        URL:  url,
        Code: code,
    }
}

func (s *JSONSender) Name() string { return s.name }

type jsonPayload struct {
    Code    string `json:"code"`
    Target  string `json:"target"`
    Content string `json:"content"`
}

func (s *JSONSender) Send(target, content string) error {
    body, err := json.Marshal(jsonPayload{
        Code:    s.Code,
        Target:  target,
        Content: content,
    })
    if err != nil {
        return err
    }

    req, err := http.NewRequest(http.MethodPost, s.URL, bytes.NewReader(body))
    if err != nil {
        return err
    }
    req.Header.Set("Content-Type", "application/json")

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
        "sender": s.name,
        "status": resp.StatusCode,
        "resp":   string(b),
    }).Info("json sms response")

    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        return fmt.Errorf("json sms: bad status %d", resp.StatusCode)
    }
    return nil
}
