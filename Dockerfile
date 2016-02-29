FROM golang:1.5.2

ENV BROKERPORT 8000
EXPOSE 8000

ENV TIME_ZONE=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TIME_ZONE /etc/localtime && echo $TIME_ZONE > /etc/timezone

RUN go get github.com/asiainfoLDP/datafoundry_servicebroker_mongodb_aws

CMD ["sh", "-c", "datafoundry_servicebroker_mongodb_aws"]