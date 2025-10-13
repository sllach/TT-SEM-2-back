package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// StringArray es un wrapper para []string que se almacena como JSONB en Postgres
type StringArray []string

// Value implementa driver.Valuer para marshal a JSON
func (a StringArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return json.Marshal([]string{}) // Retorna array vacío si nil
	}
	return json.Marshal(a)
}

// Scan implementa sql.Scanner para unmarshal desde JSONB
func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = StringArray{}
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("tipo incompatible para StringArray: esperado []byte")
	}
	return json.Unmarshal(b, a)
}
