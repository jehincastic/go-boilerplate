FROM golang:1.27.1-alpine AS build

WORKDIR /src
RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

FROM alpine:3.22

RUN apk add --no-cache ca-certificates wget \
  && adduser -D -H -u 65532 app

COPY --from=build /out/api /api

USER app
EXPOSE 4000

ENTRYPOINT ["/api"]
