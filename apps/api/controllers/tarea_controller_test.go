package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"tarea-api/config"
	"tarea-api/models"
)

func isDBAvailable() bool {
	if config.DB == nil {
		config.ConnectDB()
	}
	if config.DB == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	return config.DB.PingContext(ctx) == nil
}

func TestNewID(t *testing.T) {
	id1 := newID()
	id2 := newID()
	if len(id1) != 16 {
		t.Errorf("esperaba id de longitud 16, obtuve %d", len(id1))
	}
	if id1 == id2 {
		t.Errorf("se generaron dos IDs idénticos: %s", id1)
	}
}

func TestGetTareas_Options(t *testing.T) {
	req := httptest.NewRequest(http.MethodOptions, "/api/tareas", nil)
	rec := httptest.NewRecorder()

	GetTareas(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("esperaba status 200 para OPTIONS, obtuve %d", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Error("esperaba header Access-Control-Allow-Origin configurado")
	}
}

func TestCreateTarea_Options(t *testing.T) {
	req := httptest.NewRequest(http.MethodOptions, "/api/tareas", nil)
	rec := httptest.NewRecorder()

	CreateTarea(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("esperaba status 200 para OPTIONS, obtuve %d", rec.Code)
	}
}

func TestCreateTarea_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/tareas", bytes.NewBufferString("{invalido}"))
	rec := httptest.NewRecorder()

	CreateTarea(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("esperaba status 400 para JSON inválido, obtuve %d", rec.Code)
	}
}

func TestCreateTarea_EmptyTitulo(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/tareas", bytes.NewBufferString(`{"titulo": ""}`))
	rec := httptest.NewRecorder()

	CreateTarea(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("esperaba status 400 para título vacío, obtuve %d", rec.Code)
	}
}

func TestCreateTarea_Success(t *testing.T) {
	if !isDBAvailable() {
		t.Skip("PostgreSQL no disponible en entorno local, omitiendo prueba con DB")
	}

	body := `{"titulo": "Tarea creada en test", "estado": "pendiente"}`
	req := httptest.NewRequest(http.MethodPost, "/api/tareas", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	CreateTarea(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("esperaba status 201 Created, obtuve %d: %s", rec.Code, rec.Body.String())
	}

	var creada models.Tarea
	if err := json.NewDecoder(rec.Body).Decode(&creada); err != nil {
		t.Fatalf("error decodificando respuesta JSON: %v", err)
	}
	if creada.Titulo != "Tarea creada en test" {
		t.Errorf("esperaba título 'Tarea creada en test', obtuve '%s'", creada.Titulo)
	}
}

func TestCreateTarea_DefaultEstado(t *testing.T) {
	if !isDBAvailable() {
		t.Skip("PostgreSQL no disponible en entorno local, omitiendo prueba con DB")
	}

	body := `{"titulo": "Tarea sin estado explícito"}`
	req := httptest.NewRequest(http.MethodPost, "/api/tareas", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	CreateTarea(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("esperaba status 201 Created, obtuve %d: %s", rec.Code, rec.Body.String())
	}

	var creada models.Tarea
	_ = json.NewDecoder(rec.Body).Decode(&creada)
	if creada.Estado != "pendiente" {
		t.Errorf("esperaba estado por defecto 'pendiente', obtuve '%s'", creada.Estado)
	}
}

func TestGetTareas_Success(t *testing.T) {
	if !isDBAvailable() {
		t.Skip("PostgreSQL no disponible en entorno local, omitiendo prueba con DB")
	}

	req := httptest.NewRequest(http.MethodGet, "/api/tareas", nil)
	rec := httptest.NewRecorder()

	GetTareas(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperaba status 200 OK, obtuve %d: %s", rec.Code, rec.Body.String())
	}

	var tareas []models.Tarea
	if err := json.NewDecoder(rec.Body).Decode(&tareas); err != nil {
		t.Fatalf("error decodificando tareas JSON: %v", err)
	}
}
