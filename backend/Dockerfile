FROM golang:alpine as build

ENV CGO_ENABLED=0
ENV GOOS=linux

WORKDIR /app
COPY . .
RUN go build -o clipable

FROM ghcr.io/jrottenberg/ffmpeg:5-alpine
EXPOSE 8080

RUN apk add --update nodejs npm
RUN npm --silent install --global --depth 0 pnpm

COPY migrations /migrations
COPY --from=build /app/clipable /

ENTRYPOINT /clipable
