########## Builder ##########
FROM golang:1.22-alpine AS builder

# Copy local source
COPY . /build
WORKDIR /build

# Build the binary
RUN go build -o manager ./cmd/main.go

######## Binary ###########

FROM mcr.microsoft.com/cbl-mariner/distroless/minimal:2.0-nonroot

WORKDIR /

# Kubernetes runAsNonRoot requires USER to be numeric
USER 65532:65532

COPY --from=builder /build/manager /manager

ENTRYPOINT [ "/manager" ]
