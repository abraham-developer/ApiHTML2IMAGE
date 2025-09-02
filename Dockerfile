FROM golang:1.21-alpine AS builder

# Instalar dependencias de compilación
RUN apk add --no-cache \
    git \
    gcc \
    musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main .

# Runtime image con Alpine + Chromium
FROM alpine:3.18

# Instalar Chromium y dependencias mínimas
RUN apk add --no-cache \
    chromium \
    chromium-chromedriver \
    nss \
    freetype \
    harfbuzz \
    ca-certificates \
    ttf-freefont \
    # Dependencias para Go runtime
    libc6-compat

# Configurar Chromium para Alpine
ENV CHROME_BIN=/usr/bin/chromium-browser \
    CHROME_PATH=/usr/lib/chromium/ \
    CHROMEDRIVER_PATH=/usr/bin/chromedriver \
    # Configuraciones de seguridad
    CHROMIUM_FLAGS="--no-sandbox --disable-dev-shm-usage --disable-gpu --headless"

# Crear usuario no-root
RUN addgroup -S appuser && adduser -S appuser -G appuser

WORKDIR /app

# Copiar binario
COPY --from=builder --chown=appuser:appuser /app/main .

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["./main"]