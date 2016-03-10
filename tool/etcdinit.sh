#运行etcd
HostIP=`docker-machine ip default`

HostIP="54.222.175.239"

docker run -d -v /usr/share/ca-certificates/:/etc/ssl/certs -p 4001:4001 -p 2380:2380 -p 2379:2379 \
 --name etcd quay.io/coreos/etcd:v2.2.5 \
 -name etcd0 \
 -advertise-client-urls http://${HostIP}:2379,http://${HostIP}:4001 \
 -listen-client-urls http://0.0.0.0:2379,http://0.0.0.0:4001 \
 -initial-advertise-peer-urls http://${HostIP}:2380 \
 -listen-peer-urls http://0.0.0.0:2380 \
 -initial-cluster-token etcd-cluster-1 \
 -initial-cluster etcd0=http://${HostIP}:2380 \
 -initial-cluster-state new

#访问容器
docker exec etcd /etcdctl ls /

#测试端口
curl $HostIP:2379

#建立初始化数据
##!!注意，初始化的时候，所有键值必须小写，这样程序才认识
##初始化用户名和密码部分
docker exec etcd /etcdctl mkdir /servicebroker
docker exec etcd /etcdctl mkdir /servicebroker/mongodb_aws
docker exec etcd /etcdctl set /servicebroker/mongodb_aws/username asiainfoLDP
docker exec etcd /etcdctl set /servicebroker/mongodb_aws/password 2016asia

##初始化catalog
docker exec etcd /etcdctl mkdir /servicebroker/mongodb_aws/catalog

###创建服务
docker exec etcd /etcdctl mkdir /servicebroker/mongodb_aws/catalog/A25DE423-484E-4252-B6FE-EA4F347BCE3D #服务id

###创建服务级的配置
docker exec etcd /etcdctl set /servicebroker/mongodb_aws/catalog/A25DE423-484E-4252-B6FE-EA4F347BCE3D/name "mongodb_aws"
docker exec etcd /etcdctl set /servicebroker/mongodb_aws/catalog/A25DE423-484E-4252-B6FE-EA4F347BCE3D/description "A MongoDB for AWS"
docker exec etcd /etcdctl set /servicebroker/mongodb_aws/catalog/A25DE423-484E-4252-B6FE-EA4F347BCE3D/bindable true
docker exec etcd /etcdctl set /servicebroker/mongodb_aws/catalog/A25DE423-484E-4252-B6FE-EA4F347BCE3D/planupdatable false
docker exec etcd /etcdctl set /servicebroker/mongodb_aws/catalog/A25DE423-484E-4252-B6FE-EA4F347BCE3D/tags 'amqp,rabbitmq,messaging'
docker exec etcd /etcdctl set /servicebroker/mongodb_aws/catalog/A25DE423-484E-4252-B6FE-EA4F347BCE3D/metadata '{"displayName":"CloudAMQP","imageUrl":"https://d33na3ni6eqf5j.cloudfront.net/app_resources/18492/thumbs_112/img9069612145282015279.png","longDescription":"Managed, highly available, RabbitMQ clusters in the cloud","providerDisplayName":"84codes AB","documentationUrl":"http://docs.cloudfoundry.com/docs/dotcom/marketplace/services/cloudamqp.html","supportUrl":"http://www.cloudamqp.com/support.html"}'

###创建套餐目录
docker exec etcd /etcdctl mkdir /servicebroker/mongodb_aws/catalog/A25DE423-484E-4252-B6FE-EA4F347BCE3D/plan
###创建套餐1
docker exec etcd /etcdctl mkdir /servicebroker/mongodb_aws/catalog/A25DE423-484E-4252-B6FE-EA4F347BCE3D/plan/E28FB3AE-C237-484F-AC9D-FB0A80223F85
docker exec etcd /etcdctl set /servicebroker/mongodb_aws/catalog/A25DE423-484E-4252-B6FE-EA4F347BCE3D/plan/E28FB3AE-C237-484F-AC9D-FB0A80223F85/name "shared"
docker exec etcd /etcdctl set /servicebroker/mongodb_aws/catalog/A25DE423-484E-4252-B6FE-EA4F347BCE3D/plan/E28FB3AE-C237-484F-AC9D-FB0A80223F85/description "share a mongodb instance on aws"
docker exec etcd /etcdctl set /servicebroker/mongodb_aws/catalog/A25DE423-484E-4252-B6FE-EA4F347BCE3D/plan/E28FB3AE-C237-484F-AC9D-FB0A80223F85/metadata '{"bullets":["20 GB of Disk","20 connections"],"displayName":"Shared and Free" }'
docker exec etcd /etcdctl set /servicebroker/mongodb_aws/catalog/A25DE423-484E-4252-B6FE-EA4F347BCE3D/plan/E28FB3AE-C237-484F-AC9D-FB0A80223F85/free true

###创建套餐2
docker exec etcd /etcdctl mkdir /servicebroker/mongodb_aws/catalog/A25DE423-484E-4252-B6FE-EA4F347BCE3D/plan/8C7E1AB9-DB63-4E14-9487-733BB587E1B2
docker exec etcd /etcdctl set /servicebroker/mongodb_aws/catalog/A25DE423-484E-4252-B6FE-EA4F347BCE3D/plan/8C7E1AB9-DB63-4E14-9487-733BB587E1B2/name "standalone"
docker exec etcd /etcdctl set /servicebroker/mongodb_aws/catalog/A25DE423-484E-4252-B6FE-EA4F347BCE3D/plan/8C7E1AB9-DB63-4E14-9487-733BB587E1B2/description "each user has a standalone mongodb instance on aws"
docker exec etcd /etcdctl set /servicebroker/mongodb_aws/catalog/A25DE423-484E-4252-B6FE-EA4F347BCE3D/plan/8C7E1AB9-DB63-4E14-9487-733BB587E1B2/metadata '{"bullets":["20 GB of Disk","20 connections"],"costs":[{"amount":{"usd":99.0,"eur":49.0},"unit":"MONTHLY"},{"amount":{"usd":0.99, "eur":0.49}, "unit":"1GB of messages over 20GB"} ], "displayName":"Big Bunny" }'
docker exec etcd /etcdctl set /servicebroker/mongodb_aws/catalog/A25DE423-484E-4252-B6FE-EA4F347BCE3D/plan/8C7E1AB9-DB63-4E14-9487-733BB587E1B2/free false

####创建套餐3 share and read common
docker exec etcd /etcdctl -u root:asiainfoLDP mkdir /servicebroker/mongodb_aws/catalog/A25DE423-484E-4252-B6FE-EA4F347BCE3D/plan/257C6C2B-A376-4551-90E8-82D4E619C852
docker exec etcd /etcdctl -u root:asiainfoLDP set /servicebroker/mongodb_aws/catalog/A25DE423-484E-4252-B6FE-EA4F347BCE3D/plan/257C6C2B-A376-4551-90E8-82D4E619C852/name "shareandcommon"
docker exec etcd /etcdctl -u root:asiainfoLDP set /servicebroker/mongodb_aws/catalog/A25DE423-484E-4252-B6FE-EA4F347BCE3D/plan/257C6C2B-A376-4551-90E8-82D4E619C852/description "share a mongodb instance on aws,but can select from database aqi_demo"
docker exec etcd /etcdctl -u root:asiainfoLDP set /servicebroker/mongodb_aws/catalog/A25DE423-484E-4252-B6FE-EA4F347BCE3D/plan/257C6C2B-A376-4551-90E8-82D4E619C852/metadata '{"bullets":["20 GB of Disk","20 connections"],"costs":[{"amount":{"usd":99.0,"eur":49.0},"unit":"MONTHLY"},{"amount":{"usd":0.99, "eur":0.49}, "unit":"1GB of messages over 20GB"} ], "displayName":"Big Bunny" }'
docker exec etcd /etcdctl -u root:asiainfoLDP set /servicebroker/mongodb_aws/catalog/A25DE423-484E-4252-B6FE-EA4F347BCE3D/plan/257C6C2B-A376-4551-90E8-82D4E619C852/free false

----创建服务2 mysql
###创建服务
docker exec etcd /etcdctl -u root:asiainfoLDP mkdir /servicebroker/mongodb_aws/catalog/7D2AB7B3-8AEF-45EE-BFF2-64A767DDE9DA #服务id

###创建服务级的配置
docker exec etcd /etcdctl -u root:asiainfoLDP set /servicebroker/mongodb_aws/catalog/7D2AB7B3-8AEF-45EE-BFF2-64A767DDE9DA/name "mysql"
docker exec etcd /etcdctl -u root:asiainfoLDP set /servicebroker/mongodb_aws/catalog/7D2AB7B3-8AEF-45EE-BFF2-64A767DDE9DA/description "A MYSQL Service"
docker exec etcd /etcdctl -u root:asiainfoLDP set /servicebroker/mongodb_aws/catalog/7D2AB7B3-8AEF-45EE-BFF2-64A767DDE9DA/bindable true
docker exec etcd /etcdctl -u root:asiainfoLDP set /servicebroker/mongodb_aws/catalog/7D2AB7B3-8AEF-45EE-BFF2-64A767DDE9DA/planupdatable false
docker exec etcd /etcdctl -u root:asiainfoLDP set /servicebroker/mongodb_aws/catalog/7D2AB7B3-8AEF-45EE-BFF2-64A767DDE9DA/tags 'mysql,database'
docker exec etcd /etcdctl -u root:asiainfoLDP set /servicebroker/mongodb_aws/catalog/7D2AB7B3-8AEF-45EE-BFF2-64A767DDE9DA/metadata '{"displayName":"Mysql","imageUrl":"https://d33na3ni6eqf5j.cloudfront.net/app_resources/18492/thumbs_112/img9069612145282015279.png","longDescription":"Managed, highly available, RabbitMQ clusters in the cloud","providerDisplayName":"84codes AB","documentationUrl":"http://docs.cloudfoundry.com/docs/dotcom/marketplace/services/cloudamqp.html","supportUrl":"http://www.cloudamqp.com/support.html"}'

###创建套餐目录
docker exec etcd /etcdctl -u root:asiainfoLDP mkdir /servicebroker/mongodb_aws/catalog/7D2AB7B3-8AEF-45EE-BFF2-64A767DDE9DA/plan
###创建套餐1
docker exec etcd /etcdctl -u root:asiainfoLDP mkdir /servicebroker/mongodb_aws/catalog/7D2AB7B3-8AEF-45EE-BFF2-64A767DDE9DA/plan/56660431-6032-43D0-A114-FFA3BF521B66
docker exec etcd /etcdctl -u root:asiainfoLDP set /servicebroker/mongodb_aws/catalog/7D2AB7B3-8AEF-45EE-BFF2-64A767DDE9DA/plan/56660431-6032-43D0-A114-FFA3BF521B66/name "shared"
docker exec etcd /etcdctl -u root:asiainfoLDP set /servicebroker/mongodb_aws/catalog/7D2AB7B3-8AEF-45EE-BFF2-64A767DDE9DA/plan/56660431-6032-43D0-A114-FFA3BF521B66/description "share a mysql instance on aws"
docker exec etcd /etcdctl -u root:asiainfoLDP set /servicebroker/mongodb_aws/catalog/7D2AB7B3-8AEF-45EE-BFF2-64A767DDE9DA/plan/56660431-6032-43D0-A114-FFA3BF521B66/metadata '{"bullets":["20 GB of Disk","20 connections"],"displayName":"Shared and Free" }'
docker exec etcd /etcdctl -u root:asiainfoLDP set /servicebroker/mongodb_aws/catalog/7D2AB7B3-8AEF-45EE-BFF2-64A767DDE9DA/plan/56660431-6032-43D0-A114-FFA3BF521B66/free true

----创建服务3 postgresql
###创建服务
docker exec etcd /etcdctl -u root:asiainfoLDP mkdir /servicebroker/mongodb_aws/catalog/cb2d4021-5fbc-45c2-92a9-9584583b7ce5 #服务id

###创建服务级的配置
docker exec etcd /etcdctl -u root:asiainfoLDP set /servicebroker/mongodb_aws/catalog/cb2d4021-5fbc-45c2-92a9-9584583b7ce5/name "postgresql"
docker exec etcd /etcdctl -u root:asiainfoLDP set /servicebroker/mongodb_aws/catalog/cb2d4021-5fbc-45c2-92a9-9584583b7ce5/description "A Postgresql Service"
docker exec etcd /etcdctl -u root:asiainfoLDP set /servicebroker/mongodb_aws/catalog/cb2d4021-5fbc-45c2-92a9-9584583b7ce5/bindable true
docker exec etcd /etcdctl -u root:asiainfoLDP set /servicebroker/mongodb_aws/catalog/cb2d4021-5fbc-45c2-92a9-9584583b7ce5/planupdatable false
docker exec etcd /etcdctl -u root:asiainfoLDP set /servicebroker/mongodb_aws/catalog/cb2d4021-5fbc-45c2-92a9-9584583b7ce5/tags 'postgresql,database,experiment'
docker exec etcd /etcdctl -u root:asiainfoLDP set /servicebroker/mongodb_aws/catalog/cb2d4021-5fbc-45c2-92a9-9584583b7ce5/metadata '{"displayName":"Postgresql","imageUrl":"https://d33na3ni6eqf5j.cloudfront.net/app_resources/18492/thumbs_112/img9069612145282015279.png","longDescription":"Managed, highly available, RabbitMQ clusters in the cloud","providerDisplayName":"84codes AB","documentationUrl":"http://docs.cloudfoundry.com/docs/dotcom/marketplace/services/cloudamqp.html","supportUrl":"http://www.cloudamqp.com/support.html"}'

###创建套餐目录
docker exec etcd /etcdctl -u root:asiainfoLDP mkdir /servicebroker/mongodb_aws/catalog/cb2d4021-5fbc-45c2-92a9-9584583b7ce5/plan
###创建套餐1
docker exec etcd /etcdctl -u root:asiainfoLDP mkdir /servicebroker/mongodb_aws/catalog/cb2d4021-5fbc-45c2-92a9-9584583b7ce5/plan/bd9a94f2-5718-4dde-a773-61ff4ad9e843
docker exec etcd /etcdctl -u root:asiainfoLDP set /servicebroker/mongodb_aws/catalog/cb2d4021-5fbc-45c2-92a9-9584583b7ce5/plan/bd9a94f2-5718-4dde-a773-61ff4ad9e843/name "shared"
docker exec etcd /etcdctl -u root:asiainfoLDP set /servicebroker/mongodb_aws/catalog/cb2d4021-5fbc-45c2-92a9-9584583b7ce5/plan/bd9a94f2-5718-4dde-a773-61ff4ad9e843/description "share a postgresql instance on aws"
docker exec etcd /etcdctl -u root:asiainfoLDP set /servicebroker/mongodb_aws/catalog/cb2d4021-5fbc-45c2-92a9-9584583b7ce5/plan/bd9a94f2-5718-4dde-a773-61ff4ad9e843/metadata '{"bullets":["20 GB of Disk","20 connections"],"displayName":"Shared and Free" }'
docker exec etcd /etcdctl -u root:asiainfoLDP set /servicebroker/mongodb_aws/catalog/cb2d4021-5fbc-45c2-92a9-9584583b7ce5/plan/bd9a94f2-5718-4dde-a773-61ff4ad9e843/free true


##初始化instance
docker exec etcd /etcdctl mkdir /servicebroker/mongodb_aws/instance

