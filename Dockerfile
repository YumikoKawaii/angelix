FROM node:22-alpine AS ui
WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=ui /app/web/dist ./web/dist
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /bin/server ./cmd/server

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /bin/server /server
EXPOSE 8080
ENTRYPOINT ["/server"]
