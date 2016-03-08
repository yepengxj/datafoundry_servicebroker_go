package handler

import "testing"
import "github.com/pivotal-cf/brokerapi"
import "fmt"

func TestBroker(t *testing.T) {

	test := &Mysql_sharedHandler{}
	detail := brokerapi.ProvisionDetails{}
	_, _, err := test.DoProvision("common", detail, false)
	fmt.Println(err)
}
