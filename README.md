# webkit

web应用开发框架，集合http+rpc微服务+mysql，redis数据库+日志+配置加载配置中心rpc获取+auth用户token授权服务。
gohttp：基于gin的http服务，提供简单可配置的http server启动，可选中间件rsa，aes，token解析用户信息全链路传输；
gorpc：基于rpcx的高效稳定的rpc服务，提供可配置启动，日志拦截，token校验，tls配置功能；
database：提供基于sqlx的mysql服务，根据username随机分库分表及访问方案，redis访问，trylock等集成；
logging：日志功能，每日零点按日期分隔日志；
auth：用户认证授权中心，sso，用户token生成与鉴权服务，双token（refreshtoken和accesstoken）机制；
config：配置服务，提供配置中心接口及配置中心启动，从配置中心读取webkit各模块启动所需的配置信息。
具体应用详见各package test文件。