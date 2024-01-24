FROM golang:1.21-bookworm as build

WORKDIR /go/src/github.com/eliblock/less-advanced-security

# Ordered from least likely to most likely to be updated.
COPY *.go ./
COPY ./sarif ./sarif
COPY ./github ./github
COPY go.mod go.sum ./

RUN go mod download
RUN CGO_ENABLED=0 go build -o /go/bin/less-advanced-security

FROM gcr.io/distroless/base-debian12 AS distroless
COPY --from=build /go/bin/less-advanced-security /usr/bin/

ENTRYPOINT ["/usr/bin/less-advanced-security"]
