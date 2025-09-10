# --- Etapa 1: Builder ---
FROM golang:1.25.1-alpine AS builder

WORKDIR /app

# Copia archivos de gestión de dependencias
COPY go.mod go.sum ./
RUN go mod download

# Copia el resto del código fuente
COPY . .

# Compila la aplicación
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix 'static' -o app ./cmd/server

# --- Etapa 2: Final ---
FROM gcr.io/distroless/static-debian11

WORKDIR /

# Copia el binario compilado
COPY --from=builder /app/app .


# Expone el puerto
EXPOSE 3000

# Comando para ejecutar
ENTRYPOINT ["/app"]