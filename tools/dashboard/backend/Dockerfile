# --- Stage 1: Build ---
# Use an official Go image as a builder image
FROM golang:1.21-alpine AS builder

# Set the working directory
WORKDIR /app

ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown
ARG GO_VERSION=unknown

# Copy the Go module files
# We copy them first to leverage Docker layer caching
COPY go.mod go.sum ./
# Because of the 'replace' directive, we need the parent module's go.mod
COPY ../go.mod ../go.sum ../

# Copy the source code of the backend and the operator API definitions
COPY . .
COPY ../api ../api

# Download dependencies. The replace directive will be respected.
RUN go mod tidy
RUN go mod download

# Build the Go app, creating a static binary
# CGO_ENABLED=0 is important for a static build
# -ldflags="-w -s" strips debug information to reduce binary size
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags="-w -s \
    -X 'polardbx-dashboard-backend/pkg/api/router.Version=${VERSION}' \
    -X 'polardbx-dashboard-backend/pkg/api/router.Commit=${COMMIT}' \
    -X 'polardbx-dashboard-backend/pkg/api/router.BuildDate=${BUILD_DATE}' \
    -X 'polardbx-dashboard-backend/pkg/api/router.GoVersion=${GO_VERSION}'" \
    -o /polardbx-dashboard-backend ./main.go

# --- Stage 2: Final ---
# Use a minimal image for the final stage
FROM scratch

# Copy the static binary from the builder stage
COPY --from=builder /polardbx-dashboard-backend /polardbx-dashboard-backend

# Expose the port the app runs on
EXPOSE 8080

# The command to run the application
ENTRYPOINT ["/polardbx-dashboard-backend"] 