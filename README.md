# <img src="static/veyra.png" alt="Veyra Logo" height="32" align="center" /> Veyra

Veyra is a single-binary local reverse proxy control plane based on embedded Caddy. It manages domain mapping, reverse proxy rules, and dynamic config reloading for local development and self-hosted environments.

Veyra operates as a control plane that orchestrates Caddy as the data plane, exposing a clean command-line interface and configuration registry.

## The Problem It Solves

When exposing local web applications to the public internet using tunnels, absolute URLs, session cookies, and redirects can break due to domain and port mismatches.

Veyra acts as an ingress control plane to translate and bridge public and private domains cleanly. For example:

```
Internet -> public.example.com -> Tunnel -> Veyra -> local.example.test -> Local Application
```

> [!IMPORTANT]
> **Tunnel Configuration**: You must configure your tunnel service client (e.g., Pangolin, ngrok, or Cloudflare Tunnel) to forward incoming public traffic to Veyra's local port (default: HTTP `8080` or HTTPS `8443`) rather than routing it directly to the local application's port. Veyra will then receive the requests and proxy them to the correct local target (`local.example.test`).

- **Request Translation**: Incoming requests for a public domain (`public.example.com`) are dynamically translated into the internal target host (`local.example.test`).
- **Response URL Rewriting**: Rewrites absolute URL references in response bodies (HTML, JS, CSS, JSON) from the backend host back to the client's public domain.
- **Header & Cookie Mapping**: Adjusts `Location` headers for redirects, domain mappings in `Set-Cookie` headers, and forwards standard proxy headers (`X-Forwarded-Host`, `X-Forwarded-Port`, `X-Forwarded-Proto`).

## Key Features

- **Embedded Caddy**: Bundled Caddy runtime within a single Go binary.
- **BoltDB Registry**: Persistent storage for all rules and proxy settings.
- **Zero Downtime Hot Reload**: Dynamic configuration updates via Caddy's Admin API.
- **Wildcard Capture & Flatten**: Flexible subdomain routing and capture backreferences.
- **Minimalist CLI**: Easy-to-use commands for rule and port management.

## CLI Usage

### Start the Daemon

Starts the proxy control plane daemon (default HTTP: `8080`, HTTPS: `8443`, Admin: `2019`).

```bash
veyra start
```

### Add a Route

Maps a public hostname to an internal target.

```bash
veyra add public.example.com https://local.example.test
```

### Remove a Route

```bash
veyra remove public.example.com
```

### List Routes

Lists all registered routing rules.

```bash
veyra list
```

### Get Route Details

```bash
veyra get public.example.com
```

### Validate Configuration

Validates rules format, detects conflicts, and checks adapted Caddyfile syntax.

```bash
veyra validate
```

### Reload Caddy

```bash
veyra reload
```

### Export Configuration

Exports active Caddyfile, rules JSON, or full registry backup.

```bash
veyra export caddy
veyra export json rules.json
veyra export backup backup.json
```

### Import Configuration

```bash
veyra import json rules.json
veyra import backup backup.json
```

## Docker Deployment

You can pull the official multi-architecture Docker image from GitHub Container Registry (GHCR):

```bash
docker pull ghcr.io/korfra/veyra:latest
docker run -d -p 8080:8080 -p 8443:8443 -p 2019:2019 -v veyra_data:/data ghcr.io/korfra/veyra:latest
```

### Docker Compose

Configure a `compose.yml` to run Veyra:

```yaml
services:
  veyra:
    image: ghcr.io/korfra/veyra:latest
    ports:
      - "8080:8080"
      - "8443:8443"
      - "2019:2019"
    volumes:
      - veyra-data:/data
    restart: unless-stopped

volumes:
  veyra-data:
```

Launch Veyra with Docker Compose:

```bash
docker compose up -d
```

## License

This project is licensed under [MIT License](LICENSE).
