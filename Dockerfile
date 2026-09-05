# Load golang image
FROM golang:1.27-alpine@sha256:cf6fca6641884b8433441b2b0652976f975e1d0fdd26d177eaaf8596087f3125 AS builder

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
FROM gcr.io/distroless/static-debian12:nonroot@sha256:afa5c872c891853ca7fcf1f12c3edb23f7eeef36189728842dd51042ff57f7ab

WORKDIR /app

COPY urns.yml /app/urns.yml

# Set our runtime environment
ENV ENV_DOCKER=true

COPY --from=builder /go/src/app/finger /usr/local/bin/finger

HEALTHCHECK CMD [ "finger", "healthcheck" ]

EXPOSE 8080

ENTRYPOINT [ "finger" ]
CMD [ "serve" ]
