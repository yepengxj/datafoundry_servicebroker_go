curl -X GET http://asiainfoLDP:2016asia@localhost:8000/v2/catalog


#########################生成实例######################
##错误
curl -i -X PUT http://asiainfoLDP:2016asia@localhost:8000/v2/service_instances/98A763D7-CE08-4E0D-B139-769F80B6DEFD -d '{
  "service_id":"service-guid-111",
  "plan_id":"plan-guid",
  "organization_guid": "org-guid",
  "space_guid":"space-guid",
  "parameters": {"ami_id":"ami-ecb68a84"}
}' -H "Content-Type: application/json"
##正确
curl -i -X PUT http://asiainfoLDP:2016asia@localhost:8000/v2/service_instances/9BECC443-7BBC-411F-AEDA-60070173DAE9 -d '{
  "service_id":"A25DE423-484E-4252-B6FE-EA4F347BCE3D",
  "plan_id":"E28FB3AE-C237-484F-AC9D-FB0A80223F85",
  "organization_guid": "org-guid",
  "space_guid":"space-guid",
  "parameters": {"ami_id":"ami-ecb68a84"}
}' -H "Content-Type: application/json"

###没有提供下列查询实例状态的命令###
curl -X GET http://asiainfoLDP:2016asia@localhost:8000/v2/service_instances/instance_guid-111
curl -X GET http://asiainfoLDP:2016asia@localhost:8000/v2/service_instances/instance_guid-222

curl -X PUT http://asiainfoLDP:2016asia@localhost:8000/v2/service_instances/98A763D7-CE08-4E0D-B139-769F80B6DEFD/service_bindings/6853A95B-428B-4C10-99FC-1BE6CBFBE176 -d '{
  "plan_id":        "plan-guid",
  "service_id":     "service-guid-111",
  "app_guid":       "app-guid"
}' -H "Content-Type: application/json"

curl -X PUT http://asiainfoLDP:2016asia@localhost:8000/v2/service_instances/instance_guid-222/service_bindings/binding_guid-222 -d '{
  "plan_id":        "plan-guid",
  "service_id":     "service-guid-222",
  "app_guid":       "app-guid"
}' -H "Content-Type: application/json"

curl -i -X DELETE http://asiainfoLDP:2016asia@localhost:8000/v2/service_instances/98A763D7-CE08-4E0D-B139-769F80B6DEFD/service_bindings/6853A95B-428B-4C10-99FC-1BE6CBFBE176
curl -i -X DELETE http://asiainfoLDP:2016asia@localhost:8000/v2/service_instances/98A763D7-CE08-4E0D-B139-769F80B6DEFD

curl -i -X DELETE http://asiainfoLDP:2016asia@localhost:8000/v2/service_instances/9BECC443-7BBC-411F-AEDA-60070173DAE9/service_bindings/binding_guid-222
curl -i -X DELETE http://asiainfoLDP:2016asia@localhost:8000/v2/service_instances/9BECC443-7BBC-411F-AEDA-60070173DAE9