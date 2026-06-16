FROM golang:1.25-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/hachisocial ./cmd/app

FROM alpine:3.22

RUN addgroup -S hachi && adduser -S hachi -G hachi
WORKDIR /app
COPY --from=build /out/hachisocial /usr/local/bin/hachisocial
RUN mkdir -p /app/out/logs && chown -R hachi:hachi /app

USER hachi
EXPOSE 8080
ENTRYPOINT ["hachisocial"]
