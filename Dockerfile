FROM scratch
COPY shepard /shepard
ENTRYPOINT ["/goreleaser"]
