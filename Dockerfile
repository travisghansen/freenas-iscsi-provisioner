FROM golang:alpine as build
RUN apk add --no-cache git
WORKDIR /src
COPY . .
RUN go get -d -v .
RUN go build -v -o app .

FROM alpine
RUN apk add --no-cache ca-certificates
WORKDIR /svc
COPY --from=build /src/app .
ENTRYPOINT ["./app"]
