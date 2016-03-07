docker build -t mongodb_aws .

docker run -d -p 8000:8000 \
	-e "ETCDENDPOINT=http://192.168.99.100:2379"  \
	-e "BROKERPORT=8000"	 \
	-e "MONGOURL=54.222.155.67:27017"	 \
	-e "MONGOADMINUSER=asiainfoLDP"  \
	-e "MONGOADMINPASSWORD=6ED9BA74-75FD-4D1B-8916-842CB936AC1A"  \
	-e "AWS_ACCESS_KEY_ID=XXX"
	-e "AWS_SECRET_ACCESS_KEY=XXX"
	--name mongodb_aws mongodb_aws
