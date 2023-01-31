FROM golang:alpine as build

ENV CGO_ENABLED=0
ENV GOOS=linux

WORKDIR /app
COPY . .
RUN go build -o webserver

FROM alpine
EXPOSE 8080

COPY migrations /migrations
COPY --from=build /app/webserver /

CMD /webserver
