package dto

import (
	"fmt"
	"reflect"
	"time"

	"github.com/ppkg/stark/enum"
)

type Date time.Time

func (s *Date) UnmarshalJSON(data []byte) (err error) {
	if len(data) == 0 {
		return nil
	}
	now, err := time.ParseInLocation(`"`+enum.DateTpl+`"`, string(data), time.Local)
	if err != nil {
		return err
	}
	*s = Date(now)
	return
}

func (s Date) MarshalJSON() ([]byte, error) {
	myTime := time.Time(s)
	if myTime.IsZero() {
		return []byte(`""`), nil
	}
	b := make([]byte, 0, len(enum.DateTpl)+2)
	b = append(b, '"')
	b = myTime.AppendFormat(b, enum.DateTpl)
	b = append(b, '"')
	return b, nil
}

func (s *Date) Scan(v interface{}) error {
	switch vt := v.(type) {
	case time.Time:
		*s = Date(vt)
	default:
		return fmt.Errorf("can not convert %+v to time.Time", v)
	}
	return nil
}

func (s Date) String() string {
	return time.Time(s).Format(enum.DateTimeTpl)
}

type DateTime time.Time

func (s *DateTime) UnmarshalJSON(data []byte) (err error) {
	if len(data) == 0 || reflect.DeepEqual(data, []byte("\"\"")) {
		return nil
	}
	now, err := time.ParseInLocation(`"`+enum.DateTimeTpl+`"`, string(data), time.Local)
	if err != nil {
		return err
	}
	*s = DateTime(now)
	return
}

func (s DateTime) MarshalJSON() ([]byte, error) {
	myTime := time.Time(s)
	if myTime.IsZero() {
		return []byte(`""`), nil
	}
	b := make([]byte, 0, len(enum.DateTimeTpl)+2)
	b = append(b, '"')
	b = myTime.AppendFormat(b, enum.DateTimeTpl)
	b = append(b, '"')
	return b, nil
}

func (s *DateTime) Scan(v interface{}) error {
	switch vt := v.(type) {
	case time.Time:
		*s = DateTime(vt)
	default:
		return fmt.Errorf("can not convert %+v to time.Time", v)
	}
	return nil
}

func (s DateTime) String() string {
	return time.Time(s).Format(enum.DateTimeTpl)
}
