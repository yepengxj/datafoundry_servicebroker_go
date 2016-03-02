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
##初始化instance
docker exec etcd /etcdctl mkdir /servicebroker/mongodb_aws/instance

###创建一个用于测试的服务实例
docker exec etcd /etcdctl mkdir /servicebroker/mongodb_aws/instance/98A763D7-CE08-4E0D-B139-769F80B6DEFD

###创建一个用于测试的绑定实例
docker exec etcd /etcdctl mkdir /servicebroker/mongodb_aws/instance/98A763D7-CE08-4E0D-B139-769F80B6DEFD/binding
docker exec etcd /etcdctl mkdir /servicebroker/mongodb_aws/instance/98A763D7-CE08-4E0D-B139-769F80B6DEFD/binding/6853A95B-428B-4C10-99FC-1BE6CBFBE176


