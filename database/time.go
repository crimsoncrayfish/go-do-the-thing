package database

import (
	"encoding/json"
	"strings"
	"time"
)

type SqLiteTime struct {
	time.Time
}

func (t *SqLiteTime) Scan(v interface{}) error {
	vt, err := time.Parse(DateTimeFormat, v.(string))
	if err != nil {
		return err
	}
	*t = SqLiteTime{vt}

	return nil
}

func (t *SqLiteTime) Format(formatString string) string {
	return t.Time.Format(formatString)
}

func (t *SqLiteTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Time)
}

const DateTimeFormat = "2006-01-02 15:04:05"
const DateFormat = "2006-01-02"

func (t *SqLiteTime) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), "\"")
	if s == "null" {
		*t = SqLiteTime{time.Time{}}
		return nil
	}
	temp, err := time.Parse(DateTimeFormat, s)
	if err != nil {
		return err
	}
	*t = SqLiteTime{temp}
	return nil
}

func (t *SqLiteTime) String() string {
	return t.Time.Format(DateTimeFormat)
}
func (t *SqLiteTime) StringF(format string) string {
	return t.Time.Format(format)
}

func (t *SqLiteTime) BeforeNow() bool {
	return t.Time.Before(time.Now())
}
