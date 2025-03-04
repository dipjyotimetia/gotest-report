FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY . .
RUN go build -o gotest-report .

FROM alpine:latest

LABEL org.opencontainers.image.title="gotest-report"
LABEL org.opencontainers.image.description="A tool to generate markdown reports from Go test JSON output"
LABEL org.opencontainers.image.source="https://github.com/dipjyotimetia/gotest-report"
LABEL org.opencontainers.image.licenses="MIT"
LABEL maintainer="Dipjyoti Metia"

RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/gotest-report /usr/local/bin/

ENTRYPOINT ["gotest-report"]