FROM golang:1.25.5-alpine AS builder
RUN mkdir /build && \
    apk add --no-cache build-base sqlite-dev
WORKDIR /build
COPY . .
ARG CGO_ENABLED=1
RUN go mod tidy && \
    go build -o policy-backend

FROM alpine:latest
COPY --from=builder /build/policy-backend /usr/local/bin/policy-backend
RUN chmod +x /usr/local/bin/policy-backend
RUN mkdir -p /app/config
ENV DATABASE_URL="sqlite3:///app/policy.db" \
    SERVER_ADDRESS=":8080" \
    JWT_TOKEN_DURATION=24
ENV JWT_SECRET_KEY=""
COPY config.yaml /app/config/
WORKDIR /app
EXPOSE 8080
CMD ["/usr/local/bin/policy-backend"]