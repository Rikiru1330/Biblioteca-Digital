# Dockerfile - Sin CGO para usar modernc.org/sqlite
FROM golang:1.25-alpine AS go-builder

WORKDIR /app

# Copiar archivos de dependencias
COPY go.mod go.sum ./
RUN go mod download

# Copiar código fuente
COPY . .

# Compilar SIN CGO
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o library-api main.go

# Etapa de ejecución
FROM alpine:latest AS runner

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copiar binario
COPY --from=go-builder /app/library-api .

# Crear directorio para datos
RUN mkdir -p /data

# Exponer puerto
EXPOSE 8080

# Variables de entorno POR DEFECTO (se sobrescriben desde docker-compose)
ENV GIN_MODE=release
ENV STORAGE_TYPE=sqlite
ENV DB_PATH=/data/library.db

# Comando
CMD ["./library-api"]