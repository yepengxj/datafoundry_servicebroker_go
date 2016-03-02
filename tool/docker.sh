docker build -t mongodb_aws .

docker run -d -p 8000:8000 \
	-e "ETCDENDPOINT=http://192.168.99.100:2379"  \
	-e "BROKERPORT=8000"	 \
	-e "MONGOURL=54.222.155.67:27017"	 \
	-e "MONGOADMINUSER=asiainfoLDP"  \
	-e "MONGOADMINPASSWORD=6ED9BA74-75FD-4D1B-8916-842CB936AC1A"  \
	--name mongodb_aws mongodb_aws


	//aws客户端还需要额外两个环境变量
	export AWS_ACCESS_KEY_ID=AKIAO2SO52RKIE7BCSHA
	export AWS_SECRET_ACCESS_KEY=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