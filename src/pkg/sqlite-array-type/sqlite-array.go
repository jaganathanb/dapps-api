package sqlite_array

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type SqliteStrArray []string

func (a *SqliteStrArray) Scan(value interface{}) error {
	if value == nil {
		return nil // case when value from the db was NULL
	}

	s, ok := value.(string)
	if !ok {
		return fmt.Errorf("failed to cast value to string: %v", value)
	}

	var val []string
	err := json.Unmarshal([]byte(s), &val)

	if err == nil {
		*a = val
	}

	return err
}

func (strarr SqliteStrArray) Value() (driver.Value, error) {
	if strarr != nil {
		resarr, err := json.Marshal(strarr)

		if err != nil {
			return nil, err
		}

		return string(resarr), nil
	}

	return nil, nil
}

func (SqliteStrArray) GormDataType() string {
	return "text[]"
}

func (SqliteStrArray) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	return "TEXT"
}

func (u *SqliteStrArray) String() string {
	val, _ := u.Value()

	return val.(string)
}

// func (u SqliteStrArray) MarshalJSON() ([]byte, error) {
// 	return json.Marshal(u.String())
// }

// func (u *SqliteStrArray) UnmarshalJSON(data []byte) error {
// 	// ignore null
// 	if string(data) == "null" {
// 		return nil
// 	}

// 	uu := string(data)

// 	if uu == "[]" {
// 		*u = SqliteStrArray{}
// 	} else {
// 		_ = json.Unmarshal(data, &u)
// 	}

// 	return nil
// }
