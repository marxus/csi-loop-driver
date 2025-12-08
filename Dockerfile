#syntax=docker/dockerfile:1.7-labs
ARG BUILDSTAGE=build

# Stage for CI pre-built binaries
FROM scratch AS prebuilt
ARG TARGETARCH
WORKDIR /app
COPY csi-loop-driver-${TARGETARCH} ./csi-loop-driver

# Stage for local builds
FROM golang:1.24-alpine3.22 AS build
WORKDIR /app
COPY --parents go.mod go.sum cmd pkg ./
RUN go mod download
RUN CGO_ENABLED=0 go build -o csi-loop-driver cmd/csi-loop-driver/main.go

FROM ${BUILDSTAGE} AS binary
FROM alpine:3.22
WORKDIR /app
COPY --from=binary /app/csi-loop-driver ./csi-loop-driver
RUN apk add --no-cache util-linux btrfs-progs
ENTRYPOINT ["./csi-loop-driver"]
