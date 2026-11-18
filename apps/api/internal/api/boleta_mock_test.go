package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBoletaMockEndpoints(t *testing.T) {
	mux := http.NewServeMux()
	RegistrarRutasBoletaMock(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/boletas?estado=EMITIDA", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("esperaba 200 en listar boletas, obtuvo %d", rec.Code)
	}

	bodyEmitir := []byte(`{"venta_id":"123e4567-e89b-12d3-a456-426614174000","rut_comprador":"12345678-5","items":[{"producto_id":"a1b2c3d4-e5f6-7890-abcd-ef1234567890","nombre_producto":"Item","cantidad":2,"precio_unitario":10000}]}`)
	reqSinKey := httptest.NewRequest(http.MethodPost, "/api/boletas", bytes.NewReader(bodyEmitir))
	recSinKey := httptest.NewRecorder()
	mux.ServeHTTP(recSinKey, reqSinKey)
	if recSinKey.Code != http.StatusBadRequest {
		t.Fatalf("esperaba 400 sin Idempotency-Key, obtuvo %d", recSinKey.Code)
	}

	reqConKey := httptest.NewRequest(http.MethodPost, "/api/boletas", bytes.NewReader(bodyEmitir))
	reqConKey.Header.Set("Idempotency-Key", "test-idem-1")
	recConKey := httptest.NewRecorder()
	mux.ServeHTTP(recConKey, reqConKey)
	if recConKey.Code != http.StatusCreated {
		t.Fatalf("esperaba 201 al emitir boleta, obtuvo %d", recConKey.Code)
	}

	reqDetalle := httptest.NewRequest(http.MethodGet, "/api/boletas/550e8400-e29b-41d4-a716-446655440000", nil)
	recDetalle := httptest.NewRecorder()
	mux.ServeHTTP(recDetalle, reqDetalle)
	if recDetalle.Code != http.StatusOK {
		t.Fatalf("esperaba 200 en obtener boleta, obtuvo %d", recDetalle.Code)
	}

	bodyAnular := []byte(`{"motivo":"Devolución de productos solicitada por cliente"}`)
	reqAnular := httptest.NewRequest(http.MethodPost, "/api/boletas/550e8400-e29b-41d4-a716-446655440000/anulacion", bytes.NewReader(bodyAnular))
	reqAnular.Header.Set("Idempotency-Key", "test-idem-2")
	recAnular := httptest.NewRecorder()
	mux.ServeHTTP(recAnular, reqAnular)
	if recAnular.Code != http.StatusOK {
		t.Fatalf("esperaba 200 al anular boleta, obtuvo %d", recAnular.Code)
	}

	reqAnularConflicto := httptest.NewRequest(http.MethodPost, "/api/boletas/660e8400-e29b-41d4-a716-446655440111/anulacion", bytes.NewReader(bodyAnular))
	reqAnularConflicto.Header.Set("Idempotency-Key", "test-idem-3")
	recAnularConflicto := httptest.NewRecorder()
	mux.ServeHTTP(recAnularConflicto, reqAnularConflicto)
	if recAnularConflicto.Code != http.StatusConflict {
		t.Fatalf("esperaba 409 al anular boleta ya anulada, obtuvo %d", recAnularConflicto.Code)
	}
}