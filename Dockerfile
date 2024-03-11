FROM alpine:3.19.1

 RUN apk update && apk --no-cache add ca-certificates && \
  update-ca-certificates

 ADD ./release-exporter /usr/local/bin/release-exporter
ENTRYPOINT ["/usr/local/bin/release-exporter"]
