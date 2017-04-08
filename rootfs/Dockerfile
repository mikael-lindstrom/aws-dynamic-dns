FROM scratch

# Add CA certificates extracted from Mozilla
ADD https://curl.haxx.se/ca/cacert.pem /etc/ssl/certs/ca-certificates.crt

COPY aws-dynamic-dns /

ENTRYPOINT ["/aws-dynamic-dns"]
