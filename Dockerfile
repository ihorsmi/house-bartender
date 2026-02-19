# ---- build ----
FROM golang:1.22-alpine AS build

RUN apk add --no-cache build-base

WORKDIR /src

# Copy mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download || true

# Copy the rest
COPY . .

# ---- build ----
FROM golang:1.22-alpine AS build

RUN apk add --no-cache build-base
WORKDIR /src

# Copy mod files first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest
COPY . .

# IMPORTANT: host COPY may overwrite go.sum; fix it inside the build
RUN go mod tidy

ENV CGO_ENABLED=1
RUN go build -trimpath -ldflags="-s -w" -o /out/housebartender ./cmd/housebartender

# ---- runtime ----
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata && update-ca-certificates

# Non-root
RUN addgroup -S app && adduser -S -G app -u 10001 app

WORKDIR /app
COPY --from=build /out/housebartender /app/housebartender
COPY views/ /app/views/
COPY static/ /app/static/

ENV ADDR=:8080
ENV DATA_DIR=/data
ENV DB_PATH=/data/housebartender.db
ENV UPLOAD_DIR=/data/uploads

# ✅ Create data dirs and ensure ownership for the non-root user
RUN mkdir -p /data/uploads && chown -R app:app /data

USER app
EXPOSE 8080
ENTRYPOINT ["/app/housebartender"]
