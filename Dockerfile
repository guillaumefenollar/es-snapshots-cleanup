FROM golang:1.17-alpine AS builder
WORKDIR /build
ADD main.go go.mod ./
RUN go build -o /build/app

FROM alpine
COPY --from=builder /build/app /usr/bin/app
ENTRYPOINT ["/usr/bin/app"]
