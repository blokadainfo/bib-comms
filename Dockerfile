FROM golang:1.25 AS build
WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY . .
RUN apt update && \
    apt install -y \
    libopenal-dev \
    libopus-dev \
    libasound2-dev \
    pkg-config
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -trimpath -o bib-comms .

FROM debian:trixie-slim
WORKDIR /app
ENV DEBIAN_FRONTEND=noninteractive
RUN apt update && \
    apt install -y \
    ca-certificates \
    libopenal-dev \
    libopus-dev \
    libasound2-dev \
    pkg-config \
    ffmpeg
RUN rm -rf /var/lib/apt/lists/*

COPY --from=build /app/bib-comms /app/bib-comms
ENTRYPOINT ["./bib-comms"]
