**Just a prototype now!**
## What is it?

TLDR; Distribute your file(object) to many storages(eg, webdav, local storage), and encode them into RS(for fault tolerance), get or degrade get them after this.

## Why?

- [Data center fire in Singapore impacts Alibaba Cloud, causes ByteDance outage](https://www.datacenterdynamics.com/en/news/data-center-fire-in-singapore-impacts-alibaba-cloud-causes-bytedance-outage/)
- [Data Leak of AliYunPan](https://www.msn.cn/zh-cn/news/other/%E9%98%BF%E9%87%8C%E4%BA%91%E7%9B%98%E9%99%B7%E9%9A%90%E7%A7%81%E6%B3%84%E9%9C%B2%E9%A3%8E%E6%B3%A2-2%E4%BA%BF%E7%94%A8%E6%88%B7%E6%95%B0%E6%8D%AE%E5%AE%89%E5%8D%B1%E5%BC%95%E5%85%B3%E6%B3%A8/ar-AA1qLD4z?ocid=BingNewsSerp)
- limited rate of cloud storage (eg. Baidu Cloud)
- hard to maintain storage hardware for individual
## Features

- [x] Sharded download with parallel recovery capabilities
- [x] Independent storage of data shards, allowing for a certain degree of random read/write capability
- [x] Automatic orchestration based on speed tests
- [ ] Re-layout recovery capability 
- [ ] Streaming transmission capability
- [ ] Shard encryption and decryption

## How to build

```shell
# go enviroment is required
make
```

## How to use

### create

First, you need create a profile like this:

```json
{
    "name": "default",
    "stripeDepth": 2097152,
    "minDepth": 262144,
    "K": 2,
    "M": 1,
    "tolerance": 1,
    "servers": [
        {
            "type": "local",
            "path": "path/to/localdata1",
            "id": "localdata-1"
        },
        {
            "type": "local",
            "path": "path/to/localdata2",
            "id": "localdata-2"
        },
        {
            "type": "webdav",
            "path": "web/dav/url",
            "username": "xxx",
            "password": "xxx",
            "id": "xxx-cloud"
        },
    ]
}
```
  
Parameters:

| key         | explanation                                                                                                                                                                                                                                                                                               |
| ----------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| name        | the config name                                                                                                                                                                                                                                                                                           |
| stripeDepth | max size of a shard                                                                                                                                                                                                                                                                                       |
| minDepth    | min size of a shard                                                                                                                                                                                                                                                                                       |
| K           | RS K                                                                                                                                                                                                                                                                                                      |
| M           | RS M                                                                                                                                                                                                                                                                                                      |
| tolerance   | tolerance means that how many storages you can allow to lose at the same time.<br><br>sometime, many shards in one stripe could be stored in one faster storage to accelerate transmit speed, but weaken fault tolerance, this is a trade off, the max number of shards stored in one storage = M / tolerance |
| type        | 'webdav' or 'local' for now                                                                                                                                                                                                                                                                               |
| path        | url or local path                                                                                                                                                                                                                                                                                         |
| username    | username                                                                                                                                                                                                                                                                                                  |
| password    | password                                                                                                                                                                                                                                                                                                  |
| id          | storage identifier                                                                                                                                                                                                                                                                                        |

and then create config using:

```shell
./rnas create --config your_config.json
```

### test

test speed all storage and reorder them

```shell
./rnas test --config your_config.json
```

`--config` can be omitted and the configuration named `default` is read by default.

### put

```shell
./rnas put --config your_config.json path/to/object objectName
```

`--config` can be omitted and the configuration named `default` is read by default.
### get

```shell
./rnas get --config your_config.json objectName path/to/object
```

`--config` can be omitted and the configuration named `default` is read by default.


## Test Result

Origin speed:

|local|jianguoyun|aliyunpan|
|-|-|-|
|↑ 1165750217.35B/s|↑ 222773.69B/s|↑ 3221078.60B/s|
|↓ 8639321098.27B/s|↓ 1328854.22B/s|↓ 495322.04B/s|

Result speed:

↑ 1481376.94B/s 

↓ 11984124.16B/s

