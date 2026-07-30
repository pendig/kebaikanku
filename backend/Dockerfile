FROM golang:1.26.2-alpine AS build

RUN apk add --no-cache build-base
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -trimpath -ldflags="-s -w" -o /api ./cmd/api
RUN CGO_ENABLED=1 go build -trimpath -ldflags="-s -w" -o /bootstrap ./cmd/bootstrap

FROM alpine:3.21

RUN apk add --no-cache ca-certificates \
	&& addgroup -S app \
	&& adduser -S -G app app
COPY --from=build /api /usr/local/bin/api
COPY --from=build /bootstrap /usr/local/bin/bootstrap
USER app
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 CMD wget -qO- http://127.0.0.1:8080/readyz >/dev/null || exit 1
ENTRYPOINT ["/usr/local/bin/api"]
