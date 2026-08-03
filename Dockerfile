# Load golang image
FROM golang:1.26-alpine@sha256:0178a641fbb4858c5f1b48e34bdaabe0350a330a1b1149aabd498d0699ff5fb2 AS builder

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
FROM gcr.io/distroless/static-debian12:nonroot@sha256:f5b485ea962d9bd1186b2f6b3a061191539b905b82ec395de78cbfae51f20e35

WORKDIR /app

COPY urns.yml /app/urns.yml

# Set our runtime environment
ENV ENV_DOCKER=true

COPY --from=builder /go/src/app/finger /usr/local/bin/finger

HEALTHCHECK CMD [ "finger", "healthcheck" ]

EXPOSE 8080

ENTRYPOINT [ "finger" ]
CMD [ "serve" ]
