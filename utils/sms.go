package utils

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"sms-webhook/config"

	"github.com/sirupsen/logrus"
)

type SMSRequest struct {
	Code    string `json:"code"`
	Target  string `json:"target"`
	Content string `json:"content"`
}

func SendSMS(cfg config.Config, sms SMSRequest) error {
	jsonData, err := json.Marshal(sms)
	if err != nil {
		logrus.WithError(err).Error("Failed to marshal SMS request")
		return err
	}

	req, err := http.NewRequest("POST", cfg.SMSAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		logrus.WithError(err).Error("Failed to create new HTTP request")
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	logrus.WithField("url", cfg.SMSAPIURL).Info("Sending SMS request")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logrus.WithError(err).Error("Failed to send HTTP request")
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logrus.WithError(err).Error("Failed to read response body")
		return err
	}

	logrus.WithFields(logrus.Fields{
		"status_code": resp.StatusCode,
		"response":    string(body),
	}).Info("Received response from SMS API")

	return nil
}
