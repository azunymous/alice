FROM golang:1.12
WORKDIR /alice
COPY . /alice/
RUN go build

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /alice/
COPY --from=0 /alice/alice .
CMD ["./alice"]