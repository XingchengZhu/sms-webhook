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

type jsonPayload struct {
	Code    string `json:"code"`
	Target  string `json:"target"`
	Content string `json:"content"`
}

func NewJSONSender(name, url, code string) *JSONSender {
	if name == "" {
		name = "json"
	}
	return &JSONSender{
		name: name,
		URL:  url,
		Code: code,
	}
}

func (s *JSONSender) Name() string {
	return s.name
}

func (s *JSONSender) Send(target, content string) error {
	data := jsonPayload{
		Code:    s.Code,
		Target:  target,
		Content: content,
	}
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, s.URL, bytes.NewReader(b))
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
		"sender": "json",
		"status": resp.StatusCode,
		"resp":   string(respBody),
	}).Info("json sms response")

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("json sms: unexpected status %d", resp.StatusCode)
	}

	return nil
}
