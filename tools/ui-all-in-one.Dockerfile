FROM node:18-alpine AS frontend-build

WORKDIR /workspace/polardbx-ui
COPY polardbx-ui/package*.json ./
RUN npm ci
COPY polardbx-ui/ ./
RUN npm run build

FROM golang:1.24-alpine AS backend-build

WORKDIR /workspace
# Copy root go.mod first (for replace directives)
COPY go.mod go.sum ./
# Copy backend go.mod
COPY backend/go.mod backend/go.sum ./backend/
# Download dependencies
RUN go mod download
# Copy source code: backend, api, and pkg
# Note: backend has its own pkg/api/router, root pkg/ is for operator packages
COPY backend/ ./backend/
COPY api/ ./api/
COPY pkg/ ./pkg/
# The replace directive in backend/go.mod points to ../, so structure should be:
# /workspace/backend/ (backend code)
# /workspace/api/ (api definitions)
# /workspace/pkg/ (operator packages like hpfs, etc.)
# Build from backend directory
WORKDIR /workspace/backend
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/polardbx-ui-backend ./main.go

FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /app
COPY --from=backend-build /bin/polardbx-ui-backend /app/polardbx-ui-backend
# Angular build outputs to dist/polardbx-ui/browser/ (new Angular build system)
COPY --from=frontend-build /workspace/polardbx-ui/dist/polardbx-ui/browser/ /app/ui/

ENV UI_STATIC_DIR=/app/ui
EXPOSE 8080

ENTRYPOINT ["/app/polardbx-ui-backend"]


