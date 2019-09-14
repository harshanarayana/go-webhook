# -------- Setup and build the golang binary -------- #
FROM golang:1.12.9-alpine3.10 AS build-env

# Add Dependencies required for building the binaries
RUN apk --no-cache add build-base git bzr mercurial gcc
RUN apk --no-cache add curl wget
RUN curl https://glide.sh/get | sh

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

#RUN mkdir -p /data/ssl
#COPY ssl/*.pem /data/ssl/
COPY --from=build-env /tmp/go-webhook /go-webhook
RUN chmod 755 /go-webhook

ENTRYPOINT ["go-webhook"]
