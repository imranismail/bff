FROM alpine
COPY bff /usr/local/bin
WORKDIR /srv
ENTRYPOINT ["/usr/local/bin/bff"]