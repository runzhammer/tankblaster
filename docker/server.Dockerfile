FROM golang:1.26.5-alpine AS build

ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_TIME=unknown

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -trimpath \
    -ldflags="-s -w -X github.com/runzhammer/tankblaster/pkg/buildinfo.Version=${VERSION} -X github.com/runzhammer/tankblaster/pkg/buildinfo.Commit=${COMMIT} -X github.com/runzhammer/tankblaster/pkg/buildinfo.BuildTime=${BUILD_TIME}" \
    -o /out/tankblaster-server \
    ./cmd/tankblaster-server
RUN mkdir -p /out/data && touch /out/data/.keep

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /
COPY --from=build /out/tankblaster-server /usr/local/bin/tankblaster-server
COPY docker/server.yaml /etc/tankblaster/server.yaml
COPY --from=build --chown=nonroot:nonroot /out/data /data

USER nonroot:nonroot
EXPOSE 8765
VOLUME ["/data"]

ENTRYPOINT ["/usr/local/bin/tankblaster-server"]
CMD ["-config", "/etc/tankblaster/server.yaml"]
