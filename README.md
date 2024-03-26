# im
```
im是一个微服务.
运行cmd脚本可查看使用方法.windows下将./cmd.sh换为cmd.bat
./cmd.sh help 输出帮助信息
./cmd.sh pb 解析proto文件,生成桩代码
./cmd.sh sub 在该项目中创建一个新的子服务
./cmd.sh kube 新建kubernetes的配置
./cmd.sh html 新建前端html代码模版
```

## 服务端口
```
6060                                    MONITOR AND PPROF
8000                                    WEB
9000                                    CRPC
10000                                   GRPC
8080                                    RawTcp or WebSocket
```

## 环境变量
```
LOG_LEVEL                               日志等级,debug,info(default),warn,error
LOG_TRACE                               是否开启链路追踪,1-开启,0-关闭(default)
LOG_TARGET                              日志输出目标,std-输出到标准输出,file-输出到文件(当前工作目录的./log/目录下)
PROJECT                                 该项目所属的项目,[a-z][0-9],第一个字符必须是[a-z]
GROUP                                   该项目所属的项目下的小组,[a-z][0-9],第一个字符必须是[a-z]
RUN_ENV                                 当前运行环境,如:test,pre,prod
DEPLOY_ENV                              部署环境,如:ali-kube-shanghai-1,ali-host-hangzhou-1
MONITOR                                 是否开启系统监控采集,0关闭,1开启

CONFIG_TYPE                             配置类型:0-使用本地配置.1-使用admin服务的远程配置中心功能
REMOTE_CONFIG_SECRET                    当CONFIG_TYPE为1时,admin服务中,该服务使用的配置加密密钥,最长31个字符
ADMIN_SERVICE_PROJECT                   当使用admin服务的远程配置中心,服务发现,权限管理功能时,需要设置该环境变量,该变量为admin服务所属的项目,[a-z][0-9],第一个字符必须是[a-z]
ADMIN_SERVICE_GROUP                     当使用admin服务的远程配置中心,服务发现,权限管理功能时,需要设置该环境变量,该变量为admin服务所属的项目下的小组,[a-z][0-9],第一个字符必须是[a-z]
ADMIN_SERVICE_WEB_HOST                  当使用admin服务的远程配置中心,服务发现,权限管理功能时,需要设置该环境变量,该变量为admin服务的host,不带scheme(tls取决于NewSdk时是否传入tls.Config)
ADMIN_SERVICE_WEB_PORT                  当使用admin服务的远程配置中心,服务发现,权限管理功能时,需要设置该环境变量,该变量为admin服务的web端口,默认为80/443(取决于NewSdk时是否使用tls)
ADMIN_SERVICE_CONFIG_ACCESS_KEY         当使用admin服务的远程配置中心功能时,admin服务的授权码
ADMIN_SERVICE_DISCOVER_ACCESS_KEY       当使用admin服务的服务发现功能时,admin服务的授权码
ADMIN_SERVICE_PERMISSION_ACCESS_KEY     当使用admin服务的权限控制功能时,admin服务的授权码
```

## 配置文件
```
AppConfig.json该文件配置了该服务需要使用的业务配置,可热更新
SourceConfig.json该文件配置了该服务需要使用的资源配置,不热更新
```

## Cache
### Redis(Version >= 7.0)
#### gate_redis(Cluster mode is better)
#### im_redis(Cluster mode is better)

## DB
### Mongo(Shard Mode)(Version >= 6.0)
#### im
```
database: im

collection: msg
{
    _id:ObjectId("xxx"),//provide timestamp
    chat_key:"",//user: join(sort(sender,target),"-"),group:group_id
    sender:"",//sender user_id
    msg:"",
    extra:"",
    msg_index:1,//start from 1
    recall_index:1,//not exist or start from 1,this field exist means recall
}
//手动创建数据库
use im;
db.createCollection("msg");
db.msg.createIndex({chat_key:1,msg_index:1},{unique:true});
db.msg.createIndex({chat_key:1,recall_index:1},{unique:true,sparse:true});
sh.shardCollection("im.msg",{chat_key:"hashed"});

collection: ack
{
    chat_key:"",//user: join(sort(sender,target),"-"),group:group_id
    acker:"",//acker user_id
    read_msg_index:1,
}
use im;
db.createCollection("ack");
db.ack.createIndex({chat_key:1,acker:1},{unique:true});
sh.shardCollection("im.ack",{chat_key:"hashed"});

collection: relation
{
    main:"",
    main_type:"",//user or group
    sub:"",
    sub_type:"",//user or group
    duty:0,//this is used when main_type is group and sub is not empty,0-normal,1-system owner,2-owner,3-admin
    name:"",//this is used when sub is empty
}
use im;
db.createCollection("relation");
db.relation.createIndex({main:1,main_type:1,sub:1,sub_type:1},{unique_id:true});
sh.shardCollection("im.relation",{main:"hashed"});
```
