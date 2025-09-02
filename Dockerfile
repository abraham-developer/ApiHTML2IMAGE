# Builder stage
FROM golang:1.21-alpine AS builder

# Instalar dependencias de compilación
RUN apk add --no-cache \
    git \
    gcc \
    musl-dev

WORKDIR /app

# Copiar solo los archivos de dependencias primero para mejor caching
COPY go.mod go.sum ./

# Descargar dependencias
RUN go mod download

# Copiar el código fuente
COPY . .

# Compilar la aplicación
RUN go build -o main .

# Runtime stage
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

# Configurar Chromium
ENV CHROME_BIN=/usr/bin/chromium-browser \
    CHROME_PATH=/usr/lib/chromium/ \
    CHROMIUM_FLAGS="--no-sandbox --disable-dev-shm-usage --disable-gpu --headless"

# Crear usuario no-root para seguridad
RUN addgroup -S appuser && adduser -S appuser -G appuser

WORKDIR /app

# Copiar el binario compilado
COPY --from=builder --chown=appuser:appuser /app/main .

# Cambiar al usuario no-root
USER appuser

# Exponer puerto
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Comando de ejecución
CMD ["./main"]