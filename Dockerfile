FROM golang:1.18 as builder

WORKDIR /app-controller
COPY . .
RUN GOPROXY="https://goproxy.cn" GO111MODULE=on CGO_ENABLED=0 go build -o exec main.go

FROM alpine:3.15.3

WORKDIR /app-controller
COPY --from=builder /app-controller/exec .

ENTRYPOINT ["./exec"]

