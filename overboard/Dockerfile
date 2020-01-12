FROM golang:1.12
WORKDIR /overboard
COPY . /overboard/
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /overboard/
COPY --from=0 /overboard/overboard .
CMD ["./overboard"]