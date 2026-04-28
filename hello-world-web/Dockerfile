FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY go.mod ./
COPY main.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/hello-world-web ./...

FROM scratch
COPY --from=builder /out/hello-world-web /hello-world-web
EXPOSE 9090
ENTRYPOINT ["/hello-world-web"]
