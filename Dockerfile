# -------- Setup and build the golang binary -------- #
FROM harshanarayana/golang:latest AS build-env

# Clone and setup the repo
RUN mkdir -p $GOPATH/src/github.com/harshanarayana/
WORKDIR $GOPATH/src/github.com/harshanarayana/

# Selectively disable cache for this
ARG CACHEBUST=1
RUN git clone https://github.com/harshanarayana/go-webhook.git

# Build the binaries as require
WORKDIR $GOPATH/src/github.com/harshanarayana/go-webhook

# Selectively disable cache for this
ARG CACHEBUST=1
RUN glide install && glide update && go build .
RUN cp $GOPATH/src/github.com/harshanarayana/go-webhook/go-webhook /tmp/go-webhook

# -------- Setup and build the runtime image -------- #
FROM alpine

RUN mkdir -p /config
RUN mkdir -p /data/ssl
COPY ssl/*.pem /data/ssl/
COPY --from=build-env /tmp/go-webhook /go-webhook
RUN chmod 755 /go-webhook

ENTRYPOINT ["/go-webhook"]
