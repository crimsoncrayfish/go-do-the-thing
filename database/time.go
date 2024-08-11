package database

import (
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type SqLiteTime struct {
	*time.Time
}

func (t *SqLiteTime) Scan(v interface{}) error {
	if v.(string) == "" {
		return nil
	}
	vt, err := time.Parse(DateTimeFormat, v.(string))
	if err != nil {
		return err
	}
	*t = SqLiteTime{&vt}

	return nil
}

func (t *SqLiteTime) Format(formatString string) (string, error) {
	if t.Time == nil {
		return "", errors.New("SqLiteTime is null")
	}
	return t.Time.Format(formatString), nil
}

func (t *SqLiteTime) MarshalJSON() ([]byte, error) {
	if t.Time == nil {
		return json.Marshal("")
	}
	return json.Marshal(t.Time)
}

const DateTimeFormat = "2006-01-02 15:04:05"
const DateFormat = "2006-01-02"

func (t *SqLiteTime) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), "\"")
	if s == "null" || s == "" {
		*t = SqLiteTime{nil}
		return nil
	}
	temp, err := time.Parse(DateTimeFormat, s)
	if err != nil {
		return err
	}
	*t = SqLiteTime{&temp}
	return nil
}

func (t *SqLiteTime) String() string {
	if t.Time == nil {
		return ""
	}
	return t.Time.Format(DateTimeFormat)
}
func (t *SqLiteTime) StringF(format string) string {
	if t.Time == nil {
		return ""
	}
	return t.Time.Format(format)
}

func (t *SqLiteTime) BeforeNow() (bool, error) {
	if t.Time == nil {
		return false, errors.New("No date configured")
	}
	return t.Time.Before(time.Now()), nil
}
