# Modos de Falla — Servicio Tareas

Documento de resiliencia y modos de falla del microservicio de tareas según el estándar **DAS (Desarrollo de Microservicios)**.

---

## 1. Matriz de Modos de Falla

| ID | Modo de Falla | Causa Raíz Potencial | Mitigación / Qué la Impide | Código HTTP / Comportamiento | Cómo se Verifica |
|:---|:---|:---|:---|:---:|:---|
| **MF-01** | Base de datos PostgreSQL no disponible o caída | Problemas de red, contenedor apagado, credenciales erróneas | Conexión con timeout explícito (5s), reintento transparente por petición y endpoint `/readyz` | `503 Service Unavailable` / `500 Internal Server Error` (sin exponer trazas internas) | `GET /readyz` devuelve 503; pruebas en `db_test.go` |
| **MF-02** | Payload JSON inválido o malformado | Peticiones HTTP con sintaxis JSON corrupta | Validación temprana en el handler mediante `json.NewDecoder` | `400 Bad Request` | `POST /api/tareas` con body `{invalido}` en `tarea_controller_test.go` |
| **MF-03** | Campo obligatorio faltante (`titulo` vacío) | Cliente envía objeto sin el campo requerido | Validación explícita de campos antes de invocar la capa de persistencia | `400 Bad Request` | `POST /api/tareas` con `{"titulo": ""}` en `tarea_controller_test.go` |
| **MF-04** | Solicitud OPTIONS / CORS desde orígenes externos | Peticiones preflight desde navegadores de orígenes no autorizados | Handler centralizado de cabeceras CORS (`setCORS`) con origen específico (`ALLOWED_ORIGIN`) | `200 OK` para preflight autorizado | `OPTIONS /api/tareas` en `tarea_controller_test.go` y `main_test.go` |
| **MF-05** | Fallo en lectura o timeouts en conexiones | Consultas lentas o bloqueo de tablas | Uso obligatorio de `context.WithTimeout` (máximo 5 segundos) en toda llamada a la base de datos | `500 Internal Server Error` (timeout cancelado) | Cancelación de contexto probada en `context.WithTimeout` |
| **MF-06** | Inyección SQL | Entrada maliciosa del usuario en campos de texto | Consultas 100% parametrizadas con marcadores de posición nativos (`$1, $2, ...`) | Prevención total a nivel de driver | Verificación estática SonarQube y pruebas unitarias |

---

## 2. Estrategia de Verificación y Salud

1. **Salud del Proceso (`GET /healthz`):**
   * Verifica que el servidor HTTP esté vivo y aceptando solicitudes.
   * Retorna `200 OK` con contenido `ok`.

2. **Salud de la Dependencia (`GET /readyz`):**
   * Realiza un ping activo a PostgreSQL con timeout de 2 segundos.
   * Retorna `200 OK` (`ready`) si la base contesta, o `503 Service Unavailable` si PostgreSQL está temporalmente caído.

