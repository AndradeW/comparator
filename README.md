# comparator

Servicio HTTP en Go que compara dos respuestas HTTP (status, headers y cuerpos JSON) y devuelve las diferencias.

## Requisitos

- Go 1.26+

## Uso

Levantar el servidor:

```bash
go run ./cmd
```

Escucha en `:8080` por defecto.

Test, vet y lint:

```bash
go test ./...
go vet ./...
golangci-lint run ./...
```

También hay un `Makefile` con `make build`, `make run`, `make test`, `make vet`, `make lint` y `make fmt`.

## API

### `POST /compare`

Recibe dos descripciones de request y compara sus respuestas.

```json
{
  "request1": { "url": "https://api.example.com/users/1", "method": "GET", "headers": {}, "params": {}, "body": "" },
  "request2": { "url": "https://api.example.com/users/1", "method": "GET", "headers": {}, "params": {}, "body": "" }
}
```

Cada `requestN` admite:

- `url` (requerido): URL del upstream.
- `method` (opcional): `GET` (default), `POST`, `PUT`, `PATCH`, `DELETE`, `HEAD` u `OPTIONS`.
- `headers` (opcional): mapa `clave -> valor` de headers a enviar.
- `params` (opcional): mapa `clave -> valor` de query params, combinados con los de la URL.
- `body` (opcional): cuerpo crudo de la petición.

Las peticiones al upstream se identifican con `User-Agent: comparator/1.0` salvo que el request defina uno propio.

Respuesta:

```json
{
  "status_codes": [200, 200],
  "headers": { "content-type": ["application/json", "text/html"] },
  "body_differences": [
    { "path": "user.name", "tipo": "string", "values": ["Ana", "Leo"] }
  ]
}
```

- `status_codes` siempre trae ambos códigos de estado (`[respuesta1, respuesta2]`).
- `headers` incluye solo los headers que difieren.
- `body_differences` es un array de diferencias con `path` (ruta en el JSON, ej. `a[0].b`), `tipo` (`object`, `array`, `string`, `number`, `boolean`, `null`, `mixed`, `missing`, `error`) y `values` (ambos valores; sentinels `key not found in first/second JSON` o `different lengths` para esos casos). Soporta objetos, arrays y escalares como cuerpo raíz.

Cada respuesta incluye el header `X-Request-ID` (propio o recibido) y los logs estructurados usan ese ID.

Errores:

- `400` — request inválido (falta `url`, body inválido, etc.).
- `502` — error del upstream (timeout, DNS, URL no permitida, body enorme).
- `500` — error interno.

## Frontend

- Sitio en vivo: <https://andradew.github.io/comparator/>
- Estático sin build (HTML/CSS/JS puro) en `frontend/`.
- `frontend/config.js` **no se versiona** (está en `.gitignore`); la CI de GitHub Pages lo genera con la variable `API_BASE_URL`.
- Para desarrollo local, crear `frontend/config.js`:

  ```js
  const API_BASE_URL = "http://localhost:8080";
  ```

El ledger de resultados muestra siempre los códigos de estado y agrupa las diferencias de headers y body con su conteo (`Headers · 7`, `Body · 310`). Incluye:

- **Ocultar headers volátiles**: los headers que cambian en cada respuesta (`Date`, `Age`, `Cf-Ray`, `Report-To`, `Server-Timing`, …) se ocultan por defecto para que dos respuestas idénticas den "Coinciden".
- **Filtrar diferencias por ruta**: un campo sobre el ledger que deja solo las filas cuyo path (o header) contenga el texto (`moves`, `sprites`, `[0]`).

## Variables de entorno

| Variable | Default | Descripción |
| --- | --- | --- |
| `PORT` | `8080` | Puerto del listener HTTP. |
| `HTTP_TIMEOUT` | `10s` | Timeout del cliente HTTP (formato `time.ParseDuration`). |
| `MAX_BODY_SIZE` | `1 MiB` | Límite del request entrante en bytes. |
| `MAX_RESPONSE_SIZE` | `10 MiB` | Límite del cuerpo de la respuesta upstream en bytes. |
| `CORS_ALLOWED_ORIGIN` | *(vacío)* | Origin permitido para CORS. Si no está seteada, CORS queda deshabilitado. |

## CORS

El middleware de CORS (`internal/cors`) responde el preflight `OPTIONS` con `204` y los headers de CORS únicamente si el header `Origin` del request coincide con `CORS_ALLOWED_ORIGIN`. Si el origin no coincide o la variable no está seteada, las respuestas de CORS no se agregan.

El origin de GitHub Pages es `https://andradew.github.io`; al desplegar, `CORS_ALLOWED_ORIGIN` debe coincidir con el origin desde el que se sirve el frontend.

## Estructura

```text
cmd/                        # entrypoint (wiring)
config/                     # env vars
frontend/                   # frontend estático (GitHub Pages)
internal/
  api/                      # handler HTTP y mapeo de errores
  comparator/               # lógica de comparación
  cors/                     # middleware CORS
  dtos/                     # tipos de entrada/salida de la API
  httpclient/               # cliente HTTP con timeout
  requestlog/               # request IDs y logs estructurados
  routes/                   # montado de rutas (Go 1.22 method patterns)
.github/workflows/ci.yml    # build, test, vet y golangci-lint
.github/workflows/pages.yml # deploy de frontend a GitHub Pages
```

## Convenciones

- Comentarios y TODOs en español.
- Las interfaces se definen en el consumidor (ej: `api.Handler` recibe la interfaz de `comparator.Service`).
- Errores tipados (`ValidationError`, `UpstreamError`) mapeados a `400` / `502` / `500` en el handler.
- Commits con Conventional Commits (`feat:`, `fix:`, `docs:`, ...).
- Los cambios a la API viven en ramas `feature/*` y se integran por PR.
- Los planes de trabajo y configuraciones locales no se versionan.
