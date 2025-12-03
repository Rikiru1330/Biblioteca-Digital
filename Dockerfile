# Dockerfile - Sin CGO para evitar problemas
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copiar archivos de dependencias
COPY go.mod go.sum ./
RUN go mod download

# Copiar código fuente
COPY . .

# Compilar SIN CGO
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o library-api main.go

# Etapa de ejecución
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copiar binario
COPY --from=builder /app/library-api .

# Exponer puerto
EXPOSE 8080

# Variables de entorno
ENV GIN_MODE=release
ENV STORAGE_TYPE=memory 

# Comando
CMD ["./library-api"]