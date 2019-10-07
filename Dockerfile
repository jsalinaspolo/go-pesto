FROM golang:1.12.7 AS builder
WORKDIR /build/
COPY . .
RUN GO111MODULE=on CGO_ENABLED=0 GOOS=linux go build -tags netgo -a -v -o /app .

FROM alpine:3.8
COPY --from=builder /app /
COPY .zopa_certs /usr/local/share/ca-certificates/
RUN apk add --update ca-certificates openssl
RUN apk add --no-cache libc6-compat
RUN update-ca-certificates
RUN rm -rf /var/cache/apk/*
RUN mkdir -p /opt/certs
RUN openssl req -subj "/C=UK/ST=London/L=London/O=Zopa/OU=Development/CN=*" \
                -x509 -nodes -days 3650 -newkey rsa:2048 \
                -keyout /opt/certs/tls.key \
                -out /opt/certs/tls.crt

ENTRYPOINT ["/app"]