FROM golang:1.20.3-alpine3.17 AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o /app/app .

FROM alpine:3.17
COPY --from=builder /app/app /usr/bin/app
ENTRYPOINT ["/usr/bin/app"]
CMD ["run"]

