FROM golang:1.5.2

ENV BROKERPORT 8000
EXPOSE 8000

ENV TIME_ZONE=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TIME_ZONE /etc/localtime && echo $TIME_ZONE > /etc/timezone


COPY . /usr/src/servicebroker

WORKDIR /usr/src/servicebroker

RUN go get github.com/tools/godep \
    && godep go build 

CMD ["sh", "-c", "./servicebroker"]