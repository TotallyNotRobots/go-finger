# Load golang image
FROM golang:1.26-alpine@sha256:91eda9776261207ea25fd06b5b7fed8d397dd2c0a283e77f2ab6e91bfa71079d AS builder

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
FROM gcr.io/distroless/static-debian12:nonroot@sha256:a9329520abc449e3b14d5bc3a6ffae065bdde0f02667fa10880c49b35c109fd1

WORKDIR /app

COPY urns.yml /app/urns.yml

# Set our runtime environment
ENV ENV_DOCKER=true

COPY --from=builder /go/src/app/finger /usr/local/bin/finger

HEALTHCHECK CMD [ "finger", "healthcheck" ]

EXPOSE 8080

ENTRYPOINT [ "finger" ]
CMD [ "serve" ]
