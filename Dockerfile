FROM alpine
ADD gopath/bin/gif-maker /gif-maker
ENTRYPOINT ["/gif-maker"]
