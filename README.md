# OpenWA Go SDK

Cliente HTTP Go para el contrato OpenWA `0.23.4`. Reutilizable entre proyectos, sin dependencias externas y compatible con Go 1.22 o posterior.

```sh
go get github.com/unkcode-org/openwa-go-sdk@v0.1.0
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
| `GetSession` | `GET /api/sessions/{id}` | Lee estado, identidad y `engineLoaded`, sin iniciar la sesión. |
| `StartSession` | `POST /api/sessions/{id}/start` | Inicia una sesión existente; conserva los errores de conflicto del gateway. |
| `GetQRCode` | `GET /api/sessions/{id}/qr` | Obtiene el QR disponible; no espera un escaneo ni inicia sesiones. |
| `CheckNumber` | `GET /api/sessions/{id}/contacts/check/{number}` | Consulta existencia y devuelve el identificador canónico del proveedor. |
| `SendText` | `POST /api/sessions/{id}/messages/send-text` | Espera aceptación; soporta menciones, respuesta citada y opciones de preview. |
| `SendDocument` | `POST /api/sessions/{id}/messages/send-document` | Envía bytes base64 o una URL de documento, con nombre, MIME y caption. |

La administración de claves, creación/borrado de sesiones, webhooks, grupos y otras operaciones de OpenWA quedan fuera de esta primera versión. El SDK no ejecuta reconexión, logout, stop ni fallback automáticamente. Los estados nuevos no vacíos se conservan para compatibilidad futura. La aplicación decide cuándo iniciar una sesión según `EngineLoaded`; `disconnected` también puede significar que un motor está reconectando.

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

`Base64` contiene base64 estándar, sin prefijo data URL ni saltos de línea. Se requiere exactamente una fuente: `Base64` o `URL`. Las URLs las descarga el gateway, por lo que el límite local de bytes de documento solo aplica a base64. Restringir URLs admitidas por la aplicación para evitar SSRF en el gateway; preferir base64 para archivos privados. El SDK no normaliza teléfonos por país ni convierte IDs de WhatsMeow: usar el formato del proveedor o el ID de `CheckNumber`.

## Errores y reintentos

El contrato confirmado no documenta idempotencia en `send-text` ni `send-document`. El cliente no realiza reintentos automáticos de ninguna operación ni agrega encabezados de idempotencia.

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

Para cada cliente/ambiente, configurar el ID real de su sesión y una clave restringida mediante `allowedSessions`. El SDK no impone pertenencia entre tenants: corresponde al consumidor y al gateway. Mantener credenciales administrativas fuera de las aplicaciones y proteger el QR como material de autenticación.

## Contrato fijado

- Fuente indicada por el usuario: [GoFip Backend OpenAPI](https://github.com/unkcode-org/gofip-backend/blob/67b3ead89c23307fa58a26c885329a7c662874e4/backend/docs/openwa-openapi.json).
- Versión declarada: `0.23.4`.
- Copia exacta: `testdata/openapi.json`.
- SHA-256: `2497289cbaf479776511df91bf8278e8298d30f49a43aee4633be84e559c11d2`.

Los tests verifican el hash y los nombres de campos/operaciones soportados. No representan una prueba contra una instancia real ni garantizan compatibilidad con otras versiones.

## Desarrollo

```sh
gofmt -w .
go vet ./...
go test -race -coverprofile=coverage.out ./...
go build ./...
```

El CI ejecuta los checks en Go 1.22 y 1.26. Las pruebas usan `httptest` o transportes locales; no requieren secretos ni envían mensajes reales.
