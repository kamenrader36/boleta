package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthz(t *testing.T) {
	mux := setupRouter()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperaba status 200 para /healthz, obtuve %d", rec.Code)
	}

	body, err := io.ReadAll(rec.Body)
	if err != nil {
		t.Fatalf("error leyendo body: %v", err)
	}

	if string(body) != "ok" {
		t.Errorf("esperaba 'ok', obtuve '%s'", string(body))
	}
}

func TestReadyz(t *testing.T) {
	mux := setupRouter()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	// Sin DB conectada da 503, con DB da 200
	if rec.Code != http.StatusOK && rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("esperaba status 200 o 503 para /readyz, obtuve %d", rec.Code)
	}
}

func TestRouterRoutes(t *testing.T) {
	mux := setupRouter()

	// Probar OPTIONS en /api/tareas y /api/v1/tareas
	for _, path := range []string{"/api/tareas", "/api/v1/tareas"} {
		req := httptest.NewRequest(http.MethodOptions, path, nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("esperaba status 200 para OPTIONS %s, obtuve %d", path, rec.Code)
		}
	}
}
