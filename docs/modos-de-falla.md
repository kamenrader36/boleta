# Modos de Falla — Servicio Boleta y Anulación (Servicio 4)

Documento de resiliencia, concurrencia y modos de falla del microservicio de **Boleta y anulación** (registro de folio, anulación y devolución de stock) según los lineamientos técnicos del curso **DAS — Desarrollo de Microservicios**.

---

## 1. Matriz de Modos de Falla

| ID | Modo de Falla | Causa Raíz Potencial | Mitigación / Qué la Impide | Código HTTP / Comportamiento | Cómo se Verifica |
|:---|:---|:---|:---|:---:|:---|
| **MF-01** | **Colisión de folios por emisión concurrente** | Dos o más peticiones simultáneas intentan emitir una boleta al mismo tiempo y leen el mismo último folio disponible. | Secuencia atómica en PostgreSQL / restricción `UNIQUE (folio)` e inserción transaccional junto con soporte de `Idempotency-Key`. | `201 Created` con folios correlativos únicos, o `409 Conflict` (`application/problem+json`) si hay clave duplicada en vuelo. | Prueba E2E de concurrencia lanzando emisiones simultáneas y validando unicidad de folios asignados. |
| **MF-02** | **Doble anulación concurrente (Doble devolución de stock)** | Dos solicitudes simultáneas intentan anular la misma boleta `EMITIDA`, arriesgando devolver el stock dos veces al servicio de Inventario/Stock. | Bloqueo pesimista a nivel de fila (`SELECT ... FOR UPDATE`) verificando estado `EMITIDA` + envío obligatorio de `Idempotency-Key` hacia el servicio de Stock. | La primera petición retorna `200 OK` (`ANULADA`); la segunda recibe `409 Conflict` indicando que la boleta ya fue anulada. | Prueba E2E con dos peticiones concurrentes a `POST /api/boletas/{id}/anulacion` sobre el mismo UUID. |
| **MF-03** | **Fallo o Timeout del servicio externo de Stock al anular** | El microservicio de Stock está caído, responde `5xx` o tarda más de 5 segundos al solicitar la devolución de inventario durante una anulación. | Cliente HTTP con timeout explícito de `5s` (`context.WithTimeout`). La anulación local solo se confirma si el servicio de Stock confirma la devolución (rollback transaccional si falla). | `502 Bad Gateway` o `504 Gateway Timeout` en formato RFC 9457 (`application/problem+json`). La boleta mantiene su estado `EMITIDA`. | Simulación de timeout/error del servicio de Stock verificando que el estado de la boleta no cambie a `ANULADA`. |
| **MF-04** | **Reintento de red en emisión o anulación (Falta de Idempotencia)** | El cliente reintenta `POST /api/boletas` o `POST /api/boletas/{id}/anulacion` por corte de red tras haberse procesado la operación. | Exigencia y registro del header `Idempotency-Key` por operación crítica, devolviendo el resultado original almacenado sin re-ejecutar efectos secundarios. | `200 OK` / `201 Created` devolviendo exactamente el mismo payload original o `400 Bad Request` si falta la cabecera. | Envío repetido del mismo `Idempotency-Key` comprobando que no se genera un segundo folio ni otra llamada a Stock. |
| **MF-05** | **Datos de entrada o montos inválidos** | Envío de montos negativos, decimales (`float`), lista de ítems vacía o identificadores que no cumplen formato UUID. | Validación estricta en el contrato OpenAPI 3.1 (`minimum`, `minItems`, `format: uuid`) y uso exclusivo de enteros en pesos (`int64` / `bigint`). | `422 Unprocessable Entity` / `400 Bad Request` en formato RFC 9457 antes de tocar la capa de servicio. | Peticiones con montos negativos o sin ítems rechazadas por validación de esquema. |
| **MF-06** | **Base de datos PostgreSQL no disponible** | Contenedor apagado, saturación del pool de conexiones o caída de red interna. | Pool con `context.WithTimeout` (5s en consultas, 2s en readiness) y monitoreo mediante `/readyz`. | `503 Service Unavailable` en `/readyz` y `500 Internal Server Error` seguro (sin exponer SQL ni trazas). | `GET /readyz` con base desconectada retorna `503`. |

---

## 2. Estrategia de Salud y Observabilidad

1. **Liveness Probe (`GET /healthz`):**
   * Verifica que el proceso del microservicio esté activo. No requiere autenticación.
   * Responde `200 OK`.

2. **Readiness Probe (`GET /readyz`):**
   * Verifica conectividad real contra PostgreSQL con timeout estricto de 2 segundos. No requiere autenticación.
   * Responde `200 OK` únicamente si PostgreSQL contesta, o `503 Service Unavailable` en caso contrario.

3. **Propagación de Trazabilidad:**
   * Toda llamada saliente hacia el servicio de Stock reenvía el JWT del usuario, el header `X-Request-Id` para correlación de logs estructurados (`log/slog`) y el `Idempotency-Key` correspondiente.