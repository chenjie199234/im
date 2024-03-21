package dao

import (
	"github.com/chenjie199234/im/config"
	// "github.com/chenjie199234/im/model"
	// discoversdk "github.com/chenjie199234/admin/sdk/discover"
	// "github.com/chenjie199234/Corelib/discover"
	// "github.com/chenjie199234/Corelib/cgrpc"
	// "github.com/chenjie199234/Corelib/crpc"
	// "github.com/chenjie199234/Corelib/web"
)

//var ExampleCGrpcApi example.ExampleCGrpcClient
//var ExampleCrpcApi example.ExampleCrpcClient
//var ExampleWebApi  example.ExampleWebClient

// NewApi create all dependent service's api we need in this program
func NewApi() error {
	//init dns discover for example server
	//exampleDnsDiscover, e := discover.NewDNSDiscover("exampleproject", "examplegroup", "examplename", "dnshost", time.Second*10, 9000, 10000, 8000)
	//if e != nil {
	//	return e
	//}
	//
	//init static discover for example server
	//exampleStaticDiscover, e := discover.NewStaticDiscover("exampleproject", "examplegroup", "examplename", []string{"addr1","addr2"}, 9000, 10000, 8000)
	//if e != nil {
	//	return e
	//}
	//
	//init kubernetes discover for example server
	//exampleKubeDiscover, e := discover.NewKubernetesDiscover("exampleproject", "examplegroup", "examplename", "namespace", "fieldselector", "labelselector", 9000, 10000, 8000)
	//if e != nil {
	//	return e
	//}
	//
	//init admin discover for example server
	//if admin service needs tls,you need to specific the config
	//var admintlsc *tlsc.Config{}
	//if adminNeedTLS {
	//	admintlsc = &tlsc.Config{}
	//	...
	//}
	//exampleAdminDiscover, e := discoversdk.NewAdminDiscover(model.Project, model.Group, model.Name, "exampleproject", "examplegroup", "examplename", admintlsc)
	//if e != nil {
	//	return e
	//}

	//if example service needs tls,you need to specific the config
	//var exampletlsc *tls.Config
	// if exampleNeedTLS {
	// 	exampletlsc = &tls.Config{}
	// 	...
	// }

	cgrpcc := config.GetCGrpcClientConfig().ClientConfig
	_ = cgrpcc //avoid unuse

	//init cgrpc client below
	//examplecgrpc, e := cgrpc.NewCGrpcClient(cgrpcc, examplediscover, model.Project, model.Group, model.Name, "exampleproject", "examplegroup", "examplename", exampletlsc)
	//if e != nil {
	//         return e
	//}
	//ExampleCGrpcApi = example.NewExampleCGrpcClient(examplecgrpc)

	crpcc := config.GetCrpcClientConfig().ClientConfig
	_ = crpcc //avoid unuse

	//init crpc client below
	//examplecrpc, e := crpc.NewCrpcClient(crpcc, examplediscover, model.Project, model.Group, model.Name, "exampleproject", "examplegroup", "examplename", exampletlsc)
	//if e != nil {
	// 	return e
	//}
	//ExampleCrpcApi = example.NewExampleCrpcClient(examplecrpc)

	webc := config.GetWebClientConfig().ClientConfig
	_ = webc //avoid unuse

	//init web client below
	//exampleweb, e := web.NewWebClient(webc, examplediscover, model.Project, model.Group, model.Name, "exampleproject", "examplegroup", "examplename", exampletlsc)
	//if e != nil {
	// 	return e
	//}
	//ExampleWebApi = example.NewExampleWebClient(exampleweb)

	return nil
}

func UpdateAppConfig(ac *config.AppConfig) {

}
