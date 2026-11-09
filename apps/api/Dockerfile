FROM golang:1.26-alpine AS build

WORKDIR /app

# Copiar archivos de dependencias desde apps/api
COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download

# Copiar el código fuente y compilar
COPY apps/api/ ./
RUN CGO_ENABLED=0 go build -o /app/main .

# Runtime ligero
FROM alpine:3.21
WORKDIR /app
RUN apk update && apk upgrade --no-cache && \
    addgroup -S nonroot -g 10001 && adduser -S nonroot -G nonroot -u 10001
COPY --from=build --chown=10001:10001 /app/main /app/main

USER 10001:10001
EXPOSE 8080
CMD ["/app/main"]
