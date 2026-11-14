# OpenWA Go SDK

Cliente HTTP Go para el contrato OpenWA `0.30.1`. Reutilizable entre proyectos, sin dependencias externas y compatible con Go 1.22 o posterior.

```sh
go get github.com/unkcode-org/openwa-go-sdk@v0.5.0
```

Si el repositorio es privado, el consumidor necesita acceso Git y una configuración
`GOPRIVATE` que incluya `github.com/unkcode-org/*` para evitar consultas al proxy
y al servicio público de checksums.

```go
import openwa "github.com/unkcode-org/openwa-go-sdk"
```

## Uso

```go
client, err := openwa.NewClient(openwa.Config{
    BaseURL: os.Getenv("OPENWA_BASE_URL"),
    APIKey:  os.Getenv("OPENWA_API_KEY"),
    Timeout: 30 * time.Second,
})
if err != nil {
    return err
}

ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
defer cancel()

result, err := client.SendText(ctx, os.Getenv("OPENWA_SESSION_ID"), openwa.TextMessage{
    ChatID: "5493510000000@c.us",
    Text:   "Tu reserva está confirmada",
})
if err != nil {
    // Ver la clasificación de errores antes de decidir un reintento.
    return err
}
// result.MessageID identifica la aceptación en el gateway, no la entrega.
_ = result
```

`BaseURL` es la raíz del gateway, por ejemplo `https://whatsapp.example.com` o una base con prefijo como `https://example.com/openwa`. El SDK agrega `/api/...`. Se permite HTTP en redes privadas; usar HTTPS al cruzar redes no confiables.

El cliente es seguro para uso concurrente con un transporte que también lo sea. No lee variables de entorno, conecta sesiones ni hace requests durante su construcción. Las variables del ejemplo las administra la aplicación.

## Operaciones incluidas

| Método | Endpoint | Comportamiento |
| --- | --- | --- |
| `CreateSession` | `POST /api/sessions` | Crea una sesión con nombre único; el ID lo devuelve el gateway. |
| `ListSessions` | `GET /api/sessions` | Lee una página para recuperar por nombre una creación previa. |
| `GetSession` | `GET /api/sessions/{id}` | Lee estado, identidad y `engineLoaded`, sin iniciar la sesión. |
| `StartSession` | `POST /api/sessions/{id}/start` | Inicia una sesión existente; conserva los errores de conflicto del gateway. |
| `GetQRCode` | `GET /api/sessions/{id}/qr` | Obtiene el QR disponible; no espera un escaneo ni inicia sesiones. |
| `CheckNumber` | `GET /api/sessions/{id}/contacts/check/{number}` | Consulta existencia y devuelve el identificador canónico del proveedor. |
| `SendText` | `POST /api/sessions/{id}/messages/send-text` | Espera aceptación; soporta menciones, respuesta citada y opciones de preview. |
| `SendDocument` | `POST /api/sessions/{id}/messages/send-document` | Envía bytes base64 o una URL de documento, con nombre, MIME y caption. |
| `SendImage` | `POST /api/sessions/{id}/messages/send-image` | Envía una imagen base64 o por URL, con MIME y caption. |
| `ListWebhooks` | `GET /api/sessions/{id}/webhooks` | Lista las suscripciones de la sesión. |
| `CreateWebhook` | `POST /api/sessions/{id}/webhooks` | Registra una suscripción con eventos y secreto de firma opcional. |
| `UpdateWebhook` | `PUT /api/sessions/{id}/webhooks/{webhookId}` | Actualiza los campos suministrados, incluido el estado activo. |

La administración de claves y grupos queda fuera de esta versión. Stop, logout y delete son operaciones explícitas y nunca se ejecutan como fallback. El cliente HTTP no reconecta automáticamente; el supervisor opcional descrito abajo sí puede recuperar motores descargados cuando la aplicación lo habilita. Los estados nuevos no vacíos se conservan para compatibilidad futura. La aplicación decide cuándo iniciar una sesión según `EngineLoaded`; `disconnected` también puede significar que un motor está reconectando.

Para enviar un PDF privado:

```go
result, err := client.SendDocument(ctx, sessionID, openwa.DocumentMessage{
    ChatID:   chatID,
    Base64:   base64.StdEncoding.EncodeToString(pdfBytes),
    MIMEType: "application/pdf",
    Filename: "factura.pdf",
    Caption:  "Tu factura",
})
```

`Base64` contiene base64 estándar, sin prefijo data URL ni saltos de línea. Se requiere exactamente una fuente: `Base64` o `URL`. Las URLs las descarga el gateway, por lo que el límite local de bytes de documento solo aplica a base64. Restringir URLs admitidas por la aplicación para evitar SSRF en el gateway; preferir base64 para archivos privados. El SDK no normaliza teléfonos por país: usar el ID de `CheckNumber`. OpenWA 0.30.1 devuelve ese ID en formato `<phone>@c.us` en ambos motores.

### Imágenes y webhooks

```go
result, err := client.SendImage(ctx, sessionID, openwa.ImageMessage{
    ChatID:   chatID,
    Base64:   base64.StdEncoding.EncodeToString(imageBytes),
    MIMEType: "image/png",
    Caption:  "Tu pedido",
})
```

`MaxRequestBytes` limita el JSON de la imagen, incluida la expansión de base64. El gateway aplica sus límites de medios. Las URL son descargadas por el gateway.

```go
hook, err := client.CreateWebhook(ctx, sessionID, openwa.CreateWebhookInput{
    URL:    callbackURL,
    Events: []string{"message.ack"},
    Secret: signingSecret,
})
```

Usá `ListWebhooks` para recuperar la suscripción por URL antes de repetir una creación cuyo resultado quedó incierto. Usá `UpdateWebhook` para actualizar una suscripción existente por ID. El secreto y los encabezados configurados son de escritura. El receptor del webhook verifica la firma y procesa los eventos.

## Errores y reintentos

OpenWA admite idempotencia en rutas de envío; este SDK todavía no expone esa opción. El cliente no realiza reintentos automáticos de ninguna operación ni agrega encabezados de idempotencia.

```go
if errors.Is(err, openwa.ErrOutcomeUnknown) {
    // El mensaje pudo haber sido aceptado: persistir y reconciliar el intento.
    // No repetirlo automáticamente.
}

var apiErr *openwa.APIError
if errors.As(err, &apiErr) {
    // StatusCode, Code y RetryAfter permiten clasificar la respuesta.
    // RetryAfter no acredita por sí solo que sea seguro reenviar.
}

if errors.Is(err, context.DeadlineExceeded) {
    // Se conserva la causa aunque el error también sea ErrOutcomeUnknown.
}
```

| Caso de envío | Resultado |
| --- | --- |
| Configuración/input inválido, límite local o contexto ya cancelado | Fallo local anterior al dispatch; no se clasifica como resultado incierto. |
| Respuesta HTTP 4xx, excepto 408 | `APIError`: el gateway rechazó el request. No significa que repetir vaya a resolverlo. |
| HTTP 408/5xx u otro estado inesperado | `SendError`, identificable con `errors.Is(err, ErrOutcomeUnknown)`; conserva el `APIError`. |
| Fallo de transporte, cancelación/timeout tras dispatch, respuesta truncada, demasiado grande o inválida | `SendError` con resultado incierto. |
| HTTP 201 con `messageId` y `timestamp` válidos | Aceptación confirmada por el gateway; no confirma entrega ni lectura. |

La clasificación es conservadora: un error de transporte después de invocar el cliente HTTP se considera incierto incluso si la conexión pudo haber fallado antes de enviar bytes. La persistencia de intentos, deduplicación por evento y reconciliación pertenecen a la aplicación. Una respuesta `201` puede existir incluso para un destinatario no registrado; usar `CheckNumber` cuando sea necesario verificarlo.

Los errores imprimibles omiten cuerpos remotos, claves, URLs y mensajes del proveedor. `APIError.Code`, los errores subyacentes accesibles con `errors.Unwrap`, los datos de sesión y el QR siguen siendo datos sensibles/no confiables: no registrarlos sin tratamiento. Un error de lookup nunca se transforma en `Exists=false`.

## Límites y transporte

| Configuración | Default |
| --- | --- |
| `Timeout` | 30 segundos; si se inyecta un cliente con timeout positivo se conserva, salvo override explícito. |
| `MaxResponseBytes` | 1 MiB por respuesta. |
| `MaxRequestBytes` | 16 MiB de JSON serializado. |
| `MaxDocumentBytes` | 10 MiB decodificados, para base64. |

Cero selecciona el default; valores negativos son inválidos. Configurar límites acordes al proxy/gateway, considerando la expansión base64. Texto: hasta 4096 caracteres; caption: hasta 1024. `HTTPClient` se copia sin mutar la instancia inyectada. Todas las redirecciones se rechazan, incluso hacia el mismo host. Los transportes personalizados deben respetar contextos y no agregar reintentos de envíos ni logs de credenciales.

Para cada cliente/ambiente, persistir la identidad de su sesión. Las claves restringidas mediante `allowedSessions` sirven para operar sesiones ya asignadas. Crear y recuperar sesiones requiere una credencial con permisos de creación/listado; el SDK no amplía permisos ni modifica claves automáticamente. El SDK no impone pertenencia entre tenants: corresponde al consumidor y al gateway. Mantener credenciales administrativas fuera de las aplicaciones y proteger el QR como material de autenticación.

## Contrato fijado

- Fuente: [OpenWA v0.30.1 OpenAPI](https://github.com/rmyndharis/OpenWA/blob/v0.30.1/openapi.json).
- Versión declarada: `0.30.1`.
- Copia exacta: `testdata/openapi.json`.
- SHA-256: `1f0c3ba6f7be80a44f2f27ab881531a291563182ea04c240279014adcd015afe`.

Los tests verifican el hash y los nombres de campos/operaciones soportados. No representan una prueba contra una instancia real ni garantizan compatibilidad con otras versiones.

## Desarrollo

```sh
gofmt -w .
go vet ./...
go test -race -coverprofile=coverage.out ./...
go build ./...
```

El CI ejecuta los checks en Go 1.22 y 1.26. Las pruebas usan `httptest` o transportes locales; no requieren secretos ni envían mensajes reales.

## Creación recuperable de sesiones

```go
created, err := client.CreateSession(ctx, openwa.CreateSessionInput{Name: persistedUniqueName})
page, err := client.ListSessions(ctx, &openwa.ListSessionsOptions{Limit: 100, Offset: 0})
```

Persistir un nombre único por instalación antes del POST. Si se pierde la respuesta,
recuperar mediante listado paginado y coincidencia exacta del nombre; no elegir la
primera sesión ni generar otro nombre al reintentar. Un 409 indica que el nombre
ya existe. La creación no vincula automáticamente la cuenta: iniciar la sesión
y presentar su QR sigue siendo un paso independiente. El SDK no persiste estado
ni reintenta automáticamente la creación.

## Ciclo de vida explícito (v0.3.0)

`StopSession`, `LogoutSession`, `DeleteSession` y `RequestPairingCode` completan
las operaciones de administración usadas por Gofip. Son llamadas únicas: no
reintentan ni convierten un fallo de teardown en éxito. Logout elimina credenciales
y exige volver a vincular. El código de vinculación es material de autenticación.
`CreateSessionInput` admite `ProxyURL` y `ProxyType` (http/https/socks4/socks5).

`GetSessionConfig` y `UpdateSessionConfig` consultan y modifican configuración
sin reiniciar. En `SessionConfigPatch`, un campo nil se omite; un
`&openwa.ConfigValue[int]{}` envía null y restaura el default del gateway:

```go
_, err = client.UpdateSessionConfig(ctx, sessionID, openwa.SessionConfigPatch{
    MaxReconnectAttempts: &openwa.ConfigValue[int]{}, // ilimitado
})
```

Para enviar cero/false explícitamente, usar `Value: &value` dentro de ConfigValue.
**Cero deshabilita la reconexión**; nil/null significa ilimitado. Los cambios de
reconexión se aplican en el próximo start. No cambian un backoff ya en curso.

## Supervisión opcional de sesiones (v0.3.0)

```go
supervisor, err := openwa.NewSupervisor(client, openwa.SupervisorConfig{
    PollInterval: 30 * time.Second,
    OnChange: func(result openwa.SupervisionResult) {
        // Persistir/publicar estado mediante la aplicación. No incluye QR ni teléfono.
    },
})
if err != nil { return err }
// Esta sesión ya fue validada como propia y deseada activa por la aplicación.
// Run bloquea; el caller puede ejecutarlo en su worker y debe esperar su retorno.
err = supervisor.Run(ctx, sessionID)
```

No se inicia ninguna goroutine al construir el cliente o el supervisor. `Run`
consulta inmediatamente y periódicamente, usa espera creciente con jitter ante
recuperación/fallos y devuelve `ctx.Err()` cuando termina. Defaults: polling 30 s,
timeout por reconciliación 15 s, backoff máximo 5 min y concurrencia máxima 4.
`OnChange` recibe transiciones y debe retornar rápidamente; corre sin locks de
ciclo de vida. Un segundo Run para la misma sesión en el mismo supervisor devuelve
`ErrAlreadySupervised`.

`Reconcile(ctx, sessionID)` realiza una sola observación/recuperación, útil para
workers con scheduling durable propio. Serializa llamadas por sesión dentro de
esa instancia; la aplicación conserva backoff, locks distribuidos y scope.

El supervisor sólo llama GET y, si corresponde, START. No interrumpe un motor
cargado aunque el estado sea disconnected. Espera durante autenticación y ante
QR, restricciones, acción humana o estados desconocidos; una sesión 404 se informa
como missing y nunca se recrea. Tras 400/409 de start vuelve a consultar una vez.
Los motores descargados en estados conocidos recuperables pueden iniciarse.
Un motor estancado pero aún cargado se informa como recovering: no se fuerza un
stop, logout, delete ni force-kill.

La conexión WebSocket y sus heartbeats pertenecen al motor del gateway. Leer
estado no es un ping extremo a extremo ni garantiza la recepción de un mensaje.

La aplicación debe:

- Registrar sólo sesiones propias y deseadas activas; persistir esa intención y
  reconstruir los workers tras reinicios.
- Cancelar Run y esperar su retorno antes de detener, desvincular o eliminar una
  sesión manualmente. Reconcile debe coordinarse con esas operaciones del mismo modo.
- Coordinar exclusión entre procesos/réplicas; los locks del SDK sólo son locales.
- Mantener la cola de mensajes y reactivarla al recuperar disponibilidad. Nunca
  consumir todos los intentos por una caída ni repetir `ErrOutcomeUnknown` a ciegas.

Ni Run ni Reconcile llaman a métodos de envío. El cambio no altera la política
sin reintentos automáticos de mensajes.

## Compatibilidad con OpenWA 0.30.1 (v0.4.0)

El SDK cubre únicamente las operaciones listadas arriba y las de ciclo de vida y configuración; no implementa toda la API. `CustomLinkPreview` admite `ThumbnailURL` y `ThumbnailBase64` (si se envían ambos, el gateway prioriza base64). La URL de imagen la descarga el gateway y debe tratarse como dato no confiable. `RestrictionPasskeyRequired` identifica una solicitud de passkey; se conservan valores desconocidos. Los códigos de vinculación requieren entre 6 y 15 dígitos.

Revisar antes de actualizar el servidor desde 0.23.4: crear sesiones con proxy requiere clave ADMIN, comprobar números requiere al menos OPERATOR y el rechazo por alcance responde 403 en vez de 401. `ProxyType` está obsoleto y lo determina el esquema de `ProxyURL`. Hay validaciones más estrictas, migraciones y condiciones de rollback entre versiones; consultar las [notas de actualización](https://github.com/rmyndharis/OpenWA/blob/v0.30.1/docs/14-migration-guide.md) y probar en staging con copias de seguridad. Este SDK no garantiza compatibilidad automática de todas las integraciones ni realiza el despliegue.
