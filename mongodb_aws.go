package main

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/coreos/etcd/client"
	"github.com/pivotal-cf/brokerapi"
	"github.com/pivotal-golang/lager"
	"golang.org/x/net/context"
	"gopkg.in/mgo.v2"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type myServiceBroker struct {
}

type serviceInfo struct {
	Service_name   string `json:"service_name"`
	Plan_name      string `json:"plan_name"`
	Url            string `json:"url"`
	Admin_user     string `json:"admin_user,omitempty"`
	Admin_password string `json:"admin_password,omitempty"`
	Database       string `json:"database,omitempty"`
	User           string `json:"user"`
	Password       string `json:"password"`
}

type myCredentials struct {
	Uri      string `json:"uri"`
	Hostname string `json:"host"`
	Port     string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Name     string `json:"name,omitempty"`
}

func (myBroker *myServiceBroker) Services() []brokerapi.Service {
	//初始化一系列所需要的结构体，好累啊
	myServices := []brokerapi.Service{}
	myService := brokerapi.Service{}
	myPlans := []brokerapi.ServicePlan{}
	myPlan := brokerapi.ServicePlan{}
	var myPlanfree bool
	//todo还需要考虑对于service和plan的隐藏参数，status，比如可以用，不可用，已经删除等。删除应该是软删除，后两者不予以显示，前者表示还有数据
	//获取catalog信息
	resp, err := etcdapi.Get(context.Background(), "/servicebroker/"+servcieBrokerName+"/catalog", &client.GetOptions{Recursive: true}) //改为环境变量
	if err != nil {
		logger.Error("Can not get catalog information from etcd", err) //所有这些出错消息最好命名为常量，放到开始的时候
	} else {
		logger.Debug("Successful get catalog information from etcd. NodeInfo is " + resp.Node.Key)
	}

	for i := 0; i < len(resp.Node.Nodes); i++ {
		//为旗下发现的每一个service进行迭代，不过一般情况下，应该只有一个service
		logger.Debug("Start to Parse Service " + resp.Node.Nodes[i].Key)
		//在下一级循环外设置id，因为他是目录名字，注意，如果按照这个逻辑，id一定要是uuid，中间一定不能有目录符号"/"
		myService.ID = strings.Split(resp.Node.Nodes[i].Key, "/")[len(strings.Split(resp.Node.Nodes[i].Key, "/"))-1]
		//开始取service级别除了ID以外的其他参数
		for j := 0; j < len(resp.Node.Nodes[i].Nodes); j++ {
			if !resp.Node.Nodes[i].Nodes[j].Dir {
				switch strings.ToLower(resp.Node.Nodes[i].Nodes[j].Key) {
				case strings.ToLower(resp.Node.Nodes[i].Key) + "/name":
					myService.Name = resp.Node.Nodes[i].Nodes[j].Value
				case strings.ToLower(resp.Node.Nodes[i].Key) + "/description":
					myService.Description = resp.Node.Nodes[i].Nodes[j].Value
				case strings.ToLower(resp.Node.Nodes[i].Key) + "/bindable":
					myService.Bindable, _ = strconv.ParseBool(resp.Node.Nodes[i].Nodes[j].Value)
				case strings.ToLower(resp.Node.Nodes[i].Key) + "/tags":
					myService.Tags = strings.Split(resp.Node.Nodes[i].Nodes[j].Value, ",")
				case strings.ToLower(resp.Node.Nodes[i].Key) + "/planupdatable":
					myService.PlanUpdatable, _ = strconv.ParseBool(resp.Node.Nodes[i].Nodes[j].Value)
				case strings.ToLower(resp.Node.Nodes[i].Key) + "/metadata":
					json.Unmarshal([]byte(resp.Node.Nodes[i].Nodes[j].Value), &myService.Metadata)
				}
			} else if strings.HasSuffix(strings.ToLower(resp.Node.Nodes[i].Nodes[j].Key), "plan") {
				//开始解析套餐目录中的套餐计划plan。上述判断也不是太严谨，比如有目录如果是xxxxplan怎么办？
				for k := 0; k < len(resp.Node.Nodes[i].Nodes[j].Nodes); k++ {
					logger.Debug("Start to Parse Plan " + resp.Node.Nodes[i].Nodes[j].Nodes[k].Key)
					myPlan.ID = strings.Split(resp.Node.Nodes[i].Nodes[j].Nodes[k].Key, "/")[len(strings.Split(resp.Node.Nodes[i].Nodes[j].Nodes[k].Key, "/"))-1]
					for n := 0; n < len(resp.Node.Nodes[i].Nodes[j].Nodes[k].Nodes); n++ {
						switch strings.ToLower(resp.Node.Nodes[i].Nodes[j].Nodes[k].Nodes[n].Key) {
						case strings.ToLower(resp.Node.Nodes[i].Nodes[j].Nodes[k].Key) + "/name":
							myPlan.Name = resp.Node.Nodes[i].Nodes[j].Nodes[k].Nodes[n].Value
						case strings.ToLower(resp.Node.Nodes[i].Nodes[j].Nodes[k].Key) + "/description":
							myPlan.Description = resp.Node.Nodes[i].Nodes[j].Nodes[k].Nodes[n].Value
						case strings.ToLower(resp.Node.Nodes[i].Nodes[j].Nodes[k].Key) + "/free":
							//这里没有搞懂为什么brokerapi里面的这个bool要定义为传指针的模式
							myPlanfree, _ = strconv.ParseBool(resp.Node.Nodes[i].Nodes[j].Nodes[k].Nodes[n].Value)
							myPlan.Free = brokerapi.FreeValue(myPlanfree)
						case strings.ToLower(resp.Node.Nodes[i].Nodes[j].Nodes[k].Key) + "/metadata":
							json.Unmarshal([]byte(resp.Node.Nodes[i].Nodes[j].Nodes[k].Nodes[n].Value), &myPlan.Metadata)
						}
					}
					//装配plan需要返回的值，按照有多少个plan往里面装
					myPlans = append(myPlans, myPlan)
					//重置服务变量
					myPlan = brokerapi.ServicePlan{}

				}
				//将装配好的Plan对象赋值给Service
				myService.Plans = myPlans
			}
		}

		//装配catalog需要返回的值，按照有多少个服务往里面装
		myServices = append(myServices, myService)
		//重置服务变量
		myService = brokerapi.Service{}

	}

	return myServices

}

func (myBroker *myServiceBroker) Provision(
	instanceID string,
	details brokerapi.ProvisionDetails,
	asyncAllowed bool,
) (brokerapi.ProvisionedServiceSpec, error) {

	//初始化
	var DashboardURL string
	var provsiondetail brokerapi.ProvisionedServiceSpec
	var myServiceInfo serviceInfo
	var newpassword, newusername string

	//判断实例是否已经存在，如果存在就报错
	resp, err := etcdget("/servicebroker/" + servcieBrokerName + "/instance") //改为环境变量

	if err != nil {
		logger.Error("Can't connet to etcd", err)
		return brokerapi.ProvisionedServiceSpec{}, errors.New("Can't connet to etcd")
	}

	for i := 0; i < len(resp.Node.Nodes); i++ {
		if resp.Node.Nodes[i].Dir && strings.HasSuffix(resp.Node.Nodes[i].Key, instanceID) {
			logger.Info("ErrInstanceAlreadyExists")
			return brokerapi.ProvisionedServiceSpec{}, brokerapi.ErrInstanceAlreadyExists
		}
	}

	//判断servcie_id和plan_id是否正确
	service_name := findServiceNameInCatalog(details.ServiceID)
	plan_name := findServicePlanNameInCatalog(details.ServiceID, details.PlanID)
	//todo 应该修改service broker添加一个用户输入出错的返回，而不是500
	if service_name == "" || plan_name == "" {
		logger.Info("Service_id or plan_id not correct!!")
		return brokerapi.ProvisionedServiceSpec{}, errors.New("Service_id or plan_id not correct!!")
	}
	//是否要检查service和plan的status是否允许创建 todo

	//根据不同的服务和plan，选择创建的命令 ［每次增加不同的服务或者计划，只需要修改这里就好了。］
	switch service_name {
	case managedServiceName: //需要配置为Service的环境变量
		//开始根据不同的plan进行处理
		switch plan_name {
		case "shared":
			//初始化mongodb的链接串
			session, err := mgo.Dial(mongoUrl) //连接数据库
			if err != nil {
				logger.Error("Can't connet to mongodb", err)
				return brokerapi.ProvisionedServiceSpec{}, errors.New("Can't connet to mongodb " + mongoUrl)
			}
			defer session.Close()
			session.SetMode(mgo.Monotonic, true)
			mongodb := session.DB("admin") //数据库名称
			err = mongodb.Login(mongoAdminUser, mongoAdminPassword)
			if err != nil {
				logger.Error("Can't Login to mongodb", err)
				return brokerapi.ProvisionedServiceSpec{}, errors.New("Can't Login to mongodb " + mongoUrl)
			}

			//创建一个名为instanceID的数据库，并随机的创建用户名和密码，这个用户名是该数据库的管理员
			newdb := session.DB(instanceID)
			newusername = getguid()
			newpassword = getguid()
			//为dashbord赋值 todo dashboard应该提供一个界面才对
			DashboardURL = "mongodb://" + newusername + ":" + newpassword + "@" + mongoUrl + "/" + instanceID
			//这个服务很快，所以通过同步模式直接返回了
			err = newdb.UpsertUser(&mgo.User{
				Username: newusername,
				Password: newpassword,
				Roles: []mgo.Role{
					mgo.Role(mgo.RoleDBAdmin),
				},
			})

			if err != nil {
				logger.Error("Can't Create User in mongodb", err)
				return brokerapi.ProvisionedServiceSpec{}, errors.New("Can't Create User in mongodb " + mongoUrl)
			} else {
				logger.Debug("Success Create User in mongodb. Username=" + newusername + " Password=" + newpassword)
			}

			//赋值隐藏属性
			myServiceInfo = serviceInfo{
				Service_name:   service_name,
				Plan_name:      plan_name,
				Url:            mongoUrl,
				Admin_user:     mongoAdminUser,
				Admin_password: mongoAdminPassword,
				Database:       instanceID,
				User:           newusername,
				Password:       newpassword,
			}

			provsiondetail = brokerapi.ProvisionedServiceSpec{DashboardURL: DashboardURL, IsAsync: false}

		case "standalone":

			//需要有两个环境变量 AWS_ACCESS_KEY_ID AWS_SECRET_ACCESS_KEY

			//初始化aws client
			svc := ec2.New(session.New(&aws.Config{Region: aws.String(awsRegion)}))
			//准备管理用户名和密码
			newusername = getguid()
			newpassword = getguid()
			//拼接启动aws实例用的userdata
			//userdate sample
			//#!/bin/bash
			//mongo <<!!!
			//use admin
			//db.auth('asiainfoLDP','6ED9BA74-75FD-4D1B-8916-842CB936AC1A');
			//db.createUser({ user: 'test', pwd: 'test', roles: [ { role: 'root', db: 'admin' } ] });
			//!!!

			userdata := "#!/bin/bash \n mongo <<!!! \n use admin \n db.auth('" + mongoAdminUser + "','" + mongoAdminPassword + "'); \n db.createUser({ user: '" + newusername + "', pwd: '" + newpassword + "', roles: [ { role: 'root', db: 'admin' } ] }); \n !!!"
			// Specify the details of the instance that you want to create.
			runResult, err := svc.RunInstances(&ec2.RunInstancesInput{
				ImageId:        aws.String(imageId),
				InstanceType:   aws.String(instanceType),
				KeyName:        aws.String(keyName),
				SecurityGroups: []*string{aws.String(securityGroups)},
				MinCount:       aws.Int64(1),
				MaxCount:       aws.Int64(1),
				UserData:       aws.String(base64.StdEncoding.EncodeToString([]byte(userdata))),
			})

			if err != nil {
				logger.Error("Could not create aws instance", err)
				return brokerapi.ProvisionedServiceSpec{}, errors.New("Could not create aws instance")
			}

			logger.Info("Created aws instance " + *runResult.Instances[0].InstanceId)
			awsmongourl := *runResult.Instances[0].PrivateIpAddress + ":27017"

			// Add tags to the created instance
			_, errtag := svc.CreateTags(&ec2.CreateTagsInput{
				Resources: []*string{runResult.Instances[0].InstanceId},
				Tags: []*ec2.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String("service_mongodb_" + *runResult.Instances[0].InstanceId),
					},
				},
			})
			if errtag != nil {
				logger.Error("Could not create tags for instance", err)
				return brokerapi.ProvisionedServiceSpec{}, errors.New("Could not create tags for instance" + *runResult.Instances[0].InstanceId)
			}
			logger.Info("Successfully create aws instance and tagged" + *runResult.Instances[0].InstanceId)

			//赋值隐藏属性
			myServiceInfo = serviceInfo{
				Service_name:   service_name,
				Plan_name:      plan_name,
				Url:            *runResult.Instances[0].InstanceId, //aws的实例id
				Admin_user:     mongoAdminUser,
				Admin_password: mongoAdminPassword,
				Database:       "admin",
				User:           newusername,
				Password:       newpassword,
			}

			//为dashbord赋值 todo dashboard应该提供一个界面才对
			DashboardURL = "mongodb://" + newusername + ":" + newpassword + "@" + awsmongourl
			//表示是异步返回
			provsiondetail = brokerapi.ProvisionedServiceSpec{DashboardURL: DashboardURL, IsAsync: true}

		default: //没有相关处理handle应该报错才对
			logger.Info("No Plan Handle for " + plan_name)
			return brokerapi.ProvisionedServiceSpec{}, errors.New("No Plan Handle for " + plan_name)

		}
	default: //没有相关的处理handle应该报错才对
		logger.Info("No Service Handle for " + service_name)
		return brokerapi.ProvisionedServiceSpec{}, errors.New("No Service Handle for " + service_name)
	}

	//写入etcd 话说如果这个时候写入失败，那不就出现数据不一致的情况了么！todo
	//先创建instanceid目录
	_, err = etcdapi.Set(context.Background(), "/servicebroker/"+servcieBrokerName+"/instance/"+instanceID, "", &client.SetOptions{Dir: true}) //todo这些要么是常量，要么应该用环境变量
	if err != nil {
		logger.Error("Can not create instance "+instanceID+" in etcd", err) //todo都应该改为日志key
		return brokerapi.ProvisionedServiceSpec{}, err
	} else {
		logger.Debug("Successful create instance "+instanceID+" in etcd", nil)
	}
	//然后创建一系列属性
	etcdset("/servicebroker/"+servcieBrokerName+"/instance/"+instanceID+"/organization_guid", details.OrganizationGUID)
	etcdset("/servicebroker/"+servcieBrokerName+"/instance/"+instanceID+"/space_guid", details.SpaceGUID)
	etcdset("/servicebroker/"+servcieBrokerName+"/instance/"+instanceID+"/service_id", details.ServiceID)
	etcdset("/servicebroker/"+servcieBrokerName+"/instance/"+instanceID+"/plan_id", details.PlanID)
	tmpval, _ := json.Marshal(details.Parameters)
	etcdset("/servicebroker/"+servcieBrokerName+"/instance/"+instanceID+"/parameters", string(tmpval))
	etcdset("/servicebroker/"+servcieBrokerName+"/instance/"+instanceID+"/dashboardurl", DashboardURL)
	//存储隐藏信息_info
	tmpval, _ = json.Marshal(myServiceInfo)
	etcdset("/servicebroker/"+servcieBrokerName+"/instance/"+instanceID+"/_info", string(tmpval))

	//创建绑定目录
	_, err = etcdapi.Set(context.Background(), "/servicebroker/"+servcieBrokerName+"/instance/"+instanceID+"/binding", "", &client.SetOptions{Dir: true})

	//完成所有操作后，返回DashboardURL和是否异步的标志
	logger.Info("Successful create instance " + instanceID)
	return provsiondetail, nil
}

func (myBroker *myServiceBroker) LastOperation(instanceID string) (brokerapi.LastOperation, error) {
	// If the broker provisions asynchronously, the Cloud Controller will poll this endpoint
	// for the status of the provisioning operation.

	//去读取进展状态，如果有错误，返回错误，如果没有错误，返回对象LastOperation! todo 如果同步模式，不用实现这个接口

	var myServiceInfo serviceInfo
	var lastOperation brokerapi.LastOperation
	//判断实例是否已经存在，如果不存在就报错
	resp, err := etcdapi.Get(context.Background(), "/servicebroker/"+servcieBrokerName+"/instance/"+instanceID, &client.GetOptions{Recursive: true}) //改为环境变量

	if err != nil || !resp.Node.Dir {
		logger.Error("Can not get instance information from etcd", err)
		return brokerapi.LastOperation{}, brokerapi.ErrInstanceDoesNotExist
	} else {
		logger.Debug("Successful get instance information from etcd. NodeInfo is " + resp.Node.Key)
	}

	//隐藏属性不得不单独获取
	resp, err = etcdget("/servicebroker/" + servcieBrokerName + "/instance/" + instanceID + "/_info")
	json.Unmarshal([]byte(resp.Node.Value), &myServiceInfo)

	//根据不同的服务和plan，选择创建的命令 ［每次增加不同的服务或者计划，只需要修改这里就好了。］
	switch myServiceInfo.Service_name {
	case managedServiceName:
		switch myServiceInfo.Plan_name {
		case "shared":
			//因为是同步模式，协议里面并没有说怎么处理啊，统一反馈成功吧！
			lastOperation = brokerapi.LastOperation{
				State:       brokerapi.Succeeded,
				Description: "It's a sync method!",
			}
		case "standalone":
			//初始化aws client
			svc := ec2.New(session.New(&aws.Config{Region: aws.String(awsRegion)}))
			//从service_instance的etcd目录中取出aws instance的id，以便查询
			awsInstaceID := myServiceInfo.Url
			//查询实例
			runResult, err := svc.DescribeInstances(&ec2.DescribeInstancesInput{
				InstanceIds: []*string{
					aws.String(awsInstaceID), // Required
					// More values...
				},
			})

			//出错后显示出错
			if err != nil {
				logger.Error("Could not get aws instance status "+awsInstaceID, err)
				return brokerapi.LastOperation{}, errors.New("Could not get aws instance status " + awsInstaceID)
			}

			//注意：由于service_broker.go中并没有像规范里面一样定义410 gone这个状态，因此，暂时不对update和deporivion返回 to
			if len(runResult.Reservations) == 0 {
				logger.Error("aws instance status not exist"+awsInstaceID, err)
				return brokerapi.LastOperation{}, errors.New("aws instance status not exist " + awsInstaceID)
			}

			switch *runResult.Reservations[0].Instances[0].State.Name {
			case "pending":
				lastOperation = brokerapi.LastOperation{
					State:       brokerapi.InProgress,
					Description: "creating service instance " + awsInstaceID,
				}
			case "running":
				lastOperation = brokerapi.LastOperation{
					State:       brokerapi.Succeeded,
					Description: "successfully created service instance " + awsInstaceID,
				}
			default:
				lastOperation = brokerapi.LastOperation{
					State:       brokerapi.Failed,
					Description: "failed to create service instance " + awsInstaceID,
				}
			}
		default: //没有相关处理handle应该报错才对
			logger.Info("No Plan Handle for " + myServiceInfo.Service_name)
			return brokerapi.LastOperation{}, errors.New("No Plan Handle for " + myServiceInfo.Service_name)
		}
	default: //没有相关的处理handle应该报错才对
		logger.Info("No Service Handle for " + myServiceInfo.Plan_name)
		return brokerapi.LastOperation{}, errors.New("No Service Handle for " + myServiceInfo.Plan_name)
	}

	//一切正常，返回结果
	logger.Info("Successful query last operation for service instance" + instanceID)
	return lastOperation, nil
}

func (myBroker *myServiceBroker) Deprovision(instanceID string, details brokerapi.DeprovisionDetails, asyncAllowed bool) (brokerapi.IsAsync, error) {

	var myServiceInfo serviceInfo

	//判断实例是否已经存在，如果不存在就报错
	resp, err := etcdapi.Get(context.Background(), "/servicebroker/"+servcieBrokerName+"/instance/"+instanceID, &client.GetOptions{Recursive: true})

	if err != nil || !resp.Node.Dir {
		logger.Error("Can not get instance information from etcd", err)
		return brokerapi.IsAsync(false), brokerapi.ErrInstanceDoesNotExist
	} else {
		logger.Debug("Successful get instance information from etcd. NodeInfo is " + resp.Node.Key)
	}

	var servcie_id, plan_id string

	//从etcd中取得参数。
	for i := 0; i < len(resp.Node.Nodes); i++ {
		if !resp.Node.Nodes[i].Dir {
			switch strings.ToLower(resp.Node.Nodes[i].Key) {
			case strings.ToLower(resp.Node.Key) + "/service_id":
				servcie_id = resp.Node.Nodes[i].Value
			case strings.ToLower(resp.Node.Key) + "/plan_id":
				plan_id = resp.Node.Nodes[i].Value
			}
		}
	}

	//并且要核对一下detail里面的service_id和plan_id。出错消息现在是500，需要更改一下源代码，以便更改出错代码
	if servcie_id != details.ServiceID || plan_id != details.PlanID {
		logger.Info("ServiceID or PlanID not correct!!")
		return brokerapi.IsAsync(false), errors.New("ServiceID or PlanID not correct!! instanceID " + instanceID)
	}
	//是否要判断里面有没有绑定啊？todo

	//根据存储在etcd中的service_name和plan_name来确定到底调用那一段处理。注意这个时候不能像Provision一样去catalog里面读取了。
	//因为这个时候的数据不一定和创建的时候一样，plan等都有可能变化。同样的道理，url，用户名，密码都应该从_info中解码出来

	//隐藏属性不得不单独获取
	resp, err = etcdget("/servicebroker/" + servcieBrokerName + "/instance/" + instanceID + "/_info")
	json.Unmarshal([]byte(resp.Node.Value), &myServiceInfo)

	//根据不同的服务和plan，选择创建的命令 ［每次增加不同的服务或者计划，只需要修改这里就好了。］
	switch myServiceInfo.Service_name {
	case managedServiceName:
		switch myServiceInfo.Plan_name {
		case "shared":
			//初始化mongodb的链接串
			session, err := mgo.Dial(myServiceInfo.Url) //连接数据库
			if err != nil {
				logger.Error("Can't connet to mongodb "+myServiceInfo.Url, err)
				return brokerapi.IsAsync(false), errors.New("Can't connet to mongodb " + myServiceInfo.Url)
			}
			defer session.Close()
			session.SetMode(mgo.Monotonic, true)
			mongodb := session.DB("admin") //数据库名称
			err = mongodb.Login(myServiceInfo.Admin_user, myServiceInfo.Admin_password)
			if err != nil {
				logger.Error("Can't Login to mongodb "+myServiceInfo.Url, err)
				return brokerapi.IsAsync(false), errors.New("Can't Login to mongodb " + myServiceInfo.Url)
			}

			//选择服务创建的数据库
			userdb := session.DB(myServiceInfo.Database)
			//这个服务很快，所以通过同步模式直接返回了
			err = userdb.DropDatabase()

			if err != nil {
				logger.Error("Can't DropDatabase in mongodb", err)
				return brokerapi.IsAsync(false), errors.New("Can't DropDatabase in mongodb " + myServiceInfo.Url)
			} else {
				logger.Debug("Success DropDatabase in mongodb. database name=" + myServiceInfo.Database)
			}

		case "standalone":
			//初始化aws client
			svc := ec2.New(session.New(&aws.Config{Region: aws.String(awsRegion)}))
			//从service_instance的etcd目录中取出aws instance的id，以便查询
			awsInstaceID := myServiceInfo.Url
			//查询实例
			runResult, err := svc.DescribeInstances(&ec2.DescribeInstancesInput{
				InstanceIds: []*string{
					aws.String(awsInstaceID), // Required
					// More values...
				},
			})

			//出错后显示出错
			if err != nil {
				logger.Error("Could not get aws instance status "+awsInstaceID, err)
				return brokerapi.IsAsync(false), errors.New("Could not get aws instance status " + awsInstaceID)
			}

			//没有找到任何aws的实例
			if len(runResult.Reservations) == 0 {
				logger.Error("aws instance status not exist"+awsInstaceID, err)
				return brokerapi.IsAsync(false), errors.New("aws instance status not exist " + awsInstaceID)
			}

			switch *runResult.Reservations[0].Instances[0].State.Name {
			case "pending":
				//不可以执行的状态，反馈400，但是目前servie_broker的包里面没有这个状态的判断，只有暂时返回500 todo
				logger.Error("Another operation for this service instance is in progress", err)
				return brokerapi.IsAsync(false), errors.New("Another operation for this service instance is in progress")
			case "running":
				//可以删除的状态

				//开始删除实例
				_, err := svc.TerminateInstances(&ec2.TerminateInstancesInput{
					InstanceIds: []*string{
						aws.String(awsInstaceID), // Required
						// More values...
					},
				})

				if err != nil {
					logger.Error("can not delete aws instance "+awsInstaceID, err)
					return brokerapi.IsAsync(false), errors.New("can not delete aws instance " + awsInstaceID)
				}
				//成功删除
				logger.Info("successfully detele aws instance " + awsInstaceID)

			default:
				//不可以执行的状态，反馈400，但是目前servie_broker的包里面没有这个状态的判断，只有暂时返回500 todo
				logger.Error("Another operation for this service instance is in progress", err)
				return brokerapi.IsAsync(false), errors.New("Another operation for this service instance is in progress")
			}

		default: //没有相关处理handle应该报错才对
			logger.Info("No Plan Handle for " + myServiceInfo.Service_name)
			return brokerapi.IsAsync(false), errors.New("No Plan Handle for " + myServiceInfo.Service_name)
		}
	default: //没有相关的处理handle应该报错才对
		logger.Info("No Service Handle for " + myServiceInfo.Plan_name)
		return brokerapi.IsAsync(false), errors.New("No Service Handle for " + myServiceInfo.Plan_name)
	}

	//然后删除etcd里面的纪录，这里也有可能有不一致的情况
	_, err = etcdapi.Delete(context.Background(), "/servicebroker/"+servcieBrokerName+"/instance/"+instanceID, &client.DeleteOptions{Recursive: true, Dir: true}) //todo这些要么是常量，要么应该用环境变量
	if err != nil {
		logger.Error("Can not delete instance "+instanceID+" in etcd", err) //todo都应该改为日志key
		return brokerapi.IsAsync(false), errors.New("Internal Error!!")
	} else {
		logger.Debug("Successful delete instance " + instanceID + " in etcd")
	}

	logger.Info("Successful Deprovision instance " + instanceID)
	return brokerapi.IsAsync(false), nil
}

func (myBroker *myServiceBroker) Bind(instanceID, bindingID string, details brokerapi.BindDetails) (brokerapi.Binding, error) {
	var mycredentials myCredentials
	var myBinding brokerapi.Binding
	//判断实例是否已经存在，如果不存在就报错
	resp, err := etcdget("/servicebroker/" + servcieBrokerName + "/instance/" + instanceID)
	if err != nil || !resp.Node.Dir {
		logger.Error("Can not get instance information from etcd", err) //所有这些出错消息最好命名为常量，放到开始的时候
		return brokerapi.Binding{}, brokerapi.ErrInstanceDoesNotExist
	} else {
		logger.Debug("Successful get instance information from etcd. NodeInfo is " + resp.Node.Key)
	}

	//判断绑定是否存在，如果存在就报错
	resp, err = etcdget("/servicebroker/" + servcieBrokerName + "/instance/" + instanceID + "/binding")
	for i := 0; i < len(resp.Node.Nodes); i++ {
		if resp.Node.Nodes[i].Dir && strings.HasSuffix(resp.Node.Nodes[i].Key, bindingID) {
			logger.Info("ErrBindingAlreadyExists " + instanceID)
			return brokerapi.Binding{}, brokerapi.ErrBindingAlreadyExists
		}
	}

	//对于参数中的service_id和plan_id仅做校验，不再在binding中存储
	var servcie_id, plan_id string

	//从etcd中取得参数。
	resp, err = etcdapi.Get(context.Background(), "/servicebroker/"+servcieBrokerName+"/instance/"+instanceID, &client.GetOptions{Recursive: true}) //改为环境变量

	for i := 0; i < len(resp.Node.Nodes); i++ {
		if !resp.Node.Nodes[i].Dir {
			switch strings.ToLower(resp.Node.Nodes[i].Key) {
			case strings.ToLower(resp.Node.Key) + "/service_id":
				servcie_id = resp.Node.Nodes[i].Value
			case strings.ToLower(resp.Node.Key) + "/plan_id":
				plan_id = resp.Node.Nodes[i].Value
			}
		}
	}

	//并且要核对一下detail里面的service_id和plan_id。出错消息现在是500，需要更改一下源代码，以便更改出错代码
	if servcie_id != details.ServiceID || plan_id != details.PlanID {
		logger.Info("ServiceID or PlanID not correct!!")
		return brokerapi.Binding{}, errors.New("ServiceID or PlanID not correct!! instanceID " + instanceID)
	}

	//隐藏属性不得不单独获取。取得当时绑定服务得到信息
	var myServiceInfo serviceInfo
	resp, err = etcdget("/servicebroker/" + servcieBrokerName + "/instance/" + instanceID + "/_info")
	json.Unmarshal([]byte(resp.Node.Value), &myServiceInfo)

	//根据绑定要求，在数据库里面创建一个readwrite权限的用户
	//根据不同的服务和plan，选择创建的命令 ［每次增加不同的服务或者计划，只需要修改这里就好了。］
	switch myServiceInfo.Service_name {
	case managedServiceName:
		//由于bind的处理逻辑对于shared模式和standalone模式差不多
		var mongodburl string
		var mongodbname string
		var mongodbrole mgo.Role
		//判断采用何种模式
		switch myServiceInfo.Plan_name {
		case "shared":
			//初始化mongodb的两个变量
			mongodburl = myServiceInfo.Url
			//share 模式只能是该数据库
			mongodbname = myServiceInfo.Database
			//share 模式，只是这个数据库的读写
			mongodbrole = mgo.RoleReadWrite
		case "standalone":
			//初始化aws client
			svc := ec2.New(session.New(&aws.Config{Region: aws.String(awsRegion)}))
			//从service_instance的etcd目录中取出aws instance的id，以便查询
			awsInstaceID := myServiceInfo.Url
			//查询实例
			runResult, err := svc.DescribeInstances(&ec2.DescribeInstancesInput{
				InstanceIds: []*string{
					aws.String(awsInstaceID), // Required
					// More values...
				},
			})

			//出错后显示出错
			if err != nil {
				logger.Error("Could not get aws instance status "+awsInstaceID, err)
				return brokerapi.Binding{}, errors.New("Could not get aws instance status " + awsInstaceID)
			}

			//没有找到任何aws的实例
			if len(runResult.Reservations) == 0 {
				logger.Error("aws instance status not exist"+awsInstaceID, err)
				return brokerapi.Binding{}, errors.New("aws instance status not exist " + awsInstaceID)
			}

			switch *runResult.Reservations[0].Instances[0].State.Name {
			case "pending":
				//不可以执行的状态，反馈400，但是目前servie_broker的包里面没有这个状态的判断，只有暂时返回500 todo
				logger.Error("Another operation for this service instance is in progress", err)
				return brokerapi.Binding{}, errors.New("Another operation for this service instance is in progress")
			case "running":
				//可以绑定的状态
				//暂时利用公网地址，以后都部署在一个云平台上和一个安全域内部的话，可以使用私网地址
				mongodburl = *runResult.Reservations[0].Instances[0].PublicIpAddress + ":27017"
				//对于standalone的情况，应该创建一个dbadmin on any database的角色，所以数据库应该是admin
				mongodbname = "admin"
				//角色
				mongodbrole = mgo.RoleDBAdminAny

			default:
				//不可以执行的状态，反馈400，但是目前servie_broker的包里面没有这个状态的判断，只有暂时返回500 todo
				logger.Error("Another operation for this service instance is in progress", err)
				return brokerapi.Binding{}, errors.New("Another operation for this service instance is in progress")
			}
		default: //没有相关处理handle应该报错才对
			logger.Info("No Plan Handle for " + myServiceInfo.Service_name)
			return brokerapi.Binding{}, errors.New("No Plan Handle for " + myServiceInfo.Service_name)
		} //end switch myServiceInfo.Plan_name

		//完成变量赋值以后，开始准备创建用户
		//初始化mongodb的链接串
		session, err := mgo.Dial(mongodburl) //连接数据库
		if err != nil {
			logger.Error("Can't connet to mongodb "+mongodburl, err)
			return brokerapi.Binding{}, errors.New("Can't connet to mongodb " + mongodburl)
		}
		defer session.Close()
		session.SetMode(mgo.Monotonic, true)
		mongodb := session.DB("admin") //数据库名称
		err = mongodb.Login(myServiceInfo.Admin_user, myServiceInfo.Admin_password)
		if err != nil {
			logger.Error("Can't Login to mongodb "+mongodburl, err)
			return brokerapi.Binding{}, errors.New("Can't Login to mongodb " + mongodburl)
		}

		//去创建一个用户，权限为RoleReadWrite
		userdb := session.DB(mongodbname)
		newusername := getguid()
		newpassword := getguid()
		//这个服务很快，所以通过同步模式直接返回了。再说了目前bind的协议只有同步的模式
		err = userdb.UpsertUser(&mgo.User{
			Username: newusername,
			Password: newpassword,
			Roles: []mgo.Role{
				mongodbrole,
			},
		})

		if err != nil {
			logger.Error("Can't DropDatabase in mongodb", err)
			return brokerapi.Binding{}, errors.New("Can't CreateUser in mongodb " + mongodburl + " as user:" + newusername)
		} else {
			logger.Debug("Success CreateUser in mongodb. database name="+mongodbname+" as user:"+newusername, nil)
		}
		//如果是admin，就不应该返回数据库的名字，并允许应用自己创建数据库
		if mongodbname == "admin" {
			mycredentials = myCredentials{
				Uri:      "mongo://" + newusername + ":" + newpassword + "@" + mongodburl,
				Hostname: strings.Split(mongodburl, ":")[0],
				Port:     strings.Split(mongodburl, ":")[1],
				Username: newusername,
				Password: newpassword,
			}
		} else {
			mycredentials = myCredentials{
				Uri:      "mongo://" + newusername + ":" + newpassword + "@" + mongodburl + "/" + mongodbname,
				Hostname: strings.Split(mongodburl, ":")[0],
				Port:     strings.Split(mongodburl, ":")[1],
				Username: newusername,
				Password: newpassword,
				Name:     mongodbname,
			}
		}
		myBinding = brokerapi.Binding{Credentials: mycredentials}

	default: //没有相关的处理handle应该报错才对
		logger.Info("No Service Handle for " + myServiceInfo.Plan_name)
		return brokerapi.Binding{}, errors.New("No Service Handle for " + myServiceInfo.Plan_name)
	}

	//把信息存储到etcd里面，同样这里有同步性的问题 todo怎么解决呢？
	//先创建bindingID目录
	_, err = etcdapi.Set(context.Background(), "/servicebroker/"+servcieBrokerName+"/instance/"+instanceID+"/binding/"+bindingID, "", &client.SetOptions{Dir: true}) //todo这些要么是常量，要么应该用环境变量
	if err != nil {
		logger.Error("Can not create binding "+bindingID+" in etcd", err) //todo都应该改为日志key
		return brokerapi.Binding{}, err
	} else {
		logger.Debug("Successful create binding "+bindingID+" in etcd", nil)
	}
	//然后创建一系列属性
	etcdset("/servicebroker/"+servcieBrokerName+"/instance/"+instanceID+"/binding/"+bindingID+"/app_guid", details.AppGUID)
	tmpval, _ := json.Marshal(details.Parameters)
	etcdset("/servicebroker/"+servcieBrokerName+"/instance/"+instanceID+"/binding/"+bindingID+"/parameters", string(tmpval))
	tmpval, _ = json.Marshal(myBinding)
	etcdset("/servicebroker/"+servcieBrokerName+"/instance/"+instanceID+"/binding/"+bindingID+"/binding", string(tmpval))
	//存储隐藏信息_info
	tmpval, _ = json.Marshal(mycredentials)
	etcdset("/servicebroker/"+servcieBrokerName+"/instance/"+instanceID+"/binding/"+bindingID+"/_info", string(tmpval))

	logger.Info("Successful create binding " + bindingID)
	return myBinding, nil
}

func (myBroker *myServiceBroker) Unbind(instanceID, bindingID string, details brokerapi.UnbindDetails) error {

	var mycredentials myCredentials
	var myServiceInfo serviceInfo
	//判断实例是否已经存在，如果不存在就报错
	resp, err := etcdapi.Get(context.Background(), "/servicebroker/"+servcieBrokerName+"/instance/"+instanceID, &client.GetOptions{Recursive: true}) //改为环境变量
	if err != nil || !resp.Node.Dir {
		logger.Error("Can not get instance information from etcd", err)
		return brokerapi.ErrInstanceDoesNotExist //这几个错误返回为空，是detele操作的要求吗？
	} else {
		logger.Debug("Successful get instance information from etcd. NodeInfo is " + resp.Node.Key)
	}

	var servcie_id, plan_id string

	//从etcd中取得参数。
	for i := 0; i < len(resp.Node.Nodes); i++ {
		if !resp.Node.Nodes[i].Dir {
			switch strings.ToLower(resp.Node.Nodes[i].Key) {
			case strings.ToLower(resp.Node.Key) + "/service_id":
				servcie_id = resp.Node.Nodes[i].Value
			case strings.ToLower(resp.Node.Key) + "/plan_id":
				plan_id = resp.Node.Nodes[i].Value
			}
		}
	}

	//并且要核对一下detail里面的service_id和plan_id。出错消息现在是500，需要更改一下源代码，以便更改出错代码
	if servcie_id != details.ServiceID || plan_id != details.PlanID {
		logger.Info("ServiceID or PlanID not correct!!")
		return errors.New("ServiceID or PlanID not correct!! instanceID " + instanceID)
	}

	//判断绑定是否存在，如果不存在就报错
	resp, err = etcdget("/servicebroker/" + servcieBrokerName + "/instance/" + instanceID + "/binding/" + bindingID)
	if err != nil || !resp.Node.Dir {
		logger.Error("Can not get binding information from etcd", err)
		return brokerapi.ErrBindingDoesNotExist //这几个错误返回为空，是detele操作的要求吗？
	} else {
		logger.Debug("Successful get bingding information from etcd. NodeInfo is " + resp.Node.Key)
	}

	//根据存储在etcd中的service_name和plan_name来确定到底调用那一段处理。注意这个时候不能像Provision一样去catalog里面读取了。
	//因为这个时候的数据不一定和创建的时候一样，plan等都有可能变化。同样的道理，url，用户名，密码都应该从_info中解码出来

	//隐藏属性不得不单独获取
	resp, err = etcdget("/servicebroker/" + servcieBrokerName + "/instance/" + instanceID + "/_info")
	json.Unmarshal([]byte(resp.Node.Value), &myServiceInfo)

	//隐藏属性不得不单独获取
	resp, err = etcdget("/servicebroker/" + servcieBrokerName + "/instance/" + instanceID + "/binding/" + bindingID + "/_info")
	json.Unmarshal([]byte(resp.Node.Value), &mycredentials)

	//do somthing 去删除用户名和密码
	switch myServiceInfo.Service_name {
	case managedServiceName:
		//由于bind的处理逻辑对于shared模式和standalone模式差不多
		var mongodburl string
		var mongodbname string
		//判断采用何种模式
		switch myServiceInfo.Plan_name {
		case "shared":
			//初始化mongodb的两个变量
			mongodburl = myServiceInfo.Url
			mongodbname = myServiceInfo.Database
		case "standalone":
			//初始化aws client
			svc := ec2.New(session.New(&aws.Config{Region: aws.String(awsRegion)}))
			//从service_instance的etcd目录中取出aws instance的id，以便查询
			awsInstaceID := myServiceInfo.Url
			//查询实例
			runResult, err := svc.DescribeInstances(&ec2.DescribeInstancesInput{
				InstanceIds: []*string{
					aws.String(awsInstaceID), // Required
					// More values...
				},
			})

			//出错后显示出错
			if err != nil {
				logger.Error("Could not get aws instance status "+awsInstaceID, err)
				return errors.New("Could not get aws instance status " + awsInstaceID)
			}

			//没有找到任何aws的实例
			if len(runResult.Reservations) == 0 {
				logger.Error("aws instance status not exist"+awsInstaceID, err)
				return errors.New("aws instance status not exist " + awsInstaceID)
			}

			switch *runResult.Reservations[0].Instances[0].State.Name {
			case "pending":
				//不可以执行的状态，反馈400，但是目前servie_broker的包里面没有这个状态的判断，只有暂时返回500 todo
				logger.Error("Another operation for this service instance is in progress", err)
				return errors.New("Another operation for this service instance is in progress")
			case "running":
				//可以解除绑定的状态
				//暂时利用公网地址，以后都部署在一个云平台上和一个安全域内部的话，可以使用私网地址
				mongodburl = *runResult.Reservations[0].Instances[0].PublicIpAddress + ":27017"
				//对于standalone的情况，应该创建一个dbadmin on any database的角色，所以数据库应该是admin
				mongodbname = "admin"
			default:
				//不可以执行的状态，反馈400，但是目前servie_broker的包里面没有这个状态的判断，只有暂时返回500 todo
				logger.Error("Another operation for this service instance is in progress", err)
				return errors.New("Another operation for this service instance is in progress")
			}
		default: //没有相关处理handle应该报错才对
			logger.Info("No Plan Handle for "+myServiceInfo.Service_name, nil)
			return errors.New("No Plan Handle for " + myServiceInfo.Service_name)
		}
		//初始化mongodb的链接串
		session, err := mgo.Dial(mongodburl) //连接数据库
		if err != nil {
			logger.Error("Can't connet to mongodb "+mongodburl, err)
			return errors.New("Can't connet to mongodb " + mongodburl)
		}
		defer session.Close()
		session.SetMode(mgo.Monotonic, true)
		mongodb := session.DB("admin") //数据库名称
		err = mongodb.Login(myServiceInfo.Admin_user, myServiceInfo.Admin_password)
		if err != nil {
			logger.Error("Can't Login to mongodb "+mongodburl, err)
			return errors.New("Can't Login to mongodb " + mongodburl)
		}

		//选择服务创建的数据库
		userdb := session.DB(mongodbname)
		//这个服务很快，所以通过同步模式直接返回了
		err = userdb.RemoveUser(mycredentials.Username)

		if err != nil {
			return errors.New("Can't DropUser in mongodb " + mongodburl)
			logger.Error("Can't DropUser in mongodb", err)
		} else {
			logger.Debug("Success DropUser in mongodb. user name=" + mycredentials.Username)
		}

	default: //没有相关的处理handle应该报错才对
		logger.Info("No Service Handle for " + myServiceInfo.Plan_name)
		return errors.New("No Service Handle for " + myServiceInfo.Plan_name)
	}

	//然后删除etcd里面的纪录，这里也有可能有不一致的情况
	_, err = etcdapi.Delete(context.Background(), "/servicebroker/"+servcieBrokerName+"/instance/"+instanceID+"/binding/"+bindingID, &client.DeleteOptions{Recursive: true, Dir: true}) //todo这些要么是常量，要么应该用环境变量
	if err != nil {
		logger.Error("Can not delete binding "+bindingID+" in etcd", err) //todo都应该改为日志key
		return errors.New("Can not delete binding " + bindingID + " in etcd")
	} else {
		logger.Debug("Successful delete binding "+bindingID+" in etcd", nil)
	}

	logger.Info("Successful delete binding "+bindingID, nil)
	return nil
}

func (myBroker *myServiceBroker) Update(instanceID string, details brokerapi.UpdateDetails, asyncAllowed bool) (brokerapi.IsAsync, error) {
	// Update instance here
	return brokerapi.IsAsync(true), nil
}

//定义工具函数
func etcdget(key string) (*client.Response, error) {
	resp, err := etcdapi.Get(context.Background(), key, nil)
	if err != nil {
		logger.Error("Can not get "+key+" from etcd", err)
	} else {
		logger.Debug("Successful get " + key + " from etcd. value is " + resp.Node.Value)
	}
	return resp, err
}

func etcdset(key string, value string) (*client.Response, error) {
	resp, err := etcdapi.Set(context.Background(), key, value, nil)
	if err != nil {
		logger.Error("Can not set "+key+" from etcd", err)
	} else {
		logger.Debug("Successful set " + key + " from etcd. value is " + value)
	}
	return resp, err
}

func findServiceNameInCatalog(service_id string) string {
	resp, err := etcdget("/servicebroker/" + servcieBrokerName + "/catalog/" + service_id + "/name")
	if err != nil {
		return ""
	}
	return resp.Node.Value
}

func findServicePlanNameInCatalog(service_id, plan_id string) string {
	resp, err := etcdget("/servicebroker/" + servcieBrokerName + "/catalog/" + service_id + "/plan/" + plan_id + "/name")
	if err != nil {
		return ""
	}
	return resp.Node.Value
}

func getmd5string(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

func getguid() string {
	b := make([]byte, 48)

	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return getmd5string(base64.URLEncoding.EncodeToString(b))
}

func getenv(env string) string {
	env_value := os.Getenv(env)
	if env_value == "" {
		fmt.Println("FATAL: NEED ENV", env)
		fmt.Println("Exit...........")
		os.Exit(2)
	}
	fmt.Println("ENV:", env, env_value)
	return env_value
}

//定义日志和etcd的全局变量，以及其他变量
var logger lager.Logger
var etcdapi client.KeysAPI
var servcieBrokerName string = "mongodb_aws"
var etcdEndPoint string
var serviceBrokerPort string
var mongoUrl string
var mongoAdminUser string
var mongoAdminPassword string
var managedServiceName string = "mongodb_aws" //管理的服务名，注意用这种方法，一个service broker就只能管理一个服务。后面再考虑重构
var awsRegion string = "cn-north-1"
var imageId string = "ami-b18942dc"
var instanceType string = "t2.micro"
var keyName string = "service_borker"
var securityGroups string = "service borker"

func main() {
	//初始化参数，参数应该从环境变量中获取
	var username, password string
	//todo参数应该改为从环境变量中获取
	//需要以下环境变量
	etcdEndPoint = getenv("ETCDENDPOINT")             //etcd的路径
	serviceBrokerPort = getenv("BROKERPORT")          //监听的端口
	mongoUrl = getenv("MONGOURL")                     //共享实例的mongodb地址
	mongoAdminUser = getenv("MONGOADMINUSER")         //共享实例和独立实例的管理员用户名
	mongoAdminPassword = getenv("MONGOADMINPASSWORD") //共享实例和独立实例的管理员密码

	/* 环境变量案例

	export ETCDENDPOINT="http://192.168.99.100:2379"
	export BROKERPORT="8000"
	export MONGOURL="54.222.155.67:27017"
	export MONGOADMINUSER="asiainfoLDP"
	export MONGOADMINPASSWORD="6ED9BA74-75FD-4D1B-8916-842CB936AC1A"


	//aws客户端还需要额外两个环境变量
	export AWS_ACCESS_KEY_ID=AKIAO2SO52RKIE7BCSHA
	export AWS_SECRET_ACCESS_KEY=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
	*/

	//初始化日志对象，日志输出到stdout
	logger = lager.NewLogger(servcieBrokerName)
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.INFO)) //默认日志级别

	//初始化etcd客户端
	cfg := client.Config{
		Endpoints: []string{etcdEndPoint},
		Transport: client.DefaultTransport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second,
	}
	c, err := client.New(cfg)
	if err != nil {
		logger.Error("Can not init ectd client", err)
	}
	etcdapi = client.NewKeysAPI(c)

	//初始化serviceborker对象
	serviceBroker := &myServiceBroker{}

	//取得用户名和密码
	resp, err := etcdget("/servicebroker/" + servcieBrokerName + "/username")
	if err != nil {
		logger.Error("Can not init username,Progrom Exit!", err)
		os.Exit(1)
	} else {
		username = resp.Node.Value
	}

	resp, err = etcdget("/servicebroker/" + servcieBrokerName + "/password")
	if err != nil {
		logger.Error("Can not init password,Progrom Exit!", err)
		os.Exit(1)
	} else {
		password = resp.Node.Value
	}

	//装配用户名和密码
	credentials := brokerapi.BrokerCredentials{
		Username: username,
		Password: password,
	}

	fmt.Println("START SERVICE BROKER", servcieBrokerName)
	brokerAPI := brokerapi.New(serviceBroker, logger, credentials)
	http.Handle("/", brokerAPI)
	http.ListenAndServe(":"+serviceBrokerPort, nil)
}
