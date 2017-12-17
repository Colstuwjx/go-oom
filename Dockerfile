FROM golang:1.8
MAINTAINER colstuwjx@gmail.com

WORKDIR /
COPY main.go /main.go
RUN go build -o /go-oom /main.go \
  && chmod +x /go-oom

CMD /go-oom
