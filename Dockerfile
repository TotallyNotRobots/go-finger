# Load golang image
FROM golang:1.24-alpine@sha256:8f8959f38530d159bf71d0b3eb0c547dc61e7959d8225d1599cf762477384923 AS builder

RUN apk add make

ARG VERSION=undefined

WORKDIR /go/src/app

# Set our build environment
ENV GOCACHE=/tmp/.go-build-cache
# This variable communicates to the service that it's running inside
# a docker container.
ENV ENV_DOCKER=true

# Copy dockerignore files
COPY .dockerignore ./

# Install go deps using the cache
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/tmp/.go-build-cache \
  go mod download -x

COPY Makefile ./

# Copy source files
COPY main.go ./
COPY cmd cmd
COPY internal internal
COPY webfingers webfingers
COPY handler handler

# Build it
RUN --mount=type=cache,target=/tmp/.go-build-cache \
  make build VERSION=$VERSION

# Now create a new image with just the binary
FROM gcr.io/distroless/static-debian12:nonroot@sha256:627d6c5a23ad24e6bdff827f16c7b60e0289029b0c79e9f7ccd54ae3279fb45f

WORKDIR /app

COPY urns.yml /app/urns.yml

# Set our runtime environment
ENV ENV_DOCKER=true

COPY --from=builder /go/src/app/finger /usr/local/bin/finger

HEALTHCHECK CMD [ "finger", "healthcheck" ]

EXPOSE 8080

ENTRYPOINT [ "finger" ]
CMD [ "serve" ]
