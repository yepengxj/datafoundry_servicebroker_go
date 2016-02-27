#初始化mongodb
docker run --name mongo -d -P mongo --auth

docker exec -it mongo mongo admin
db.createUser({ user: 'asiainfoLDP', pwd: '6ED9BA74-75FD-4D1B-8916-842CB936AC1A', roles: [ { role: "root", db: "admin" } ] });
db.auth("asiainfoLDP","6ED9BA74-75FD-4D1B-8916-842CB936AC1A");
db.system.users.find();
show dbs 