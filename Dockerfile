######################
# build image
######################
FROM golang:1.15.7 AS build

ARG GIT_TAG
COPY . .
RUN \
  GOPATH="" CGO_ENABLED=0 go build -a -ldflags "-extldflags '-static' -X main.AppVersion=${GIT_TAG:=HEAD}" -o /tmp/freenas-iscsi-provisioner

FROM alpine

RUN apk update && \
    apk add ca-certificates && \
    rm -rf /var/cache/apk/* && \
    update-ca-certificates

COPY --from=build /tmp/freenas-iscsi-provisioner /freenas-iscsi-provisioner

ENTRYPOINT ["/freenas-iscsi-provisioner"]
