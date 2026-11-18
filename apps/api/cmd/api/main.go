package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/kamenrader36/boleta/config"      // Hito 2: reemplazar por internal/store con pgxpool
	"github.com/kamenrader36/boleta/controllers" // Hito 2: eliminar paquete de ejemplo "tareas"
	"github.com/kamenrader36/boleta/internal/api"
)

func setupRouter() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/tareas", controllers.GetTareas)         // Hito 2: eliminar rutas legacy de tareas
	mux.HandleFunc("POST /api/tareas", controllers.CreateTarea)      // Hito 2: eliminar rutas legacy de tareas
	mux.HandleFunc("OPTIONS /api/tareas", controllers.GetTareas)     // Hito 2: eliminar rutas legacy de tareas
	mux.HandleFunc("GET /api/v1/tareas", controllers.GetTareas)      // Hito 2: eliminar rutas legacy de tareas
	mux.HandleFunc("POST /api/v1/tareas", controllers.CreateTarea)   // Hito 2: eliminar rutas legacy de tareas
	mux.HandleFunc("OPTIONS /api/v1/tareas", controllers.GetTareas)  // Hito 2: eliminar rutas legacy de tareas

	api.RegistrarRutasBoletaMock(mux) // Hito 2: envolver mux con humago.New(mux, config) y middleware JWT/slog

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		if config.DB == nil {
			http.Error(w, "base de datos no inicializada", http.StatusServiceUnavailable)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := config.DB.PingContext(ctx); err != nil {
			http.Error(w, "base de datos no disponible", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready"))
	})
	return mux
}

func main() { // Hito 2: validar variables obligatorias al iniciar y abortar si falta alguna
	config.ConnectDB() // Hito 2: solo abrir pool sin crear tablas (las migraciones las corre el contenedor migrate)

	mux := setupRouter()

	fmt.Println("Servidor corriendo en el puerto 8080 (PostgreSQL)") // Hito 2: reemplazar fmt.Println por log/slog en formato JSON
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
}