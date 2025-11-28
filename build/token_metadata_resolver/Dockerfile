# ---------------------------------------------------------------------
#  The first stage container, for building the application
# ---------------------------------------------------------------------
FROM golang:1.25.1-alpine AS builder

ENV CGO_ENABLED=0
ENV GOOS=linux

RUN apk --no-cache add ca-certificates

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ cmd/
COPY internal/ internal/
COPY pkg/ pkg/

RUN go build -ldflags="-w -s" -a -o /go/bin/token_metadata_resolver ./cmd/token_metadata_resolver/

# ---------------------------------------------------------------------
#  The second stage container, for running the application
# ---------------------------------------------------------------------
FROM scratch

WORKDIR /app

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/token_metadata_resolver /app/token_metadata_resolver
COPY configs/dipdup.yml /app/config.yml

ENTRYPOINT ["/app/token_metadata_resolver", "-c", "/app/config.yml"]