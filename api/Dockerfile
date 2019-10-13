FROM golang:1.12
WORKDIR /alice
COPY . /alice/

RUN go build
ENTRYPOINT ["./alice"]
