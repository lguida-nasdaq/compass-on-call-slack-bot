FROM golang:1.24 AS builder

WORKDIR /workspace
COPY . .

ENV GOCACHE=/cache/go
ENV GOMODCACHE=/cache/go/mod

COPY misc/ZscalerRootCertificate-2048-SHA256.crt /usr/local/share/ca-certificates/ZscalerRootCertificate-2048-SHA256.crt
RUN update-ca-certificates

RUN --mount=type=ssh --mount=type=cache,target=/cache/go \
    mkdir -p $GOMODCACHE && \
    go mod tidy

RUN --mount=type=cache,target=/cache/go \
    CGO_ENABLED=0 go build -o /server ./cmd/server

FROM gcr.io/distroless/static:nonroot AS runner

COPY --from=builder /server /server

ENTRYPOINT ["/server"]