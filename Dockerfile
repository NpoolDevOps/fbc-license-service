FROM golang:latest

ENV GO111MODULE on
ENV GOPROXY https://goproxy.cn,direct
WORKDIR /app/fbc-license-service/

ARG NAME
ARG VERSION

LABEL name=$NAME \
      version=$VERSION


COPY ./ /app/fbc-license-service/

EXPOSE 5000

Run go build /app/fbc-license-service/main.go

CMD ["/app/fbc-service-service/main"]

