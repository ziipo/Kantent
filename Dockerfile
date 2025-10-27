# Build stage for Go backend
FROM golang:1.21-alpine AS backend-builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev sqlite-dev

# Copy go mod files
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy backend source
COPY backend/ ./

# Build binary with SQLite support
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o rss-reader .

# Frontend build stage
FROM node:20-alpine AS frontend-builder

WORKDIR /app

COPY frontend/package*.json ./
RUN npm ci

COPY frontend/ ./
RUN npm run build

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates sqlite-libs

WORKDIR /app

# Copy backend binary
COPY --from=backend-builder /app/rss-reader .

# Copy frontend build (Go will serve these)
COPY --from=frontend-builder /app/dist ./frontend/dist

# Create data directory
RUN mkdir -p /data

EXPOSE 8080

CMD ["./rss-reader"]
