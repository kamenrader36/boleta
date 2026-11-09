package models

import (
	"testing"
	"time"
)

func TestTareaModel(t *testing.T) {
	now := time.Now()
	tarea := Tarea{
		ID:            "test-1",
		Titulo:        "Prueba de unidad",
		Estado:        "pendiente",
		FechaCreacion: now,
	}

	if tarea.ID != "test-1" || tarea.Titulo != "Prueba de unidad" || tarea.Estado != "pendiente" || tarea.FechaCreacion != now {
		t.Errorf("atributos del modelo Tarea no coinciden")
	}
}
