package models

import "time"

type Tarea struct {
	ID            string    `json:"id"`
	Titulo        string    `json:"titulo"`
	Estado        string    `json:"estado"`
	FechaCreacion time.Time `json:"fecha_creacion,omitempty"`
}
