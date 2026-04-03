# Stage 1: Build Vue frontend
FROM node:22-alpine AS frontend
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ .
RUN npm run build

# Stage 2: Build Go binary
FROM golang:1.25-alpine AS backend
WORKDIR /app
COPY go-backend/go.mod go-backend/go.sum ./
RUN go mod download
COPY go-backend/ .
RUN CGO_ENABLED=0 go build -o /flourish ./cmd/flourish/

# Stage 3: Final image
FROM alpine:3.21
WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=backend /flourish /usr/local/bin/flourish
COPY config/ config/
COPY --from=frontend /app/frontend/dist frontend/dist

EXPOSE 8080
CMD ["flourish", "serve", "--port", "8080"]
