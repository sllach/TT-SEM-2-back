package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type PropiedadPerceptiva struct {
	Nombre string `json:"nombre"`
	Valor  string `json:"valor"`
}

type JSONPerceptivas []PropiedadPerceptiva

func (j JSONPerceptivas) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *JSONPerceptivas) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("error de tipo: se esperaba []byte")
	}
	return json.Unmarshal(bytes, &j)
}
