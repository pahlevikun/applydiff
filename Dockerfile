# syntax=docker/dockerfile:1

FROM node:20-bookworm AS web
WORKDIR /src
COPY web ./web
RUN cd web && npm ci && npm run generate

FROM golang:1.23-bookworm AS gobuild
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=web /src/web/.output/public ./internal/webembed/public
RUN CGO_ENABLED=0 go build -o /applydiff ./cmd/applydiff

FROM debian:bookworm-slim AS runtime
RUN apt-get update && apt-get install -y --no-install-recommends \
    texlive-latex-extra \
    texlive-xetex \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*
RUN useradd -m -u 1000 applydiff
USER applydiff
WORKDIR /home/applydiff
COPY --from=gobuild /applydiff /usr/local/bin/applydiff
ENV DATABASE_PATH=/data/applydiff.db
ENV PORT=3000
EXPOSE 3000
VOLUME ["/data"]
ENTRYPOINT ["applydiff"]
CMD ["serve"]
