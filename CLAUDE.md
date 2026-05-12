# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

hc-itms is a self-hosted app distribution platform for iOS IPA (OTA install) and Android APK. It's a single Go binary with an embedded Vue 3 frontend.

## Build & Run Commands

```bash
# Development (run backend + frontend separately)
go run main.go                              # Backend on :8080
cd web && npm run dev                       # Vite dev server on :5173 (proxies /api and /d to :8080)

# Production build (single binary with embedded frontend)
cd web && npm install && npm run build      # Build frontend → web/dist/
go build -o hc-itms.exe .                   # Embeds web/dist/ via go:embed

# Makefile shortcuts
make dev-backend
make dev-frontend
make frontend
make build
make clean
```

## Architecture

**Backend-frontend connection:**
- Dev mode: Vite dev server proxies `/api/*` and `/d/*` to Go backend at `localhost:8080`
- Production: `go:embed all:web/dist` embeds the built frontend into the Go binary. `main.go`'s `setupSPA()` serves embedded files via `r.NoRoute()`, falling back to `index.html` for SPA routing.

**Route structure (defined in `main.go`):**
- `/api/auth/*` — public auth endpoints (login, register)
- `/api/apps/*`, `/api/versions/*` — JWT-protected admin CRUD
- `/d/:id/*` — public download routes (download page, IPA/APK files, plist manifest, icon, QR code)

**Key service layer:**
- `services/ipa.go` — extracts metadata + icon from IPA (ZIP + plist parsing)
- `services/apk.go` — extracts metadata + icon from APK (androidbinary library)
- `services/cgbi.go` — converts Apple's non-standard CgBI PNG to standard PNG (strip CgBI chunk, concatenate IDAT streams, decompress raw deflate, recompress standard deflate)
- `services/manifest.go` — generates Apple OTA plist XML for `itms-services://` protocol
- `services/icon.go` — saves icons with CgBI fallback, creates placeholder JPEGs

**Database:** GORM auto-migration (no migration files). Three models: `App`, `Version` (in `models/app.go`), `User` (in `models/user.go`). Uses pure-Go SQLite driver (`glebarez/sqlite`, no CGO required).

**Auth:** JWT Bearer tokens. Frontend stores token in `localStorage`, Axios interceptor attaches it to all `/api` requests.

## Configuration

`config.yaml` at project root. Key fields: `server.port`, `server.base_url` (must be HTTPS for iOS OTA), `jwt.secret`, `storage.upload_dir`, `storage.icon_dir`, `storage.max_file_size`.

## Important Implementation Details

- Registration is first-user-only (no open signup after initial admin is created).
- iOS OTA install requires `base_url` to be HTTPS — the plist manifest URLs are built from this config value.
- IPA icon extraction handles Apple's CgBI PNG format (non-standard deflate). The conversion in `cgbi.go` tries raw deflate first, then zlib-wrapped as fallback.
- The upload handler in `handlers/version.go` creates a new app if the bundle_id doesn't exist yet (app ID 0 in the URL signals "new app").
- Frontend upload uses Element Plus `el-upload` with `:limit="1"` — the `Upload.vue` component manages file state manually via ref to support re-selection after upload.
