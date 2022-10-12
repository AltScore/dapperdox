FROM alpine:latest

WORKDIR /dapperdox

COPY dapperdox /dapperdox

# Adduser without home (-H) no shell (-s)
RUN adduser -H -h /dapperdox -D -s /bin/false go-exec
RUN chown -R go-exec /dapperdox

USER go-exec

EXPOSE 3123

ENTRYPOINT [ "./dapperdox" ]
