FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.work go.work.sum ./
COPY app/go.mod app/go.sum ./app/
COPY domain/tracking/go.mod ./domain/tracking/
COPY infrastructure/go.mod ./infrastructure/
COPY pkg/go.mod ./pkg/

RUN cd app && go mod download

COPY . .
RUN go build -o /tracking-api ./app

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /tracking-api /tracking-api
COPY app/etc /etc/app

EXPOSE 8082
HEALTHCHECK --interval=15s --timeout=5s CMD wget -qO- http://localhost:8082/health || exit 1
ENTRYPOINT ["/tracking-api", "-f", "/etc/app/app.yaml"]
