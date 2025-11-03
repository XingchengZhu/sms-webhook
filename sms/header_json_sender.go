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
    name      string
    URL       string
    Code      string
    APIKey    string
    HeaderKey string
    Client    *http.Client
}

func NewHeaderJSONSender(name, url, code, apiKey, headerKey string) *HeaderJSONSender {
    return &HeaderJSONSender{
        name:      name,
        URL:       url,
        Code:      code,
        APIKey:    apiKey,
        HeaderKey: headerKey,
    }
}

func (s *HeaderJSONSender) Name() string { return s.name }

type headerPayload struct {
    Code    string `json:"code"`
    Mobile  string `json:"mobile"`
    Content string `json:"content"`
}

func (s *HeaderJSONSender) Send(target, content string) error {
    b, err := json.Marshal(headerPayload{
        Code:    s.Code,
        Mobile:  target,
        Content: content,
    })
    if err != nil {
        return err
    }

    req, err := http.NewRequest(http.MethodPost, s.URL, bytes.NewReader(b))
    if err != nil {
        return err
    }
    req.Header.Set("Content-Type", "application/json")
    if s.HeaderKey != "" && s.APIKey != "" {
        req.Header.Set(s.HeaderKey, s.APIKey)
    }

    c := s.Client
    if c == nil {
        c = &http.Client{Timeout: 5 * time.Second}
    }

    resp, err := c.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    rb, _ := io.ReadAll(resp.Body)

    logrus.WithFields(logrus.Fields{
        "sender": s.name,
        "status": resp.StatusCode,
        "resp":   string(rb),
    }).Info("header json sms response")

    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        return fmt.Errorf("header json sms: bad status %d", resp.StatusCode)
    }
    return nil
}
