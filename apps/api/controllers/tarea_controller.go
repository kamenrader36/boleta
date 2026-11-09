package controllers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"tarea-api/config"
	"tarea-api/models"
)

func setCORS(w http.ResponseWriter) {
	origin := os.Getenv("ALLOWED_ORIGIN")
	if origin == "" {
		origin = "http://localhost:5173"
	}
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func newID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func GetTareas(w http.ResponseWriter, r *http.Request) {
	setCORS(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	rows, err := config.DB.QueryContext(ctx, "SELECT id, titulo, estado, fecha_creacion FROM tareas ORDER BY fecha_creacion DESC")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tareas []models.Tarea
	for rows.Next() {
		var t models.Tarea
		if err := rows.Scan(&t.ID, &t.Titulo, &t.Estado, &t.FechaCreacion); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tareas = append(tareas, t)
	}

	if tareas == nil {
		tareas = []models.Tarea{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(tareas)
}

func CreateTarea(w http.ResponseWriter, r *http.Request) {
	setCORS(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req struct {
		Titulo string `json:"titulo"`
		Estado string `json:"estado"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	if req.Titulo == "" {
		http.Error(w, "El campo 'titulo' es obligatorio", http.StatusBadRequest)
		return
	}

	if req.Estado == "" {
		req.Estado = "pendiente"
	}

	t := models.Tarea{
		ID:            newID(),
		Titulo:        req.Titulo,
		Estado:        req.Estado,
		FechaCreacion: time.Now(),
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	_, err := config.DB.ExecContext(ctx, "INSERT INTO tareas (id, titulo, estado, fecha_creacion) VALUES ($1, $2, $3, $4)", t.ID, t.Titulo, t.Estado, t.FechaCreacion)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(t)
}
