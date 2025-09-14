# Integración segura de Webhooks de RevenueCat

Este documento explica paso a paso cómo construir un webhook seguro y confiable para RevenueCat, basado en las prácticas aplicadas en este repositorio (Go + Fiber, arquitectura limpia) y las recomendaciones generales de RevenueCat.

Contenido

- Visión general
- Requisitos y entorno (Docker, .env, config.yml)
- Diseño y contrato (handler, servicio, repos)
- Verificación de firma HMAC (rotación de secretos)
- Idempotencia y persistencia (índice único + manejo de duplicados)
- Manejo de errores y códigos HTTP
- Seguridad operativa (TLS, límites, logging, secretos)
- Testing y pruebas locales (curl / openssl, tests unitarios)
- Configurar el webhook en el dashboard de RevenueCat
- Checklist de despliegue

---

## Visión general

RevenueCat envía eventos de suscripción como webhooks (payload JSON) a un endpoint que tú configures. Para procesarlos de forma segura necesitas:

- Verificar la firma HMAC-SHA256 que RevenueCat incluye en el header `X-RevenueCat-Signature`.
- Manejar idempotencia para evitar procesar el mismo evento dos veces (p. ej. índice único por `event_id`).
- Clasificar errores: responder 401 para firmas inválidas, 5xx para problemas transitorios (permitir retry), 200 para eventos ya procesados.

## Requisitos y entorno

- Repositorio: Go 1.21+, Fiber, MongoDB (repos del proyecto muestran un ejemplo), Redis opcional para encolado.
- Variables de entorno principales (ver `.env.example`, `.env.dev`, `.env.prod`):
  - `APP_REVENUECAT_WEBHOOK_SECRET` — secret HMAC para verificar firmas (coma-separado para rotación)
  - `APP_REVENUECAT_API_KEY` — sólo si integras con la API REST de RevenueCat
  - `APP_SERVER_PORT` y datos de BD (`APP_DATABASE_MONGO_URI`, `APP_DATABASE_REDIS_ADDR`)
- `config.yml` en repo contiene defaults no sensibles (p. ej. `revenuecat.base_url`).
- Docker compose: `docker-compose.dev.yml` carga `.env.dev` y monta el código para desarrollo; `docker-compose.prod.yml` usa `.env.prod` y no expone puertos internos de DB.

## Diseño y contrato

Siguiendo Clean Architecture, divide responsabilidades:

- Adapter/HTTP (handler): valida firma, parsea JSON y delega a Service.
- Application/Service: lógica de negocio, idempotencia, sincronización/creación de usuarios y persistencia.
- Ports/Repository: abstracción del almacenamiento con métodos como `GetByEventID`, `Create`, `UpdateByEventID`.

Contrato ligero del handler:

- Entrada: HTTP POST JSON, header `X-RevenueCat-Signature` con formato `sha256=HEX`.
- Salidas:
  - 401: firma ausente/incorrecta
  - 400: payload inválido o evento no soportado (errores permanentes)
  - 500: error transitorio (DB/infra) — deja que RevenueCat vuelva a intentar
  - 200: éxito o evento ya procesado (idempotencia)

## Verificación de firma HMAC

Recomendaciones implementadas en este repo:

- Implementa una función pura `VerifyRevenueCatSignature(body []byte, signatureHeader string, secrets []string) error` que:
  1. Extrae la parte hex del header (`sha256=HEX`).
  2. Para cada secret configurado, calcula HMAC-SHA256 sobre el `body` y compara el hex con `subtle.ConstantTimeCompare`.

3.  Retorna error tipado (`ErrMissingSignature` o `ErrInvalidSignature`).

Por qué pura:

- Facilita pruebas unitarias sin depender del framework (Fiber) ni del contexto HTTP.

Rotación de secretos

- Acepta múltiples secretos (lista comma-separated en `APP_REVENUECAT_WEBHOOK_SECRET` o `revenuecat.webhook_secrets` en `config.yml`).
- Flujo recomendado para rotar:
  1. Añade el nuevo secret junto al antiguo en la configuración (ambos activos).
  2. Despliega y valida que verificaciones con el nuevo secret funcionan.
  3. Tras un periodo, elimina el secret antiguo.

## Idempotencia y persistencia

- Añade un índice único en la colección de suscripciones sobre `event_id` (el repo de ejemplo crea `unique_event_id`).
- En el repositorio, cuando `InsertOne` devuelve duplicate key, retorna un error sentinel (p. ej. `ports.ErrDuplicateEvent`).
- En el Service, trata ese error como _ya procesado_ y devuelve éxito (no reintentar ni fallar).

Esto elimina la ventana de race entre `GetByEventID` y `Create`.

## Manejo de errores y códigos HTTP

- Firma ausente/incorrecta -> 401 Unauthorized.
- Payload inválido (JSON malformado) -> 400 Bad Request.
- Evento procesado (idempotencia) -> 200 OK.
- Error transitorio (DB caído) -> 500 Internal Server Error (dejar que RevenueCat reintente).

## Seguridad operativa

- Siempre usar HTTPS (terminación TLS en la capa de ingress/API gateway) y HSTS.
- No exponer puertos de DB en producción (ver docker-compose.prod.yml en repo).
- Limitar tamaño del body para evitar DoS (p. ej. límite 64KB configurable en Fiber).
- Rate limiting / WAF: aplicar reglas para bloquear patrones maliciosos.
- Redactar PII en logs; registrar sólo `event_id`, `app_user_id` y metadata no sensible.
- Mantener secrets fuera del repo y usar un secrets manager en producción (Vault, AWS Secrets Manager, Fly secrets, etc.).

## Testing y pruebas locales

1. Test unitarios:

- Testea la función pura `VerifyRevenueCatSignature` con varios secrets (incluyendo rotación) y payloads.
- Testea el handler integrándolo con un mock del Service para verificar códigos HTTP.

2. Prueba manual local (ejemplo con openssl):

- Firma un body JSON y envía con curl:

```bash
BODY='{"id":"test","type":"INITIAL_PURCHASE","app_user_id":"user1"}'
SECRET='test_secret'
SIG=$(printf "%s" "$BODY" | openssl dgst -sha256 -hmac "$SECRET" -binary | xxd -p -c 256)
curl -v -X POST http://localhost:3000/webhooks/revenuecat \
  -H "Content-Type: application/json" \
  -H "X-RevenueCat-Signature: sha256=$SIG" \
  -d "$BODY"
```

- Alternativa usando Python:

```bash
python -c "import hmac,hashlib,sys; body=sys.stdin.read().encode(); print('sha256='+hmac.new(b'test_secret', body, hashlib.sha256).hexdigest())" <<< "$BODY"
```

Configurar tests automatizados:

- El repo ya incluye `internal/adapters/http/webhook/revenuecat_validator_test.go` que demuestra cómo probar la verificación usando `app.Test` de Fiber.

## Configurar webhook en RevenueCat (dashboard)

1. Accede a tu proyecto en RevenueCat.
2. Ve a Settings -> Webhooks (o Integrations -> Webhooks según UI).
3. Añade un nuevo endpoint con la URL pública `https://<tu-dominio>/webhooks/revenuecat`.
4. Selecciona los events que te interesan (ej. INITIAL_PURCHASE, RENEWAL, CANCELLATION, REFUND, TRANSFER, EXPIRATION, etc.).
5. En la sección de seguridad / signing, copia el _Webhook Secret_ y guárdalo en tu entorno (`APP_REVENUECAT_WEBHOOK_SECRET`).
6. Test: RevenueCat ofrece una opción para enviar un test webhook al endpoint; valida que tu endpoint responde 200 y loguea el evento.

Notas sobre retries y backoff

- RevenueCat reintentará webhooks cuando reciba respuestas 5xx o no reciba respuesta. Diseña tus handlers para devolver 5xx en errores transitorios.
- Evita devolver 200 cuando el sistema no ha procesado el evento correctamente (salvo en idempotencia explícita).

## Checklist de despliegue (pre-flight)

- [ ] Secrets configurados en entorno/secret manager (no en repo).
- [ ] `revenuecat.webhook_secrets` o `APP_REVENUECAT_WEBHOOK_SECRET` contiene al menos un secret.
- [ ] Índice único en DB creado para `event_id`.
- [ ] Limite de tamaño de request aplicado (Fiber middleware).
- [ ] Logs redactan PII.
- [ ] Métricas expuestas: webhook count, signature failures, processing errors.
- [ ] Pruebas unitarias y de integración ejecutadas (`go test ./...`).

Referencias rápidas

- Header de firma: `X-RevenueCat-Signature` (formato: `sha256=HEX`).
- Usar HMAC-SHA256 sobre el raw request body.
- Documentación oficial de RevenueCat: https://www.revenuecat.com/docs/integrations/integrations

---

## Entornos en RevenueCat (sandbox/dev y prod) y pruebas con ngrok

Cuando desarrollas y pruebas webhooks, es recomendable separar los endpoints y secretos por ambiente (sandbox/dev vs prod). Esto evita mezclar eventos de prueba con datos de producción y facilita la rotación de secretos.

1. Separar endpoints por ambiente

- Crea dos endpoints en el dashboard de RevenueCat:

  - Dev/Sandbox: `https://<tu-dev-domain>/webhooks/revenuecat` o el URL temporal que te ofrezca ngrok cuando pruebas localmente.
  - Production: `https://<tu-prod-domain>/webhooks/revenuecat`

- Asigna un Webhook Secret distinto para cada endpoint. Guarda cada secret en su entorno correspondiente:
  - Dev: `APP_REVENUECAT_WEBHOOK_SECRET=dev_secret_1,dev_secret_2` (si soportas rotación)
  - Prod: `APP_REVENUECAT_WEBHOOK_SECRET=prod_secret_1,prod_secret_2`

2. Configurar los events por ambiente

- En dev/sandbox sólo habilita los eventos que quieras probar (p. ej. INITIAL_PURCHASE, RENEWAL). En prod habilita el set completo que necesites.
- Usa el botón "Send Test Webhook" en el dashboard para verificar rápidamente la integración contra cada endpoint.

3. Pruebas locales con ngrok (exponer tu servidor local vía HTTPS)

Requisitos: ngrok instalado y una app corriendo localmente en el puerto que usa tu servidor (por defecto 3000 en este repo).

- Instala ngrok (macOS):

```bash
brew install ngrok/ngrok/ngrok
ngrok authtoken <tu-authtoken>
```

- Levanta tu app localmente y expón el puerto con ngrok:

```bash
# si tu app corre en :3000
ngrok http 3000
```

- Copia la URL pública HTTPS que te devuelve ngrok (p. ej. `https://abcd-1234.ngrok.io`) y pégala como el endpoint en RevenueCat para el ambiente de desarrollo.

4. Firmar y enviar payloads al endpoint de ngrok

- Usa el mismo método de firma HMAC-SHA256 que en producción. Ejemplo con openssl:

```bash
BODY='{"id":"test","type":"INITIAL_PURCHASE","app_user_id":"user1"}'
SECRET='dev_secret_1'
SIG=$(printf "%s" "$BODY" | openssl dgst -sha256 -hmac "$SECRET" -binary | xxd -p -c 256)
curl -v -X POST https://abcd-1234.ngrok.io/webhooks/revenuecat \
  -H "Content-Type: application/json" \
  -H "X-RevenueCat-Signature: sha256=$SIG" \
  -d "$BODY"
```

- Alternativamente, usa la opción de "Test webhook" en el dashboard de RevenueCat apuntando al URL de ngrok.

5. Buenas prácticas al usar ngrok

- Usa el authtoken de ngrok y, si necesitas estabilidad durante pruebas largas, reserva un subdominio con una cuenta paga (`ngrok http -subdomain=myapp 3000`).
- Nunca uses la URL de ngrok en producción; es sólo para pruebas locales.
- Valida que el header `X-RevenueCat-Signature` esté presente y que la firma coincida antes de procesar eventos.

6. Ejemplo de flujo de pruebas completo

- 1. Ejecuta tu servicio local: `go run cmd/server/main.go`
- 2. Ejecuta `ngrok http 3000` y copia la URL HTTPS.
- 3. En RevenueCat, crea un endpoint para dev con la URL de ngrok y copia el Webhook Secret al `.env.dev` como `APP_REVENUECAT_WEBHOOK_SECRET`.
- 4. Envía un "Test webhook" desde RevenueCat o usa el `curl` firmado arriba.
- 5. Revisa los logs locales para verificar la validación y el procesamiento.

---

## Cumplimiento Completo con Documentación RevenueCat

Esta implementación ahora mapea completamente todos los campos principales de la documentación de RevenueCat (https://docs.revenuecat.com/reference/webhooks):

### Campos Mapeados

- ✅ **id**: ID único del evento
- ✅ **type**: Tipo de evento (INITIAL_PURCHASE, RENEWAL, CANCELLATION, etc.)
- ✅ **app_user_id**: ID del usuario en RevenueCat
- ✅ **product_id**: ID del producto comprado
- ✅ **store**: Plataforma de compra ("app_store", "play_store", "amazon")
- ✅ **environment**: "PRODUCTION" o "SANDBOX"
- ✅ **event_timestamp_ms**: Timestamp del evento
- ✅ **purchased_at_ms**: Timestamp de compra (opcional)
- ✅ **expires_at_ms**: Timestamp de expiración (opcional)
- ✅ **price**: Precio del producto (opcional)
- ✅ **currency**: Moneda del precio
- ✅ **transaction_id**: ID de transacción
- ✅ **original_transaction_id**: ID de transacción original
- ✅ **period_type**: Tipo de período ("normal", "intro", "trial")
- ✅ **entitlement_id** / **entitlement_ids**: IDs de entitlements
- ✅ **presented_offering_id**: ID de la offering presentada
- ✅ **subscriber_attributes**: Atributos del subscriber
- ✅ **cancel_reason**: Razón de cancelación
- ✅ **refund_reason**: Razón de reembolso
- ✅ **billing_issue_detected_at_ms**: Timestamp de problema de facturación

### Tipos de Eventos Soportados

- ✅ **INITIAL_PURCHASE**: Compra inicial
- ✅ **RENEWAL**: Renovación automática
- ✅ **CANCEL**: Cancelación
- ✅ **UNCANCEL**: Reactivación
- ✅ **REFUND**: Reembolso
- ✅ **BILLING_ISSUE**: Problema de facturación
- ✅ **EXPIRATION**: Expiración
- ✅ **TRANSFER**: Transferencia de suscripción
- ✅ **PRODUCT_CHANGE**: Cambio de producto

### Procesamiento de Subscribers

Los webhooks se procesan automáticamente según la documentación:

- Eventos se envían cuando hay cambios en subscribers
- Respuestas HTTP correctas (200, 401, 400, 500) para controlar reintentos
- Idempotencia garantizada con índices únicos
- Logging estructurado sin exposición de PII

## Mejoras Arquitectónicas Implementadas

Esta versión del webhook incluye mejoras significativas siguiendo principios SOLID, Clean Architecture y mejores prácticas de seguridad:

### 1. **Separación de Responsabilidades (SOLID)**

- **EventProcessor Interface**: Nueva interface en `internal/ports/event_processor.go` que desacopla el handler del service concreto.
- **DTO en Ports**: El `RevenueCatEvent` se movió a ports para evitar dependencias circulares y mantener Clean Architecture.
- **Generación Segura de Contraseñas**: Los usuarios sincronizados ahora reciben contraseñas criptográficamente seguras generadas con `crypto/rand`, en lugar de valores por defecto inseguros.

### 2. **Mejoras de Seguridad**

- **Rate Limiting**: Middleware in-memory que limita a 10 requests por minuto por IP para prevenir ataques DoS.
- **Límite de Body**: Middleware que rechaza payloads mayores a 64KB para evitar ataques de denegación de servicio.
- **Contraseñas Seguras**: Generación aleatoria de 32 caracteres hex para nuevos usuarios.

### 3. **Arquitectura Mejorada**

- **Clean Architecture Reforzada**: El DTO se movió a la capa de ports, eliminando dependencias del service hacia adapters.
- **Dependency Inversion**: El handler ahora depende de la interface `EventProcessor`, facilitando testing y extensibilidad.
- **Middlewares Reutilizables**: Los middlewares de seguridad se pueden aplicar a otras rutas webhooks.

### 4. **Logging y Monitoreo**

- **Structured Logging**: Uso consistente de zerolog con campos contextuales (`event_id`, `app_user_id`, etc.).
- **Redacción de PII**: Los logs evitan información sensible, registrando solo IDs y metadatos necesarios.

### 5. **Testing y Mantenibilidad**

- **Tests Actualizados**: Los tests unitarios se actualizaron para usar el nuevo DTO en ports.
- **Compilación Segura**: Todas las dependencias circulares se resolvieron, asegurando compilación limpia.

### Checklist de Mejoras Implementadas

- [x] Refactorizar RevenueCatEvent a ports layer
- [x] Crear EventProcessor interface para dependency inversion
- [x] Implementar generación segura de contraseñas
- [x] Agregar middleware de rate limiting (10 req/min por IP)
- [x] Agregar middleware de límite de body (64KB)
- [x] Actualizar tests para nuevo DTO
- [x] Verificar soporte completo de eventos RevenueCat
- [x] Mejorar logging con structured fields
- [x] Documentar cambios y arquitectura

Estas mejoras hacen el webhook más seguro, mantenible y escalable, cumpliendo con estándares enterprise de desarrollo Go.
Con esto tendrás un flujo seguro y separado por ambientes que te permitirá probar sin poner en riesgo la producción.
