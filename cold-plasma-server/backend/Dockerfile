FROM golang:1.25.7-alpine AS build

WORKDIR /src

RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o /out/api ./cmd/api

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata vips-tools libheif && adduser -D -H -s /sbin/nologin app

WORKDIR /app

COPY --from=build /out/api /app/api
COPY assets /app/assets
RUN mkdir -p /app/uploads && chown -R app:app /app/uploads

ENV PORT=8080
EXPOSE 8080

USER app

CMD ["./api"]
