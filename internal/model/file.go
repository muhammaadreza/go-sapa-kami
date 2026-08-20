package model

import (
	"encoding/json"
	"errors"

	"database/sql/driver"
)

// File structure
type Attachment struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
	Type string `json:"type"`
	Path string `json:"path"`
}

// Attachments structure
type Attachments []Attachment

// Scanner and Valuer
// Single File Scanner
func (file *Attachment) Scan(value interface{}) error {
	if value == nil { return nil }

	b, ok := value.([]byte)
	if !ok {
		s, okString := value.(string)
		if !okString { return errors.New("type assertion to []byte or string failed") }

		b = []byte(s)
	}

	return json.Unmarshal(b, file)
}

// Single File Valuer
func (attachment Attachment) Value() (driver.Value, error) {
	if attachment.Path == "" { return nil, nil }

	return json.Marshal(attachment)
}

// Multiple Attachments Scanner
func (attachments *Attachments) Scan(value interface{}) error {
	if value == nil {
		*attachments = make(Attachments, 0)
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		s, okString := value.(string)
		if !okString { return errors.New("type assertion to []byte or string failed") }

		b = []byte(s)
	}

	return json.Unmarshal(b, attachments)
}

// Multiple Attachments Valuer
func (attachments Attachments) Value() (driver.Value, error) {
	if len(attachments) == 0 { return "[]", nil }

	return json.Marshal(attachments)
}
