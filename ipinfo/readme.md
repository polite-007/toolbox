## 写在前面，

ipinfo是为了获取ip的网络信息，包括org，asn，net，city，country，loc，postal，region，timezone等等

网上有很多免费获取ip网络信息的url，下面是收集到的一些免费的url获取ip信息的

https://ipinfo.io/8.8.8.8/json #无明确限制
https://ip9.com.cn/get?ip=8.8.8.8 #无明确限制
https://ipwho.is/8.8.8.8 #免费版每天 1000 次请求
https://ipdata.info/json/8.8.8.8 #免费版每分钟 50 次请求

他们都能返回ip的一些网络信息，虽然数据丰富度岑其不齐，但是以免费角度足够使用，

## 需求

我们想写个库，功能就是免费获取ip的网络信息，要完成这个，得先确认一致的返回数据结构，上面的源的返回数据最终是转化成同一的数据结构

可以配置什么，网络代理，超时时间，ctx等等

## 外部库依赖

所有的网络请求都依赖 github.com/polite-007/toolbox/httpx