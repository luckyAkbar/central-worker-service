package helper

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/sirupsen/logrus"
)

type Action string

const (
	ShortURL Action = "shorturl"
	Expiry   Action = "expiry"
	Delete   Action = "delete"
)

type Format string

const (
	JSON Format = "json"
)

type ExpiryOpts string

const (
	Clock ExpiryOpts = "clock"
	Click ExpiryOpts = "click"
)

type ExpiryAgeMode string

const (
	Minutes ExpiryAgeMode = "min"
	Hours   ExpiryAgeMode = "hr"
	Days    ExpiryAgeMode = "day"
)

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type BasicCreds struct {
	Action    Action `json:"action"`
	Format    Format `json:"format"`
	Signature string `json:"signature"`
}

func (bc *BasicCreds) ToKeyValue() map[string]string {
	return map[string]string{
		"action":    string(bc.Action),
		"format":    string(bc.Format),
		"signature": bc.Signature,
	}
}

type ShortingInput struct {
	Keyword string `json:"keyword,omitempty"`
	Title   string `json:"title,omitempty"`
	URL     string `json:"url"`
}

func (si *ShortingInput) ToKeyValue() map[string]string {
	m := map[string]string{
		"keyword": si.Keyword,
		"title":   si.Title,
		"url":     si.URL,
	}

	return m
}

type ActionSetExpiryInput struct {
	ShortURL string
	Expiry   ExpiryOpts
	AgeMod   ExpiryAgeMode
	Age      int64
	Count    int64
	Postx    string
}

func (asei *ActionSetExpiryInput) ToKeyValue() map[string]string {
	return map[string]string{
		"shorturl": asei.ShortURL,
		"expiry":   string(asei.Expiry),
		"ageMod":   string(asei.AgeMod),
		"age":      fmt.Sprintf("%d", asei.Age),
		"count":    fmt.Sprintf("%d", asei.Count),
		"postx":    asei.Postx,
	}
}

type ActionShortURLResponse struct {
	Status     string `json:"status,omitempty"`
	Code       string `json:"code,omitempty"`
	ErrorCode  string `json:"errorCode,omitempty"`
	Message    string `json:"message,omitempty"`
	ShortURL   string `json:"shorturl,omitempty"`
	StatusCode any    `json:"statusCode,omitempty"`
}

type ActionSetExpiryResponse struct {
	Expiry     string `json:"expiry"`
	ExpiryType string `json:"expiry_type"`
	ExpiryLife string `json:"expiry_life"`
	StatusCode string `json:"status_code"`
	Message    string `json:"message"`
	ShortURL   string `json:"short_url"`
	URL        string `json:"url"`
	Title      string `json:"title"`
}

type YourlsUtil struct {
	baseURL    string
	signature  string
	httpClient *http.Client
}

func NewYourlsUtil(baseURL, signature string, httpClient *http.Client) *YourlsUtil {
	return &YourlsUtil{
		baseURL:    baseURL,
		signature:  signature,
		httpClient: httpClient,
	}
}

func (y *YourlsUtil) Shorten(ctx context.Context, input *ShortingInput) (string, error) {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"input": input,
		"func":  "yourls.Shorten",
	})

	bc := &BasicCreds{
		Action:    ShortURL,
		Format:    JSON,
		Signature: y.signature,
	}

	q := url.Values{}
	for k, v := range bc.ToKeyValue() {
		if v != "" {
			q.Add(k, v)
		}
	}

	for k, v := range input.ToKeyValue() {
		if v != "" {
			q.Add(k, v)
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, y.baseURL, nil)
	if err != nil {
		logger.WithError(err).Error("failed to create request")
		return "", err
	}

	req.URL.RawQuery = q.Encode()

	resp, err := y.httpClient.Do(req)
	if err != nil {
		logger.WithError(err).Error("failed to do request")
		return "", err
	}

	defer resp.Body.Close()

	var response *ActionShortURLResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		logger.WithError(err).Error("failed to decode response body")
		return "", err
	}

	if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusOK {
		if response.Code == "" || response.Code == "error:url" {
			return response.ShortURL, nil
		}
	}
	logger.WithField("status_code", resp.StatusCode).Error("unexpected status code from yourls")
	return "", fmt.Errorf("unexpected status code from yourls: error code: %s; error message: %s", response.Code, response.Message)
}

func (y *YourlsUtil) SetExpiry(ctx context.Context, input *ActionSetExpiryInput) (*ActionSetExpiryResponse, error) {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"input": input,
		"func":  "yourls.SetExpiry",
	})

	bc := &BasicCreds{
		Action:    Expiry,
		Format:    JSON,
		Signature: y.signature,
	}

	q := url.Values{}
	for k, v := range bc.ToKeyValue() {
		if v != "" {
			q.Add(k, v)
		}
	}

	for k, v := range input.ToKeyValue() {
		if v != "" {
			q.Add(k, v)
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, y.baseURL, nil)
	if err != nil {
		logger.WithError(err).Error("failed to create request")
		return nil, err
	}

	req.URL.RawQuery = q.Encode()

	resp, err := y.httpClient.Do(req)
	if err != nil {
		logger.WithError(err).Error("failed to do request for expiry yourls")
		return nil, err
	}

	defer resp.Body.Close()

	switch resp.StatusCode {
	default:
		logger.WithField("status_code", resp.StatusCode).Error("unexpected status code from yourls")

		var bd *ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&bd); err != nil {
			logger.WithError(err).Error("failed to decode response body")
			return nil, err
		}

		return nil, fmt.Errorf("unexpected status code from yourls: error code: %s; error message: %s", bd.Code, bd.Message)

	case http.StatusOK:
		break
	}

	var response *ActionSetExpiryResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		logger.WithError(err).Error("failed to decode response body")
		return nil, err
	}

	return response, nil
}
