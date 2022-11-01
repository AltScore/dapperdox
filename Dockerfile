FROM alpine:latest

COPY --from=golang:1.13-alpine /usr/local/go/ /usr/local/go/
ENV PATH="/usr/local/go/bin:${PATH}"

ENTRYPOINT ["/dapperdox/dapperdox"]
WORKDIR /dapperdox
COPY assets /dappperdox/assets
RUN mkdir -p /dapperdox/main
COPY . /dapperdox/data

RUN cd data && go build -o /dapperdox/dapperdox
RUN rm -rf /dapperdox/data

RUN chmod u+x /dapperdox



# Adduser without home (-H) no shell (-s)
RUN adduser -H -h /dapperdox -D -s /bin/false go-exec
RUN chown -R go-exec /dapperdox


USER go-exec

EXPOSE 3123


