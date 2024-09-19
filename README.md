**Just a prototype now!**
## What is it?

TLDR; Distribute your file(object) to many storages(eg, webdav, local storage), and encode them into RS(for fault tolerance), get or degrade get them after this.

## Why?

- [Data center fire in Singapore impacts Alibaba Cloud, causes ByteDance outage](https://www.datacenterdynamics.com/en/news/data-center-fire-in-singapore-impacts-alibaba-cloud-causes-bytedance-outage/)
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
| tolerance   | tolerance means that how many storages you can allow to lose at the same time.<br><br>sometime, many shards in one stripe could be stored in one faster storage to faster transmit speed, but weaken fault tolerance, this is a trade off, the max number of shards stored in one storage = M / tolerance |
| type        | 'webdav' or 'local' for now                                                                                                                                                                                                                                                                               |
| path        | url or local path                                                                                                                                                                                                                                                                                         |
| username    | username                                                                                                                                                                                                                                                                                                  |
| password    | password                                                                                                                                                                                                                                                                                                  |
| id          | storage identifier                                                                                                                                                                                                                                                                                        |

and then create config using:

```shell
./rnas create --config your_config.json
```

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