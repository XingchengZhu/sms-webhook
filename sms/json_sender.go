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
    URL  string
    Code string
    // 也可以再加别的固定字段
    Client *http.Client
}

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

    respBody, _ := io.ReadAll(resp.Body)

    logrus.WithFields(logrus.Fields{
        "status": resp.StatusCode,
        "resp":   string(respBody),
    }).Info("json sms response")

    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        return fmt.Errorf("json sms: unexpected status %d", resp.StatusCode)
    }
    return nil
}
