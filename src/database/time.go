package database

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"go-do-the-thing/src/helpers/constants"
	"strings"
	"time"
)

type SqLiteTime struct {
	*time.Time
}

func NewSqliteTime(time time.Time) *SqLiteTime {
	return &SqLiteTime{&time}
}

func SqLiteNow() *SqLiteTime {
	now := time.Now()
	return &SqLiteTime{&now}
}

func (t *SqLiteTime) Scan(v interface{}) error {
	if v.(int64) < 0 {
		return nil
	}
	vt := time.Unix(v.(int64), 0)
	*t = SqLiteTime{&vt}
	return nil
}

func (t *SqLiteTime) Value() (driver.Value, error) {
	if t.Time == nil {
		return int64(0), nil
	}
	return t.Time.Unix(), nil
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

func (t *SqLiteTime) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), "\"")
	if s == "null" || s == "" {
		*t = SqLiteTime{nil}
		return nil
	}
	temp, err := time.Parse(constants.DateTimeFormat, s)
	if err != nil {
		return err
	}
	*t = SqLiteTime{&temp}
	return nil
}

func (t *SqLiteTime) ShortString() string {
	if t.Time == nil {
		return ""
	}
	return t.Time.Format(constants.DateFormat)
}

func (t *SqLiteTime) String() string {
	if t.Time == nil {
		return ""
	}
	return t.Time.Format(constants.DateTimeFormat)
}

func (t *SqLiteTime) Unix() int64 {
	if t.Time == nil {
		return 0
	}
	return t.Time.Unix()
}

func (t *SqLiteTime) StringF(format string) string {
	if t.Time == nil {
		return ""
	}
	return t.Time.Format(format)
}

func (t *SqLiteTime) Before(other *SqLiteTime) bool {
	return t.Time.Before(*other.Time)
}

func (t *SqLiteTime) BeforeNow() (bool, error) {
	if t.Time == nil {
		return false, errors.New("No date configured")
	}
	return t.Time.Before(time.Now()), nil
}
