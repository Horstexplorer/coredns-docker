FROM golang:1.26-alpine AS builder

ARG COREDNS_VERSION=v1.14.2

RUN apk add --no-cache git

WORKDIR /src
COPY . /src/coredns-docker-plugin

RUN git clone --depth 1 --branch ${COREDNS_VERSION} https://github.com/coredns/coredns.git /src/coredns

# Register the local plugin in CoreDNS's directive list.
RUN printf '\ncoredns-docker:github.com/horstexplorer/coredns-docker\n' >> /src/coredns/plugin.cfg

RUN cd /src/coredns \
    && go mod edit -replace=github.com/horstexplorer/coredns-docker=/src/coredns-docker-plugin \
    && go mod edit -require=github.com/horstexplorer/coredns-docker@v0.0.0 \
    && go generate \
    && go mod tidy \
    && CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/coredns

FROM alpine:3.22

RUN apk add --no-cache ca-certificates

COPY --from=builder /out/coredns /coredns

EXPOSE 53/tcp
EXPOSE 53/udp

ENTRYPOINT ["/coredns", "-conf", "/Corefile"]

