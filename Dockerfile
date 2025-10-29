FROM golang:1.25.3-bookworm AS builder

WORKDIR /app

RUN apt-get update && apt-get install -y git gcc unzip curl zlib1g-dev && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod tidy

COPY setup_ntgcalls.sh ./
COPY . .

RUN chmod +x setup_ntgcalls.sh
RUN ./setup_ntgcalls.sh
RUN CGO_ENABLED=1 go build -trimpath -ldflags="-w -s" -o myapp ./cmd/app/

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y ffmpeg wget zlib1g && \
    wget -O /usr/local/bin/yt-dlp https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_linux && \
    chmod +x /usr/local/bin/yt-dlp && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /app/myapp /app/
ENTRYPOINT ["/app/myapp"]
