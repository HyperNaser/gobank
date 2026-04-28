# Build stage
FROM golang:1.26-alpine3.23 AS builder
WORKDIR /app

# caching dependencies
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# CGO_ENABLED=0: statically linked binary (self contained)
# -ldflags="-s -w": strip debugging information (smaller size)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o main main.go
# install curl
RUN apk add curl
# download golang-migrate
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.19.1/migrate.linux-amd64.tar.gz | tar xvz

# Run stage
FROM alpine:3.23
# create a group and user
RUN addgroup -S bankgroup && adduser -S bankuser -G bankgroup

WORKDIR /app

#change ownership to created user in created group instead of root (security)
COPY --from=builder --chown=bankuser:bankgroup /app/main .
COPY --from=builder /app/migrate .
# temporary dev config
COPY app.env .
COPY start.sh .
COPY db/migration ./migration

USER root
RUN chmod +x start.sh && chown -R bankuser:bankgroup /app

USER bankuser

EXPOSE 8080

ENTRYPOINT [ "/app/start.sh" ]
CMD [ "/app/main" ]