// sms/header_json_sender.go
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

type HeaderJSONSender struct {
    URL       string
    Code      string
    APIKey    string
    HeaderKey string // 比如 "X-API-KEY"
    Client    *http.Client
}

type headerJSONPayload struct {
    Code    string `json:"code"`
    Mobile  string `json:"mobile"`  // 有的叫 mobile
    Content string `json:"content"`
}

func (s *HeaderJSONSender) Send(target, content string) error {
    body, err := json.Marshal(headerJSONPayload{
        Code:    s.Code,
        Mobile:  target,
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
    if s.HeaderKey != "" && s.APIKey != "" {
        req.Header.Set(s.HeaderKey, s.APIKey)
    }

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
    }).Info("header json sms response")

    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        return fmt.Errorf("header json sms: unexpected status %d", resp.StatusCode)
    }
    return nil
}
