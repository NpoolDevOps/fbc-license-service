FROM golang:latest

ENV GO111MODULE on
ENV GOPROXY https://goproxy.cn,direct
WORKDIR /app/guard_server/

ARG NAME
ARG VERSION

LABEL name=$NAME \
      version=$VERSION


COPY ./* /app/guard_server/

EXPOSE 5000 

Run go build main.go

CMD ["./main"]

