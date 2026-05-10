FROM node:22-alpine AS ui
WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM golang:1.25 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=ui /app/web/dist ./web/dist
RUN CGO_ENABLED=1 go build -trimpath -o /bin/server ./cmd/server

FROM debian:bookworm-slim
RUN apt-get update \
 && apt-get install -y --no-install-recommends ca-certificates tzdata \
 && rm -rf /var/lib/apt/lists/*
COPY --from=builder /bin/server /server
EXPOSE 8080
ENTRYPOINT ["/server"]
