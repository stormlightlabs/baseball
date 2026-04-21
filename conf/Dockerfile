# syntax=docker/dockerfile:1

FROM golang:1.24 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /baseball ./cmd/baseball && \
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /baseball-etl ./cmd/baseball-etl

FROM debian:bookworm-slim

RUN adduser --disabled-password --gecos "" app && \
	apt-get update && apt-get install -y --no-install-recommends ca-certificates git git-lfs && \
	git lfs install --system --skip-smudge && \
	rm -rf /var/lib/apt/lists/*

WORKDIR /home/app
COPY --from=build /baseball /usr/local/bin/baseball
COPY --from=build /baseball-etl /usr/local/bin/baseball-etl
COPY --from=build /app/data /home/app/data
RUN mkdir -p /home/app/data && chown -R app:app /home/app/data

USER app
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=20s --retries=3 CMD ["/usr/local/bin/baseball", "server", "health"]

ENV BASEBALL_DATA_ROOT=/home/app/data

ENTRYPOINT ["/usr/local/bin/baseball"]
CMD ["server", "start"]
