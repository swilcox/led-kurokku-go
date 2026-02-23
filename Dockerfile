FROM golang:1.25 AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /kurokku ./cmd/kurokku

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates tzdata && rm -rf /var/lib/apt/lists/*
COPY --from=builder /kurokku /usr/local/bin/kurokku
COPY config.json /etc/kurokku/config.json
ENTRYPOINT ["kurokku"]
CMD ["-display", "terminal", "-config", "/etc/kurokku/config.json"]
