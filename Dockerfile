# ─── Stage 1: Build React frontend ─────────────────────────────────────
FROM node:23-alpine AS frontend-builder

WORKDIR /frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci

COPY frontend/ ./
RUN npm run build   # outputs to /frontend/dist

# ─── Stage 2: Build Go backend ────────────────────────────────────────
FROM golang:1.24-alpine AS backend-builder

RUN apk add --no-cache git

WORKDIR /app
# cache deps
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# copy backend source
COPY backend/ ./

# bring in built static files
COPY --from=frontend-builder /frontend/dist ./frontend/dist

RUN go build -o app ./cmd/main.go

# ─── Stage 3: Final minimal image ─────────────────────────────────────
FROM alpine:latest AS debug

WORKDIR /app/backend

# copy binary + static assets
RUN mkdir backend

COPY --from=backend-builder /app/app ./app
COPY --from=backend-builder /app/.env ./.env
COPY --from=backend-builder /app/frontend/dist /app/frontend/dist

ENV GIN_MODE=release

EXPOSE 8080

ENTRYPOINT ["./app"]
