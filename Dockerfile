FROM golang:latest as builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o plantopo-api .

FROM ubuntu:latest
LABEL org.opencontainers.image.source=https://github.com/dzfranklin/plantopo-api

WORKDIR /app

RUN apt-get update && apt-get install -y ca-certificates

COPY --from=builder /build/plantopo-api /app/plantopo-api

EXPOSE 8000

ENTRYPOINT ["/app/plantopo-api"]
