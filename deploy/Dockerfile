# Runtime image for the b-eval platform.
#
# The image carries a cross-compiled static binary and nothing else: the
# frontend bundle is embedded in the binary by the Go embed directive, so
# the runtime needs no Go toolchain, no Node and no static-file server.
#
# The base image is a build argument because the server cannot reach
# registry-1.docker.io. A registry mirror provides the same Alpine image,
# and server_setup.sh tags it under the name used here.
ARG BASE_IMAGE=alpine:3.20
FROM ${BASE_IMAGE}

# ca-certificates is required for the HTTP adapter to reach an agent under
# test over HTTPS. tzdata lets the container's clock follow the operator's
# zone; the platform itself writes every timestamp in UTC.
RUN apk add --no-cache ca-certificates tzdata \
 && adduser -D -u 10001 beval

WORKDIR /app
COPY dist/b-eval-linux-amd64 /app/b-eval

USER 10001
ENV TZ=Asia/Shanghai
EXPOSE 8091

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s \
  CMD wget -qO- http://127.0.0.1:8091/api/healthz || exit 1

# The container must listen on 0.0.0.0: a published port is forwarded to the
# container's IP address, not to its loopback interface.
ENTRYPOINT ["/app/b-eval", "serve", "-addr", "0.0.0.0:8091", "-data", "/app/data"]
