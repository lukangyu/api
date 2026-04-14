# syntax=docker/dockerfile:1

FROM node:20-alpine AS frontend
WORKDIR /app/web
COPY web/package.json ./
RUN npm install
COPY web/ ./
RUN npm run build

FROM golang:1.25-alpine AS backend
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
COPY --from=frontend /app/static ./static
RUN go build -o /api-gateway ./cmd/server

FROM alpine:3.20
RUN apk --no-cache add ca-certificates tzdata wget curl
WORKDIR /app
COPY --from=backend /api-gateway ./api-gateway
EXPOSE 8080
VOLUME ["/app/data"]
ENV DB_PATH=/app/data/gateway.db
CMD ["./api-gateway"]
