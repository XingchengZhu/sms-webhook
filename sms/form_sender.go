package sms

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type FormSender struct {
	name         string
	URL          string
	CodeField    string
	PhoneField   string
	ContentField string
	CodeValue    string
	Client       *http.Client
}

func NewFormSender(
	name, urlStr, codeField, phoneField, contentField, codeValue string,
) *FormSender {
	if name == "" {
		name = "form"
	}
	return &FormSender{
		name:         name,
		URL:          urlStr,
		CodeField:    codeField,
		PhoneField:   phoneField,
		ContentField: contentField,
		CodeValue:    codeValue,
	}
}

func (s *FormSender) Name() string {
	return s.name
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

	respBody, _ := io.ReadAll(resp.Body)

	logrus.WithFields(logrus.Fields{
		"sender": "form",
		"status": resp.StatusCode,
		"resp":   string(respBody),
	}).Info("form sms response")

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("form sms: unexpected status %d", resp.StatusCode)
	}

	return nil
}
