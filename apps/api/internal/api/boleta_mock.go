package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type ItemBoleta struct {
	ProductoID     string `json:"producto_id"`     // Hito 2: agregar tags Huma (format:"uuid") para autogenerar OpenAPI
	NombreProducto string `json:"nombre_producto"` // Hito 2: agregar tag minLength:"1" maxLength:"200" en Huma
	Cantidad       int32  `json:"cantidad"`        // Hito 2: agregar tag minimum:"1" en Huma
	PrecioUnitario int64  `json:"precio_unitario"` // Hito 2: mantener int64 (bigint en BD), agregar minimum:"1"
	Subtotal       int64  `json:"subtotal"`
}

type Boleta struct {
	ID              string       `json:"id"`
	Folio           int64        `json:"folio"`
	VentaID         string       `json:"venta_id"`
	RutComprador    string       `json:"rut_comprador,omitempty"` // Hito 2: agregar pattern de RUT chileno en Huma
	Estado          string       `json:"estado"`                  // Hito 2: restringir con enum:"EMITIDA,ANULADA" en Huma
	MontoNeto       int64        `json:"monto_neto"`
	MontoIVA        int64        `json:"monto_iva"`
	MontoTotal      int64        `json:"monto_total"`
	FechaEmision    string       `json:"fecha_emision"`    // Hito 2: cambiar a time.Time para mapear timestamptz de PostgreSQL
	FechaAnulacion  *string      `json:"fecha_anulacion"`  // Hito 2: cambiar a *time.Time para mapear timestamptz nullable
	MotivoAnulacion *string      `json:"motivo_anulacion"`
	Items           []ItemBoleta `json:"items"`
}

type EmitirBoletaRequest struct {
	VentaID      string       `json:"venta_id"`      // Hito 2: envolver en struct Huma con Header Idempotency-Key y Body
	RutComprador string       `json:"rut_comprador"`
	Items        []ItemBoleta `json:"items"` // Hito 2: agregar minItems:"1" para validar en contrato
}

type AnularBoletaRequest struct {
	Motivo string `json:"motivo"` // Hito 2: envolver en struct Huma con Path id, Header Idempotency-Key y Body
}

type AnularBoletaResponse struct {
	Boleta                    Boleta `json:"boleta"`
	StockDevuelto             bool   `json:"stock_devuelto"`
	ReferenciaDevolucionStock string `json:"referencia_devolucion_stock"`
	FechaAnulacion            string `json:"fecha_anulacion"` // Hito 2: cambiar a time.Time
}

type ProblemaRFC9457 struct {
	Type     string `json:"type"`   // Hito 2: eliminar este struct cuando Huma v2 maneje RFC 9457 nativamente
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail"`
	Instance string `json:"instance"`
}

func responderProblema(w http.ResponseWriter, status int, title, detail, instance string) { // Hito 2: reemplazar por huma.Error400BadRequest, etc.
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ProblemaRFC9457{
		Type:     "about:blank",
		Title:    title,
		Status:   status,
		Detail:   detail,
		Instance: instance,
	})
}

func boletasMockBase() []Boleta { // Hito 2: eliminar datos simulados al conectar internal/service + PostgreSQL
	motivo := "Devolución total de productos por parte del cliente"
	fechaAnul := "2026-11-17T16:20:00-03:00"
	return []Boleta{
		{
			ID:           "550e8400-e29b-41d4-a716-446655440000",
			Folio:        1001,
			VentaID:      "123e4567-e89b-12d3-a456-426614174000",
			RutComprador: "12345678-5",
			Estado:       "EMITIDA",
			MontoNeto:    42017,
			MontoIVA:     7983,
			MontoTotal:   50000,
			FechaEmision: "2026-11-17T15:04:05-03:00",
			Items: []ItemBoleta{
				{
					ProductoID:     "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
					NombreProducto: "Teclado Mecánico RGB",
					Cantidad:       2,
					PrecioUnitario: 25000,
					Subtotal:       50000,
				},
			},
		},
		{
			ID:              "660e8400-e29b-41d4-a716-446655440111",
			Folio:           1002,
			VentaID:         "223e4567-e89b-12d3-a456-426614174111",
			RutComprador:    "98765432-1",
			Estado:          "ANULADA",
			MontoNeto:       21008,
			MontoIVA:        3992,
			MontoTotal:      25000,
			FechaEmision:    "2026-11-17T14:00:00-03:00",
			FechaAnulacion:  &fechaAnul,
			MotivoAnulacion: &motivo,
			Items: []ItemBoleta{
				{
					ProductoID:     "b2c3d4e5-f6a7-8901-bcde-f12345678901",
					NombreProducto: "Mouse Gamer Óptico",
					Cantidad:       1,
					PrecioUnitario: 25000,
					Subtotal:       25000,
				},
			},
		},
	}
}

func RegistrarRutasBoletaMock(mux *http.ServeMux) { // Hito 2: cambiar firma para recibir huma.API y registrar operaciones con permisos JWT
	mux.HandleFunc("GET /api/boletas", ListarBoletasMock)
	mux.HandleFunc("POST /api/boletas", EmitirBoletaMock)
	mux.HandleFunc("GET /api/boletas/{id}", ObtenerBoletaPorIDMock)
	mux.HandleFunc("POST /api/boletas/{id}/anulacion", AnularBoletaMock)
}

func ListarBoletasMock(w http.ResponseWriter, r *http.Request) {
	estadoFiltro := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("estado"))) // Hito 2: agregar también filtro opcional por query ?folio=
	boletas := boletasMockBase()                                                      // Hito 2: llamar a service.ListarBoletas(ctx, filtros)

	if estadoFiltro != "" {
		filtradas := make([]Boleta, 0)
		for _, b := range boletas {
			if b.Estado == estadoFiltro {
				filtradas = append(filtradas, b)
			}
		}
		boletas = filtradas
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(boletas)
}

func EmitirBoletaMock(w http.ResponseWriter, r *http.Request) {
	if strings.TrimSpace(r.Header.Get("Idempotency-Key")) == "" { // Hito 2: verificar en BD si esta clave ya fue procesada antes
		responderProblema(w, http.StatusBadRequest, "Bad Request", "El header Idempotency-Key es obligatorio para emitir una boleta.", r.URL.Path)
		return
	}

	var req EmitirBoletaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responderProblema(w, http.StatusBadRequest, "Bad Request", "Payload JSON inválido.", r.URL.Path)
		return
	}

	if strings.TrimSpace(req.VentaID) == "" || len(req.Items) == 0 {
		responderProblema(w, http.StatusUnprocessableEntity, "Unprocessable Entity", "Se requiere venta_id y al menos un ítem válido.", r.URL.Path)
		return
	}

	var montoTotal int64
	itemsCalculados := make([]ItemBoleta, 0, len(req.Items))
	for _, it := range req.Items {
		if it.Cantidad <= 0 || it.PrecioUnitario <= 0 {
			responderProblema(w, http.StatusUnprocessableEntity, "Unprocessable Entity", "La cantidad y el precio_unitario en pesos deben ser enteros mayores a 0.", r.URL.Path)
			return
		}
		sub := int64(it.Cantidad) * it.PrecioUnitario // Hito 2: mover cálculo de subtotales e IVA a internal/service
		it.Subtotal = sub
		montoTotal += sub
		itemsCalculados = append(itemsCalculados, it)
	}

	montoNeto := (montoTotal * 100) / 119
	montoIVA := montoTotal - montoNeto

	nueva := Boleta{
		ID:           "770e8400-e29b-41d4-a716-446655440999", // Hito 2: generar UUID real y obtener folio correlativo atómico desde PostgreSQL
		Folio:        1003,
		VentaID:      req.VentaID,
		RutComprador: req.RutComprador,
		Estado:       "EMITIDA",
		MontoNeto:    montoNeto,
		MontoIVA:     montoIVA,
		MontoTotal:   montoTotal,
		FechaEmision: time.Now().Format(time.RFC3339),
		Items:        itemsCalculados,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(nueva)
}

func ObtenerBoletaPorIDMock(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	for _, b := range boletasMockBase() { // Hito 2: reemplazar búsqueda en slice por service.ObtenerBoleta(ctx, id)
		if b.ID == id {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(b)
			return
		}
	}
	responderProblema(w, http.StatusNotFound, "Not Found", "No existe una boleta con el ID especificado.", r.URL.Path)
}

func AnularBoletaMock(w http.ResponseWriter, r *http.Request) {
	if strings.TrimSpace(r.Header.Get("Idempotency-Key")) == "" { // Hito 2: propagar este mismo Idempotency-Key al servicio de Stock
		responderProblema(w, http.StatusBadRequest, "Bad Request", "El header Idempotency-Key es obligatorio para anular una boleta y devolver stock.", r.URL.Path)
		return
	}

	id := r.PathValue("id")
	var req AnularBoletaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(strings.TrimSpace(req.Motivo)) < 5 {
		responderProblema(w, http.StatusUnprocessableEntity, "Unprocessable Entity", "Debe indicar un motivo de anulación de al menos 5 caracteres.", r.URL.Path)
		return
	}

	for _, b := range boletasMockBase() { // Hito 2: reemplazar por transacción con SELECT FOR UPDATE + llamada HTTP (5s timeout) a STOCK_URL
		if b.ID == id {
			if b.Estado == "ANULADA" {
				responderProblema(w, http.StatusConflict, "Conflict", "La boleta ya se encuentra ANULADA; no se puede volver a devolver el stock.", r.URL.Path)
				return
			}
			ahora := time.Now().Format(time.RFC3339)
			b.Estado = "ANULADA"
			b.FechaAnulacion = &ahora
			b.MotivoAnulacion = &req.Motivo

			resp := AnularBoletaResponse{
				Boleta:                    b,
				StockDevuelto:             true,                                   // Hito 2: confirmar solo si el servicio de Stock responde 200/201
				ReferenciaDevolucionStock: "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d", // Hito 2: usar ID real retornado por la API de Stock
				FechaAnulacion:            ahora,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
	}

	responderProblema(w, http.StatusNotFound, "Not Found", "No existe una boleta con el ID especificado.", r.URL.Path)
}