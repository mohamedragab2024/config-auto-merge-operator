FROM golang:1.20 as builder

WORKDIR /workspace
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o manager cmd/manager/main.go

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/manager .
USER 65532:65532

ENTRYPOINT ["/manager"] 
