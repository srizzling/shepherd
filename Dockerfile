FROM scratch
COPY shepard /shepard
ENTRYPOINT ["/shepard"]
