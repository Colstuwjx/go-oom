FROM golang:1.8
MAINTAINER colstuwjx@gmail.com

WORKDIR /
COPY go-oom.linux-amd64.bin /go-oom
RUN chmod +x /go-oom

CMD /go-oom
