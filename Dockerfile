# Multi-service Dockerfile
# Build with: docker build --build-arg SERVICE=<name> -t <name> .
# Services: api, commands, projections, scheduler, slackbot

FROM golang:1.24-alpine AS builder

ARG SERVICE=api

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/app ./cmd/${SERVICE}

FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY --from=builder /bin/app /bin/app

EXPOSE 8080

CMD ["/bin/app"]
