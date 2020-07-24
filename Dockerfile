FROM golang:1.14

WORKDIR /go/src/app
COPY . .
COPY ./files /files
COPY ./config.json /config.json


RUN go get -d -v ./...
RUN go install -v ./...

CMD ["app"]