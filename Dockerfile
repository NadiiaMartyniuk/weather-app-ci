# Etap 1
FROM golang:1.24-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o weatherapp main.go

# Etap 2
FROM alpine:latest
WORKDIR /app

# Dodajemy certyfikaty SSL i narzędzie curl (dla HEALTHCHECK)
RUN apk --no-cache add ca-certificates curl

COPY --from=builder /app/weatherapp .
LABEL maintainer="NadiiaMartyniuk"
EXPOSE 8080

# Healthcheck – sprawdzamy co 30 sekund, czy serwer HTTP odpowiada
HEALTHCHECK --interval=30s --timeout=3s \
  CMD curl -f http://localhost:8080/health || exit 1


CMD ["./weatherapp"]
