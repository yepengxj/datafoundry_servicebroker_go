#运行etcd
HostIP=`docker-machine ip default`


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
##初始化用户名和密码部分
docker exec etcd /etcdctl mkdir /servicebroker
docker exec etcd /etcdctl mkdir /servicebroker/mongodb_aws
docker exec etcd /etcdctl set /servicebroker/mongodb_aws/username asiainfoLDP
docker exec etcd /etcdctl set /servicebroker/mongodb_aws/password 2016asia

##初始化catalog
docker exec etcd /etcdctl mkdir /servicebroker/mongodb_aws/catalog


##初始化instance
docker exec etcd /etcdctl mkdir /servicebroker/mongodb_aws/instance

