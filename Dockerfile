FROM golang:1.18 as builder

WORKDIR /app-controller
COPY . .
RUN CGO_ENABLED=0 go build -o app-controller-exec main.go

FROM alpine:3.15.3

WORKDIR /app-controller
COPY --from=builder /app-controller-exec .

CMD ["./app-controller-exec"]

