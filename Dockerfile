FROM golang:latest
WORKDIR /alice
COPY ./api/ /alice/

RUN go build
ENTRYPOINT ["./alice"]
