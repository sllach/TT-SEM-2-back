package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type PropiedadEmocional struct {
	Nombre string `json:"nombre"`
	Valor  string `json:"valor"`
}

type JSONEmocionales []PropiedadEmocional

func (j JSONEmocionales) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *JSONEmocionales) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("error de tipo: se esperaba []byte")
	}
	return json.Unmarshal(bytes, &j)
}
