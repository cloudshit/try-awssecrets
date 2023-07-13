FROM golang:latest

WORKDIR /app

COPY . /app

RUN go get .

RUN go build -o main .

CMD ["/app/main"]
