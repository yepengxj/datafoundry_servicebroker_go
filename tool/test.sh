curl -X GET http://asiainfoLDP:2016asia@localhost:8000/v2/catalog


#########################生成实例######################
##错误
curl -X PUT http://asiainfoLDP:2016asia@localhost:8000/v2/service_instances/instance_guid-111 -d '{
  "service_id":"service-guid-111",
  "plan_id":"plan-guid",
  "organization_guid": "org-guid",
  "space_guid":"space-guid",
  "parameters": {"ami_id":"ami-ecb68a84"}
}' -H "Content-Type: application/json"
##正确
curl -i -X PUT http://asiainfoLDP:2016asia@localhost:8000/v2/service_instances/instance_guid-111 -d '{
  "service_id":"0A789746-596F-4CEA-BFAC-A0795DA056E3",
  "plan_id":"ABE176EE-F69F-4A96-80CE-142595CC24E3",
  "organization_guid": "org-guid",
  "space_guid":"space-guid",
  "parameters": {"ami_id":"ami-ecb68a84"}
}' -H "Content-Type: application/json"

###没有提供下列查询实例状态的命令###
curl -X GET http://asiainfoLDP:2016asia@localhost:8000/v2/service_instances/instance_guid-111
curl -X GET http://asiainfoLDP:2016asia@localhost:8000/v2/service_instances/instance_guid-222

curl -X PUT http://asiainfoLDP:2016asia@localhost:8000/v2/service_instances/instance_guid-111/service_bindings/binding_guid-111 -d '{
  "plan_id":        "plan-guid",
  "service_id":     "service-guid-111",
  "app_guid":       "app-guid"
}' -H "Content-Type: application/json"

curl -X PUT http://asiainfoLDP:2016asia@localhost:8000/v2/service_instances/instance_guid-222/service_bindings/binding_guid-222 -d '{
  "plan_id":        "plan-guid",
  "service_id":     "service-guid-222",
  "app_guid":       "app-guid"
}' -H "Content-Type: application/json"

curl -X DELETE http://asiainfoLDP:2016asia@localhost:8000/v2/service_instances/instance_guid-111/service_bindings/binding_guid-111
curl -X DELETE http://asiainfoLDP:2016asia@localhost:8000/v2/service_instances/instance_guid-111

curl -X DELETE http://asiainfoLDP:2016asia@localhost:8000/v2/service_instances/instance_guid-222/service_bindings/binding_guid-222
curl -X DELETE http://asiainfoLDP:2016asia@localhost:8000/v2/service_instances/instance_guid-222