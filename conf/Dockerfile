# syntax=docker/dockerfile:1

FROM golang:1.24 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /baseball ./cli

FROM debian:bookworm-slim

RUN adduser --disabled-password --gecos "" app && \
	apt-get update && apt-get install -y --no-install-recommends ca-certificates && \
	rm -rf /var/lib/apt/lists/*

WORKDIR /home/app
COPY --from=build /baseball /usr/local/bin/baseball

USER app
EXPOSE 8080

ENV DATABASE_URL=postgres://postgres:postgres@postgres:5432/baseball_dev?sslmode=disable

ENTRYPOINT ["/usr/local/bin/baseball"]
CMD ["server", "start"]
