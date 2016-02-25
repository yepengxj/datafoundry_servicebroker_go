package main

import (
    "github.com/pivotal-cf/brokerapi"
    "github.com/pivotal-golang/lager"
    "net/http"
    "fmt"
    "os"
    "time"
    "github.com/coreos/etcd/Godeps/_workspace/src/golang.org/x/net/context"
    "github.com/coreos/etcd/client"
    "encoding/json"
    "strconv"
    "strings"
    "errors"
    "gopkg.in/mgo.v2"
    "crypto/md5"
    "crypto/rand"
    "encoding/base64"
    "encoding/hex"
    "io"
)



type myServiceBroker struct {

}

type serviceInfo struct {
	Service_name 	string 		`json:"service_name"`
	Plan_name 		string 		`json:"plan_name"`
	Url				string 		`json:"url"`
	Admin_user		string 		`json:"admin_user"`
	Admin_password	string 		`json:"admin_password"`
	Database		string 		`json:"database"`
	User			string 		`json:"user"`
	Password		string 		`json:"password"`
}

//type planInfo struct {
//
//}

func (myBroker *myServiceBroker) Services() []brokerapi.Service {
    //初始化一系列所需要的结构体，好累啊
    myServices:=[]brokerapi.Service{}
    myService:=brokerapi.Service{}
    myPlans:=[]brokerapi.ServicePlan{}
    myPlan:=brokerapi.ServicePlan{}
    var myPlanfree bool
    //todo还需要考虑对于service和plan的隐藏参数，status，比如可以用，不可用，已经删除等。删除应该是软删除，后两者不予以显示，前者表示还有数据
    //获取catalog信息
    resp, err := etcdapi.Get(context.Background(), "/servicebroker/"+"mongodb_aws"+"/catalog", &client.GetOptions{Recursive:true}) //改为环境变量
    if err!=nil {
        logger.Error("Can not get catalog information from etcd", err) //所有这些出错消息最好命名为常量，放到开始的时候
    } else {
        logger.Debug("Successful get catalog information from etcd. NodeInfo is "+resp.Node.Key)
    }

    for i := 0; i < len(resp.Node.Nodes); i++ {
        //为旗下发现的每一个service进行迭代，不过一般情况下，应该只有一个service
        logger.Debug("Start to Parse Service "+resp.Node.Nodes[i].Key)
        //在下一级循环外设置id，因为他是目录名字，注意，如果按照这个逻辑，id一定要是uuid，中间一定不能有目录符号"/"
        myService.ID=strings.Split(resp.Node.Nodes[i].Key,"/")[len(strings.Split(resp.Node.Nodes[i].Key,"/"))-1]
        //开始取service级别除了ID以外的其他参数
        for j :=0 ;j<len(resp.Node.Nodes[i].Nodes); j++ {
            if ! resp.Node.Nodes[i].Nodes[j].Dir {
                switch strings.ToLower(resp.Node.Nodes[i].Nodes[j].Key) {
                    case strings.ToLower(resp.Node.Nodes[i].Key)+"/name":
                        myService.Name=resp.Node.Nodes[i].Nodes[j].Value 
                    case strings.ToLower(resp.Node.Nodes[i].Key)+"/description":
                        myService.Description=resp.Node.Nodes[i].Nodes[j].Value 
                    case strings.ToLower(resp.Node.Nodes[i].Key)+"/bindable":
                        myService.Bindable,_=strconv.ParseBool(resp.Node.Nodes[i].Nodes[j].Value)
                    case strings.ToLower(resp.Node.Nodes[i].Key)+"/tags":
                        myService.Tags=strings.Split(resp.Node.Nodes[i].Nodes[j].Value,",")
                    case strings.ToLower(resp.Node.Nodes[i].Key)+"/planupdatable":
                        myService.PlanUpdatable,_=strconv.ParseBool(resp.Node.Nodes[i].Nodes[j].Value)
                    case strings.ToLower(resp.Node.Nodes[i].Key)+"/metadata":
                        json.Unmarshal([]byte(resp.Node.Nodes[i].Nodes[j].Value),&myService.Metadata)     
                }  
            } else if strings.HasSuffix(strings.ToLower(resp.Node.Nodes[i].Nodes[j].Key),"plan") {
                //开始解析套餐目录中的套餐计划plan。上述判断也不是太严谨，比如有目录如果是xxxxplan怎么办？
                for k:=0 ;k <len(resp.Node.Nodes[i].Nodes[j].Nodes); k++ {
                    logger.Debug("Start to Parse Plan "+resp.Node.Nodes[i].Nodes[j].Nodes[k].Key)
                    myPlan.ID=strings.Split(resp.Node.Nodes[i].Nodes[j].Nodes[k].Key,"/")[len(strings.Split(resp.Node.Nodes[i].Nodes[j].Nodes[k].Key,"/"))-1]
                    for n:=0 ; n < len(resp.Node.Nodes[i].Nodes[j].Nodes[k].Nodes);n++ {
                        switch strings.ToLower(resp.Node.Nodes[i].Nodes[j].Nodes[k].Nodes[n].Key) {
                            case strings.ToLower(resp.Node.Nodes[i].Nodes[j].Nodes[k].Key)+"/name":
                                myPlan.Name=resp.Node.Nodes[i].Nodes[j].Nodes[k].Nodes[n].Value
                            case strings.ToLower(resp.Node.Nodes[i].Nodes[j].Nodes[k].Key)+"/description":
                                myPlan.Description=resp.Node.Nodes[i].Nodes[j].Nodes[k].Nodes[n].Value
                            case strings.ToLower(resp.Node.Nodes[i].Nodes[j].Nodes[k].Key)+"/free":
                                //这里没有搞懂为什么brokerapi里面的这个bool要定义为传指针的模式
                                myPlanfree,_=strconv.ParseBool(resp.Node.Nodes[i].Nodes[j].Nodes[k].Nodes[n].Value)
                                myPlan.Free=brokerapi.FreeValue(myPlanfree)
                            case strings.ToLower(resp.Node.Nodes[i].Nodes[j].Nodes[k].Key)+"/metadata":
                                json.Unmarshal([]byte(resp.Node.Nodes[i].Nodes[j].Nodes[k].Nodes[n].Value),&myPlan.Metadata) 
                        }
                    }
                    //装配plan需要返回的值，按照有多少个plan往里面装
                    myPlans=append(myPlans,myPlan)
                    //重置服务变量
                    myPlan=brokerapi.ServicePlan{}

                }
                //将装配好的Plan对象赋值给Service               
                myService.Plans=myPlans
            }
        }
       
        //装配catalog需要返回的值，按照有多少个服务往里面装
        myServices=append(myServices,myService)
        //重置服务变量
        myService=brokerapi.Service{}


    }
    
    return myServices

}

func (myBroker *myServiceBroker) Provision(
	instanceID string,
	details brokerapi.ProvisionDetails,
	asyncAllowed bool,
) (brokerapi.ProvisionedServiceSpec, error) {
	// Provision a new instance here. If async is allowed, the broker can still
	// chose to provision the instance synchronously.
	//初始化
	var DashboardURL string
	var myServiceInfo serviceInfo

	//判断实例是否已经存在，如果存在就报错
	resp, err := etcdget("/servicebroker/"+"mongodb_aws"+"/instance") //改为环境变量
    for i :=0 ;i<len(resp.Node.Nodes); i++ {
        if resp.Node.Nodes[i].Dir && strings.HasSuffix(resp.Node.Nodes[i].Key,instanceID) {
        	return brokerapi.ProvisionedServiceSpec{}, brokerapi.ErrInstanceAlreadyExists
        }
    }

	//判断servcie_id和plan_id是否正确
	service_name:=findServiceNameInCatalog(details.ServiceID)
	plan_name:=findServicePlanNameInCatalog(details.ServiceID,details.PlanID)
	if service_name=="" || plan_name=="" {
		return brokerapi.ProvisionedServiceSpec{}, errors.New("Service_id or plan_id not correct!!")
	}
	//是否要检查service和plan的status是否允许创建 todo

	//根据不同的服务和plan，选择创建的命令 ［每次增加不同的服务或者计划，只需要修改这里就好了。］
	switch service_name {
		case "mongodb_aws" : //需要配置为Service的环境变量
			//开始根据不同的plan进行处理
			switch plan_name {
				case "shared" :
					//初始化mongodb的链接串
					var		mongoURL = "192.168.99.100:32768" //修改为环境变量获取
					var		mongoADMINUSER="asiainfoLDP"
					var		mongoADMINPASSWORD="6ED9BA74-75FD-4D1B-8916-842CB936AC1A"
					session, err := mgo.Dial(mongoURL)  //连接数据库
  					if err != nil {
    					return brokerapi.ProvisionedServiceSpec{}, errors.New("Can't connet to mongodb "+mongoURL)
  					}
  					defer session.Close()
  					session.SetMode(mgo.Monotonic, true)
  					mongodb := session.DB("admin")	 //数据库名称
  					err = mongodb.Login(mongoADMINUSER,mongoADMINPASSWORD) 
  					if err != nil {
    					return brokerapi.ProvisionedServiceSpec{}, errors.New("Can't Login to mongodb "+mongoURL)
  					}

  					//创建一个名为instanceID的数据库，并随机的创建用户名和密码，这个用户名是该数据库的管理员
  					newdb:=session.DB(instanceID)
  					newusername:=getguid()
  					newpassword:=getguid()
  					//为dashbord赋值 todo dashboard应该提供一个界面才对
  					DashboardURL="mongodb://"+newusername+":"+newpassword+"@"+mongoURL+"/"+instanceID
  					//这个服务很快，所以通过同步模式直接返回了
  					err=newdb.UpsertUser(&mgo.User{
  							Username:	newusername,
  							Password:	newpassword,
  							Roles: []mgo.Role{
  								mgo.Role(mgo.RoleDBAdmin),
  							},
  						})

  					if err != nil {
    					return brokerapi.ProvisionedServiceSpec{}, errors.New("Can't Create User in mongodb "+mongoURL)
    					logger.Error("Can't Create User in mongodb", err)
  					} else {
  						logger.Debug("Success Create User in mongodb. Username="+newusername+" Password="+newpassword)
  					}

  					//赋值隐藏属性
					myServiceInfo=serviceInfo{
						Service_name:service_name,
						Plan_name:plan_name,
						Url:mongoURL,
						Admin_user:mongoADMINUSER,
						Admin_password:mongoADMINPASSWORD,
						Database:instanceID,
						User:newusername,
						Password:newpassword,
					}


				default : //没有相关处理handle应该报错才对
					return brokerapi.ProvisionedServiceSpec{}, errors.New("No Plan Handle for "+plan_name)

			}
		default : //没有相关的处理handle应该报错才对
			return brokerapi.ProvisionedServiceSpec{}, errors.New("No Service Handle for "+service_name)
	}

	//假设服务已经生成好了，用来做测试 
	

	//写入etcd 话说如果这个时候写入失败，那不就出现数据不一致的情况了么！todo
	//先创建instanceid目录
	_, err = etcdapi.Set(context.Background(), "/servicebroker/"+"mongodb_aws"+"/instance/"+instanceID,"", &client.SetOptions{Dir:true}) //todo这些要么是常量，要么应该用环境变量
	if err != nil {
		logger.Error("Can not create instance "+instanceID+" in etcd", err) //todo都应该改为日志key
		return brokerapi.ProvisionedServiceSpec{}, err
	} else {
		logger.Debug("Successful create instance "+instanceID+" in etcd",nil)
	}
	//然后创建一系列属性
	etcdset("/servicebroker/"+"mongodb_aws"+"/instance/"+instanceID+"/organization_guid",details.OrganizationGUID)
	etcdset("/servicebroker/"+"mongodb_aws"+"/instance/"+instanceID+"/space_guid",details.SpaceGUID)
	etcdset("/servicebroker/"+"mongodb_aws"+"/instance/"+instanceID+"/service_id",details.ServiceID)
	etcdset("/servicebroker/"+"mongodb_aws"+"/instance/"+instanceID+"/plan_id",details.PlanID)
	tmpval,_:=json.Marshal(details.Parameters)
	etcdset("/servicebroker/"+"mongodb_aws"+"/instance/"+instanceID+"/parameters",string(tmpval))
	etcdset("/servicebroker/"+"mongodb_aws"+"/instance/"+instanceID+"/dashboardurl",DashboardURL)
	//存储隐藏信息_info
	tmpval,_=json.Marshal(myServiceInfo)
	etcdset("/servicebroker/"+"mongodb_aws"+"/instance/"+instanceID+"/_info",string(tmpval))

	//创建绑定目录
	_, err = etcdapi.Set(context.Background(), "/servicebroker/"+"mongodb_aws"+"/instance/"+instanceID+"/binding","", &client.SetOptions{Dir:true})

	//完成所有操作后，返回DashboardURL
	return brokerapi.ProvisionedServiceSpec{DashboardURL: DashboardURL}, nil
}

func (myBroker *myServiceBroker) LastOperation(instanceID string) (brokerapi.LastOperation, error) {
	// If the broker provisions asynchronously, the Cloud Controller will poll this endpoint
	// for the status of the provisioning operation.

	//去读取进展状态，如果有错误，返回错误，如果没有错误，返回对象LastOperation! todo 如果同步模式，不用实现这个接口
	return brokerapi.LastOperation{}, nil
}

func (myBroker *myServiceBroker) Deprovision(instanceID string, details brokerapi.DeprovisionDetails, asyncAllowed bool) (brokerapi.IsAsync, error) {
	// Deprovision a new instance here. If async is allowed, the broker can still
	// chose to deprovision the instance synchronously, hence the first return value.

	var myServiceInfo serviceInfo

	//判断实例是否已经存在，如果不存在就报错
	resp, err := etcdapi.Get(context.Background(), "/servicebroker/"+"mongodb_aws"+"/instance/"+instanceID, &client.GetOptions{Recursive:true}) //改为环境变量
    
    if err!=nil || !resp.Node.Dir {
        logger.Error("Can not get instance information from etcd", err) //所有这些出错消息最好命名为常量，放到开始的时候
        return brokerapi.IsAsync(false), brokerapi.ErrInstanceDoesNotExist
    } else {
        logger.Debug("Successful get instance information from etcd. NodeInfo is "+resp.Node.Key)
    }

    var servcie_id,plan_id string

	//从etcd中取得参数。
	for i := 0; i < len(resp.Node.Nodes); i++ {
		if ! resp.Node.Nodes[i].Dir {
			switch strings.ToLower(resp.Node.Nodes[i].Key) {
				case strings.ToLower(resp.Node.Key)+"/service_id":
					servcie_id=resp.Node.Nodes[i].Value
				case strings.ToLower(resp.Node.Key)+"/plan_id":
					plan_id=resp.Node.Nodes[i].Value
			}
		}
	}

	//并且要核对一下detail里面的service_id和plan_id。出错消息现在是500，需要更改一下源代码，以便更改出错代码
	if servcie_id!=details.ServiceID || plan_id!=details.PlanID {
		logger.Error("ServiceID or PlanID not correct!!", nil) //所有这些出错消息最好命名为常量，放到开始的时候
    	return brokerapi.IsAsync(false), errors.New("ServiceID or PlanID not correct!! instanceID "+instanceID)
	}	
    //是否要判断里面有没有绑定啊？todo

	//根据存储在etcd中的service_name和plan_name来确定到底调用那一段处理。注意这个时候不能像Provision一样去catalog里面读取了。
	//因为这个时候的数据不一定和创建的时候一样，plan等都有可能变化。同样的道理，url，用户名，密码都应该从_info中解码出来

	//隐藏属性不得不单独获取
	resp, err=etcdget("/servicebroker/"+"mongodb_aws"+"/instance/"+instanceID+"/_info")
	json.Unmarshal([]byte(resp.Node.Value),&myServiceInfo) 

	//根据不同的服务和plan，选择创建的命令 ［每次增加不同的服务或者计划，只需要修改这里就好了。］
	switch myServiceInfo.Service_name {
		case "mongodb_aws" : 
			switch myServiceInfo.Plan_name {
				case "shared" :
					//初始化mongodb的链接串
					session, err := mgo.Dial(myServiceInfo.Url)  //连接数据库
  					if err != nil {
    					return brokerapi.IsAsync(false), errors.New("Can't connet to mongodb "+myServiceInfo.Url)
  					}
  					defer session.Close()
  					session.SetMode(mgo.Monotonic, true)
  					mongodb := session.DB("admin")	 //数据库名称
  					err = mongodb.Login(myServiceInfo.Admin_user,myServiceInfo.Admin_password) 
  					if err != nil {
    					return brokerapi.IsAsync(false), errors.New("Can't Login to mongodb "+myServiceInfo.Url)
  					}

  					//创建一个名为instanceID的数据库，并随机的创建用户名和密码，这个用户名是该数据库的管理员
  					userdb:=session.DB(myServiceInfo.Database)
  					//这个服务很快，所以通过同步模式直接返回了
  					err=userdb.DropDatabase()

  					if err != nil {
    					return brokerapi.IsAsync(false), errors.New("Can't DropDatabase in mongodb "+myServiceInfo.Url)
    					logger.Error("Can't DropDatabase in mongodb", err)
  					} else {
  						logger.Debug("Success DropDatabase in mongodb. database name="+myServiceInfo.Database)
  					}


				default : //没有相关处理handle应该报错才对
					return brokerapi.IsAsync(false), errors.New("No Plan Handle for "+myServiceInfo.Service_name)
			}
		default : //没有相关的处理handle应该报错才对
			return brokerapi.IsAsync(false), errors.New("No Service Handle for "+myServiceInfo.Plan_name)
	}

	//然后删除etcd里面的纪录，这里也有可能有不一致的情况
	_, err = etcdapi.Delete(context.Background(), "/servicebroker/"+"mongodb_aws"+"/instance/"+instanceID, &client.DeleteOptions{Recursive:true,Dir:true}) //todo这些要么是常量，要么应该用环境变量
	if err != nil {
		logger.Error("Can not delete instance "+instanceID+" in etcd", err) //todo都应该改为日志key
		return brokerapi.IsAsync(false), errors.New("Internal Error!!")
	} else {
		logger.Debug("Successful delete instance "+instanceID+" in etcd",nil)
	}

	return brokerapi.IsAsync(false), nil
}

func (myBroker *myServiceBroker) Bind(instanceID, bindingID string, details brokerapi.BindDetails) (brokerapi.Binding, error) {
	// Bind to instances here
	// Return a binding which contains a credentials object that can be marshalled to JSON,
	// and (optionally) a syslog drain URL.

	//判断实例是否已经存在，如果不存在就报错
	resp, err := etcdget("/servicebroker/"+"mongodb_aws"+"/instance/"+instanceID) //改为环境变量
    if err!=nil || !resp.Node.Dir {
        logger.Error("Can not get instance information from etcd", err) //所有这些出错消息最好命名为常量，放到开始的时候
        return brokerapi.Binding{}, brokerapi.ErrInstanceDoesNotExist
    } else {
        logger.Debug("Successful get instance information from etcd. NodeInfo is "+resp.Node.Key)
    }

    //判断绑定是否存在，如果存在就报错
    resp, err = etcdget("/servicebroker/"+"mongodb_aws"+"/instance/"+instanceID+"/binding") //改为环境变量
    for i :=0 ;i<len(resp.Node.Nodes); i++ {
        if resp.Node.Nodes[i].Dir && strings.HasSuffix(resp.Node.Nodes[i].Key,bindingID) {
        	return brokerapi.Binding{}, brokerapi.ErrBindingAlreadyExists
        }
    }


    //对于参数中的service_id和plan_id仅做校验，不再在binding中存储 todo需要将原来的包borkerapi进行修改，增加不匹配service_id和plan_id的错误类型
    //do something 来在实例中创建用户 todo
    type myCredentials struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	}

    myBinding:=brokerapi.Binding{
		Credentials: myCredentials{
			Host:     "127.0.0.1",
			Port:     3000,
			Username: "batman",
			Password: "robin",
		},
		SyslogDrainURL: "",
	}


    //把信息存储到etcd里面，同样这里有同步性的问题 todo怎么解决呢？
    //先创建bindingID目录
	_, err = etcdapi.Set(context.Background(), "/servicebroker/"+"mongodb_aws"+"/instance/"+instanceID+"/binding/"+bindingID,"", &client.SetOptions{Dir:true}) //todo这些要么是常量，要么应该用环境变量
	if err != nil {
		logger.Error("Can not create binding "+bindingID+" in etcd", err) //todo都应该改为日志key
		return brokerapi.Binding{}, err
	} else {
		logger.Debug("Successful create binding "+bindingID+" in etcd",nil)
	}
	//然后创建一系列属性
	etcdset("/servicebroker/"+"mongodb_aws"+"/instance/"+instanceID+"/binding/"+bindingID+"/app_guid",details.AppGUID)
	tmpval,_:=json.Marshal(details.Parameters)
	etcdset("/servicebroker/"+"mongodb_aws"+"/instance/"+instanceID+"/binding/"+bindingID+"/parameters",string(tmpval))
	tmpval,_=json.Marshal(myBinding)
	etcdset("/servicebroker/"+"mongodb_aws"+"/instance/"+instanceID+"/binding/"+bindingID+"/binding",string(tmpval))
	//应该具体绑定创建的时候还有一些信息要存储，这样在服务绑定的时候，就可以读取这些信息来反馈! todo

	return myBinding, nil
}

func (myBroker *myServiceBroker) Unbind(instanceID, bindingID string, details brokerapi.UnbindDetails) error {
	// Unbind from instances here
	
	//判断实例是否已经存在，如果不存在就报错
	resp, err := etcdget("/servicebroker/"+"mongodb_aws"+"/instance/"+instanceID) //改为环境变量
    if err!=nil || !resp.Node.Dir {
        logger.Error("Can not get instance information from etcd", err) //所有这些出错消息最好命名为常量，放到开始的时候
        return brokerapi.ErrInstanceDoesNotExist //这几个错误返回为空，是detele操作的要求吗？
    } else {
        logger.Debug("Successful get instance information from etcd. NodeInfo is "+resp.Node.Key)
    }

    //判断绑定是否存在，如果不存在就报错
    resp, err = etcdget("/servicebroker/"+"mongodb_aws"+"/instance/"+instanceID+"/binding/"+bindingID) //改为环境变量
    if err!=nil || !resp.Node.Dir {
        logger.Error("Can not get binding information from etcd", err) //所有这些出错消息最好命名为常量，放到开始的时候
        return brokerapi.ErrBindingDoesNotExist //这几个错误返回为空，是detele操作的要求吗？
    } else {
        logger.Debug("Successful get bingding information from etcd. NodeInfo is "+resp.Node.Key)
    }

    //double check service_id和plan_id

    //do somthing 去删除用户名和密码 todo

	//然后删除etcd里面的纪录，这里也有可能有不一致的情况
	_, err = etcdapi.Delete(context.Background(), "/servicebroker/"+"mongodb_aws"+"/instance/"+instanceID+"/binding/"+bindingID, &client.DeleteOptions{Recursive:true,Dir:true}) //todo这些要么是常量，要么应该用环境变量
	if err != nil {
		logger.Error("Can not delete binding "+bindingID+" in etcd", err) //todo都应该改为日志key
		return errors.New("Internal Error!!")
	} else {
		logger.Debug("Successful delete bingding "+bindingID+" in etcd",nil)
	}

	return nil
}

func (myBroker *myServiceBroker) Update(instanceID string, details brokerapi.UpdateDetails, asyncAllowed bool) (brokerapi.IsAsync, error) {
	// Update instance here
	return brokerapi.IsAsync(true), nil
}

//定义日志和etcd的全局变量
var logger lager.Logger
var etcdapi client.KeysAPI

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

func etcdset(key string,value string) (*client.Response, error) {
	resp, err := etcdapi.Set(context.Background(), key,value, nil)
	if err != nil {
		logger.Error("Can not set "+key+" from etcd", err)
	} else {
		logger.Debug("Successful set " + key + " from etcd. value is " + value)
	}
	return resp, err
}

func findServiceNameInCatalog(service_id string) string {
	resp, err := etcdget("/servicebroker/" + "mongodb_aws" + "/catalog/"+service_id+"/name") //需要修改为环境变量参数
	if err !=nil {
		return ""
	}
	return resp.Node.Value
}

func findServicePlanNameInCatalog(service_id,plan_id string) string {
	resp, err := etcdget("/servicebroker/" + "mongodb_aws" + "/catalog/"+service_id+"/plan/"+plan_id+"/name") //需要修改为环境变量参数
	if err !=nil {
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

func main() {
	//初始化参数，参数应该从环境变量中获取
	var username, password string
	//todo参数应该改为从环境变量中获取

	//初始化日志对象，日志输出到stdout
	logger = lager.NewLogger("mongodb_aws")                          //替换为环境变量
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.ERROR)) //默认日志级别

	//初始化etcd客户端
	cfg := client.Config{
		Endpoints: []string{"http://192.168.99.100:2379"}, //替换为环境变量
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
	resp, err := etcdget("/servicebroker/" + "mongodb_aws" + "/username") //需要修改为环境变量参数
	if err != nil {
		logger.Error("Can not init username,Progrom Exit!", err)
		os.Exit(1)
	} else {
		username = resp.Node.Value
	}

	resp, err = etcdget("/servicebroker/" + "mongodb_aws" + "/password") //需要修改为环境变量参数
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

	fmt.Println("start http server")
	brokerAPI := brokerapi.New(serviceBroker, logger, credentials)
	http.Handle("/", brokerAPI)
	http.ListenAndServe(":8000", nil) //需要修改为环境变量
}
