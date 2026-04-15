FROM node:22-bookworm AS frontend-builder

WORKDIR /workspace/frontend

COPY frontend/package*.json ./
RUN if [ -f package-lock.json ]; then npm ci; else npm install; fi

COPY frontend/ ./
RUN npm run build


FROM golang:1.25-bookworm AS backend-builder

RUN apt-get update \
  && apt-get install -y --no-install-recommends ca-certificates build-essential \
  && rm -rf /var/lib/apt/lists/*

WORKDIR /workspace/backend

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./
COPY --from=frontend-builder /workspace/frontend/dist ./web/dist

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o /out/openshare ./cmd/server
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o /out/openshare-worker ./cmd/worker


FROM debian:bookworm-slim AS runtime

RUN apt-get update \
  && apt-get install -y --no-install-recommends ca-certificates \
  && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=backend-builder /out/openshare ./openshare
COPY --from=backend-builder /out/openshare-worker ./openshare-worker
COPY --from=backend-builder /workspace/backend/config ./config

RUN mkdir -p /data/openshare

ENV OPENSHARE_SERVER_HOST=0.0.0.0 \
  OPENSHARE_SERVER_PORT=8080 \
  OPENSHARE_DATABASE_PATH=/data/openshare/openshare.db \
  OPENSHARE_STORAGE_ROOT=/data/openshare

VOLUME ["/data/openshare"]

EXPOSE 8080

CMD ["./openshare"]
