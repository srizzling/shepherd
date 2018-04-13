FROM alpine:3.7
COPY shepherd /shepherd
ENTRYPOINT ["/shepherd"]
