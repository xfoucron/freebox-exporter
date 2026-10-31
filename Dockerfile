FROM --platform=$BUILDPLATFORM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG TARGETOS TARGETARCH VERSION=dev
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags "-s -w -X main.version=$VERSION" -o /out/freebox-exporter ./cmd/freebox-exporter
RUN mkdir /data

FROM gcr.io/distroless/static:nonroot
COPY --from=build /out/freebox-exporter /freebox-exporter
COPY --from=build --chown=65532:65532 /data /data
ENV FREEBOX_DATA_DIR=/data
VOLUME /data
EXPOSE 9091
ENTRYPOINT ["/freebox-exporter"]
