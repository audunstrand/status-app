# Multi-service Dockerfile
# Build with: docker build --build-arg SERVICE=<name> -t <name> .
# Services: backend, scheduler, slackbot, migrate

FROM golang:1.24-alpine AS builder

ARG SERVICE=backend

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/app ./cmd/${SERVICE}

# Build migrate tool for all services
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/migrate ./cmd/migrate

FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY --from=builder /bin/app /bin/app
COPY --from=builder /bin/migrate /bin/migrate

# Copy migrations directory for all services
COPY migrations /migrations

EXPOSE 8080

CMD ["/bin/app"]
