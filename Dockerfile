FROM --platform=$BUILDPLATFORM golang:1.25 AS builder
ARG TARGETARCH
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH go build -o /kurokku ./cmd/kurokku
RUN CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH go build -o /kurokku-admin ./cmd/kurokku-admin

FROM debian:bookworm-slim AS kurokku
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates tzdata && rm -rf /var/lib/apt/lists/*
COPY --from=builder /kurokku /usr/local/bin/kurokku
ENTRYPOINT ["kurokku"]

FROM debian:bookworm-slim AS kurokku-admin
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates tzdata && rm -rf /var/lib/apt/lists/*
COPY --from=builder /kurokku-admin /usr/local/bin/kurokku-admin
EXPOSE 8080
ENTRYPOINT ["kurokku-admin"]
