FROM golang:1.12.9-alpine3.10 AS build-env

# Add Dependencies required for building the binaries
RUN apk --no-cache add build-base git bzr mercurial gcc
RUN apk --no-cache add curl wget
RUN curl https://glide.sh/get | sh
