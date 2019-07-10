FROM alpine

RUN apk update && \
    apk add ca-certificates && \
    rm -rf /var/cache/apk/* && \
    update-ca-certificates

COPY bin_release/freenas-iscsi-provisioner_linux-amd64 /freenas-iscsi-provisioner
ENTRYPOINT ["/freenas-iscsi-provisioner"]
