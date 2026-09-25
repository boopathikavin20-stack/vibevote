FROM node:22-alpine AS frontend-build

WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

FROM golang:1.25-alpine AS backend-build

WORKDIR /app/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 go build -o /app/server ./cmd/server

FROM alpine:3.22

RUN adduser -D -H appuser
WORKDIR /app
COPY --from=backend-build /app/server ./server
COPY --from=frontend-build /app/frontend/dist ./frontend/dist

ENV FRONTEND_DIST=/app/frontend/dist
USER appuser
EXPOSE 8080
ENTRYPOINT ["/app/server"]