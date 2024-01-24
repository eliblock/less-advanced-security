FROM golang:1.21-bookworm as build

WORKDIR /go/src/github.com/eliblock/less-advanced-security

# Ordered from least likely to most likely to be updated.
COPY *.go ./
COPY ./sarif ./sarif
COPY ./github ./github
COPY go.mod go.sum ./

RUN go mod download
RUN CGO_ENABLED=0 go build -o /go/bin/less-advanced-security

# Flavor running Debian's slim distribution flavor.
# Useful for users that need system commands (shells, package managers, utilities, etc.)
# Such use case can be GitHub workflows' `jobs.<job_id>.container` option which doesn't
# support distroless and rootless flavors.
FROM debian:bookworm-slim AS cli

# The slim image doesn't install ca-certificates by default
# this leads to such errors in Go: x509: certificate signed by unknown authority.
# Thus we need to install it manually.
RUN set -eux; \
    apt-get update; \
    apt-get install -y --no-install-recommends ca-certificates; \
    rm -r /var/lib/apt/lists/*

RUN useradd --user-group --create-home --uid 1001 --shell /bin/bash user
USER user:user

COPY --from=build /go/bin/less-advanced-security /usr/bin/

ENTRYPOINT ["/usr/bin/less-advanced-security"]

# Distroless flavor removes unneccessary binaries and Debian components
# allowing to keep the image size to the bare minimum.
FROM gcr.io/distroless/base-debian12 AS distroless-cli
COPY --from=build /go/bin/less-advanced-security /usr/bin/
ENTRYPOINT ["/usr/bin/less-advanced-security"]
