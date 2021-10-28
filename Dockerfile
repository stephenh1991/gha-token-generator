FROM golang:1.17 AS builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

COPY cmd/ cmd/

RUN --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o token-generator ./cmd/token-generator

FROM gcr.io/distroless/static:nonroot

WORKDIR /

COPY --from=builder /workspace/token-generator .

USER 65532:65532

ENTRYPOINT ["/token-generator"]
