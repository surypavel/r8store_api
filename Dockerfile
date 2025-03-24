FROM golang:1.23.3

WORKDIR /rossum-store

COPY rossum-store .

RUN go build -o main main.go

CMD ["./main"]