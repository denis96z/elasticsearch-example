FROM golang:1.15 as base
WORKDIR /root/src/app

COPY . .
RUN go build -o ./app ./main.go

FROM centos:7
WORKDIR /usr/local/bin

COPY --from=base /root/src/app/app ./app
CMD ["./app", "-p", "8080", "--esh", "127.0.0.1", "--esp", "9200"]
