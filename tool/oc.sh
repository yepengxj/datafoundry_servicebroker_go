oc login https://lab.asiainfodata.com:8443  -u hehl@asiainfo.com  -p 1857e645-c263-407d-b4e1-e82d15df8d6e --insecure-skip-tls-verify=true

oc new-app XXXXXXXX

oc new-app --docker-image=index.alauda.cn/asiainfoldp/datafoundry_servicebroker_mongo \
	--name=servicebroker-mongo \
    -e  ETCDENDPOINT="http://54.222.175.239:2379"  \
    -e  BROKERPORT="8000"  \
    -e  MONGOURL="54.222.175.239:27017"  \
    -e  MONGOADMINUSER="asiainfoLDP"   \
    -e  MONGOADMINPASSWORD="6ED9BA74-75FD-4D1B-8916-842CB936AC1A"   \
    -e  AWS_ACCESS_KEY_ID=AKIAO2SO52RKIE7BCSHA  \
    -e  AWS_SECRET_ACCESS_KEY=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

oc expose  svc servicebroker-mongo