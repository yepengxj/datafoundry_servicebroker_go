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

#创建standalone aws的
curl -i -X PUT http://asiainfoLDP:2016asia@localhost:8000/v2/service_instances/4A0CA076-D00E-4945-BC49-2FD65ECAD2D1 -d '{
  "service_id":"A25DE423-484E-4252-B6FE-EA4F347BCE3D",
  "plan_id":"8C7E1AB9-DB63-4E14-9487-733BB587E1B2",
  "organization_guid": "org-guid",
  "space_guid":"space-guid",
  "parameters": {"ami_id":"ami-ecb68a84"}
}' -H "Content-Type: application/json"

###异步查询示例状态###
###如果是同步的
curl -i -X GET http://asiainfoLDP:2016asia@localhost:8000/v2/service_instances/9BECC443-7BBC-411F-AEDA-60070173DAE9/last_operation
###如果是异步创建的
curl -i -X GET http://asiainfoLDP:2016asia@localhost:8000/v2/service_instances/4A0CA076-D00E-4945-BC49-2FD65ECAD2D1/last_operation

#测试绑定
##正确的案例，同步模式
curl -i -X PUT http://asiainfoLDP:2016asia@localhost:8000/v2/service_instances/9BECC443-7BBC-411F-AEDA-60070173DAE9/service_bindings/6853A95B-428B-4C10-99FC-1BE6CBFBE176 -d '{
  "plan_id":        "E28FB3AE-C237-484F-AC9D-FB0A80223F85",
  "service_id":     "A25DE423-484E-4252-B6FE-EA4F347BCE3D",
  "app_guid":       "app-guid"
}' -H "Content-Type: application/json"
##测试各种错误
curl -i -X PUT http://asiainfoLDP:2016asia@localhost:8000/v2/service_instances/instance_guid-222/service_bindings/binding_guid-222 -d '{
  "plan_id":        "plan-guid",
  "service_id":     "service-guid-222",
  "app_guid":       "app-guid"
}' -H "Content-Type: application/json"

##测试异步绑定
curl -i -X PUT http://asiainfoLDP:2016asia@localhost:8000/v2/service_instances/4A0CA076-D00E-4945-BC49-2FD65ECAD2D1/service_bindings/039D85CE-026B-44F4-991F-D8EA62DB334B -d '{
  "plan_id":        "8C7E1AB9-DB63-4E14-9487-733BB587E1B2",
  "service_id":     "A25DE423-484E-4252-B6FE-EA4F347BCE3D",
  "app_guid":       "app-guid"
}' -H "Content-Type: application/json"

curl -i -X DELETE http://asiainfoLDP:2016asia@localhost:8000/v2/service_instances/98A763D7-CE08-4E0D-B139-769F80B6DEFD/service_bindings/6853A95B-428B-4C10-99FC-1BE6CBFBE176
curl -i -X DELETE http://asiainfoLDP:2016asia@localhost:8000/v2/service_instances/98A763D7-CE08-4E0D-B139-769F80B6DEFD
#正确的删除绑定
curl -i -X DELETE 'http://asiainfoLDP:2016asia@localhost:8000/v2/service_instances/9BECC443-7BBC-411F-AEDA-60070173DAE9/service_bindings/6853A95B-428B-4C10-99FC-1BE6CBFBE176?service_id=A25DE423-484E-4252-B6FE-EA4F347BCE3D&plan_id=E28FB3AE-C237-484F-AC9D-FB0A80223F85' 
curl -i -X DELETE http://asiainfoLDP:2016asia@localhost:8000/v2/service_instances/9BECC443-7BBC-411F-AEDA-60070173DAE9
#正确的删除实例
curl -i -X DELETE 'http://asiainfoLDP:2016asia@localhost:8000/v2/service_instances/9BECC443-7BBC-411F-AEDA-60070173DAE9?service_id=A25DE423-484E-4252-B6FE-EA4F347BCE3D&plan_id=E28FB3AE-C237-484F-AC9D-FB0A80223F85' 

###删除异步的aws上的实例
curl -i -X DELETE 'http://asiainfoLDP:2016asia@localhost:8000/v2/service_instances/4A0CA076-D00E-4945-BC49-2FD65ECAD2D1?service_id=A25DE423-484E-4252-B6FE-EA4F347BCE3D&plan_id=8C7E1AB9-DB63-4E14-9487-733BB587E1B2' 

###解除异步aws实例的绑定
curl -i -X DELETE 'http://asiainfoLDP:2016asia@localhost:8000/v2/service_instances/4A0CA076-D00E-4945-BC49-2FD65ECAD2D1/service_bindings/039D85CE-026B-44F4-991F-D8EA62DB334B?service_id=A25DE423-484E-4252-B6FE-EA4F347BCE3D&plan_id=8C7E1AB9-DB63-4E14-9487-733BB587E1B2' 

### update 测试

curl -i -X PATCH 'http://asiainfoLDP:2016asia@localhost:8000/v2/service_instances/9BECC443-7BBC-411F-AEDA-60070173DAE9' 

-----------
#创建一个shareandcommon
curl -i -X PUT http://asiainfoLDP:2016asia@localhost:8000/v2/service_instances/bf59ec19-2da1-4a47-9a75-30df7e57fd53 -d '{
  "service_id":"A25DE423-484E-4252-B6FE-EA4F347BCE3D",
  "plan_id":"257C6C2B-A376-4551-90E8-82D4E619C852",
  "organization_guid": "org-guid",
  "space_guid":"space-guid",
  "parameters": {"ami_id":"ami-ecb68a84"}
}' -H "Content-Type: application/json"

##正确的案例，同步模式
curl -i -X PUT http://asiainfoLDP:2016asia@localhost:8000/v2/service_instances/bf59ec19-2da1-4a47-9a75-30df7e57fd53/service_bindings/ACAF564C-C3F6-4793-A96D-C86DED709E66 -d '{
  "plan_id":        "257C6C2B-A376-4551-90E8-82D4E619C852",
  "service_id":     "A25DE423-484E-4252-B6FE-EA4F347BCE3D",
  "app_guid":       "app-guid"
}' -H "Content-Type: application/json"

#正确地删除绑定
curl -i -X DELETE 'http://asiainfoLDP:2016asia@localhost:8000/v2/service_instances/bf59ec19-2da1-4a47-9a75-30df7e57fd53/service_bindings/ACAF564C-C3F6-4793-A96D-C86DED709E66?service_id=A25DE423-484E-4252-B6FE-EA4F347BCE3D&plan_id=257C6C2B-A376-4551-90E8-82D4E619C852' 


#正确的删除实例
curl -i -X DELETE 'http://asiainfoLDP:2016asia@localhost:8000/v2/service_instances/bf59ec19-2da1-4a47-9a75-30df7e57fd53?service_id=A25DE423-484E-4252-B6FE-EA4F347BCE3D&plan_id=257C6C2B-A376-4551-90E8-82D4E619C852' 


------mysql  shared测试-----

#创建一个shareandcommon
curl -i -X PUT http://asiainfoLDP:2016asia@localhost:8000/v2/service_instances/809EB82E-BAEC-4E24-A672-A63704B0C7A8 -d '{
  "service_id":"7D2AB7B3-8AEF-45EE-BFF2-64A767DDE9DA",
  "plan_id":"56660431-6032-43D0-A114-FFA3BF521B66",
  "organization_guid": "org-guid",
  "space_guid":"space-guid",
  "parameters": {"ami_id":"ami-ecb68a84"}
}' -H "Content-Type: application/json"

##正确的案例，同步模式
curl -i -X PUT http://asiainfoLDP:2016asia@localhost:8000/v2/service_instances/809EB82E-BAEC-4E24-A672-A63704B0C7A8/service_bindings/8BE63643-72F2-49AA-A89A-85EBFA999EF7 -d '{
  "plan_id":        "56660431-6032-43D0-A114-FFA3BF521B66",
  "service_id":     "7D2AB7B3-8AEF-45EE-BFF2-64A767DDE9DA",
  "app_guid":       "app-guid"
}' -H "Content-Type: application/json"

#正确地删除绑定
curl -i -X DELETE 'http://asiainfoLDP:2016asia@localhost:8000/v2/service_instances/809EB82E-BAEC-4E24-A672-A63704B0C7A8/service_bindings/8BE63643-72F2-49AA-A89A-85EBFA999EF7?service_id=7D2AB7B3-8AEF-45EE-BFF2-64A767DDE9DA&plan_id=56660431-6032-43D0-A114-FFA3BF521B66' 


#正确的删除实例
curl -i -X DELETE 'http://asiainfoLDP:2016asia@localhost:8000/v2/service_instances/809EB82E-BAEC-4E24-A672-A63704B0C7A8?service_id=7D2AB7B3-8AEF-45EE-BFF2-64A767DDE9DA&plan_id=56660431-6032-43D0-A114-FFA3BF521B66' 



