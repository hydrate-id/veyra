FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o veyra .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/veyra /veyra

EXPOSE 8080
EXPOSE 8443
EXPOSE 2019

VOLUME ["/data"]

ENTRYPOINT ["/veyra"]
CMD ["start", "--db", "/data/veyra.db", "--http-port", "8080", "--https-port", "8443"]
