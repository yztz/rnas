package rnas

import (
	"sort"

	log "github.com/sirupsen/logrus"
)

type size_t int64

type StripeConfig struct {
	K           int `json:"K"`
	M           int `json:"M"`
	MinDepth	int `json:"minDepth"`
	StripeDepth int `json:"stripeDepth"`
}

type FileStripe struct {
	ID				int
	Size			size_t
	// RealSize		size_t
	ConfigName 		string
	Filepath 		string

	StripeConfig
}

type Shard struct {
	fileID		int
	shardIndex   int
	serverID     string
	shardHashname string
	dataShard	bool
	size 		int

	// StripeConfig
}


// Config struct for storing config data
type Config struct {
	Name 		string `json:"name"`
	Tolerance   int `json:"tolerance"`
	Servers     []*Server `json:"servers"`
	
	StripeConfig

	maps map[string]*Server
	slots	[]string
	dryrun bool
}

func(c *Config) Init() {
	log.Info("Config initializing...")
	if c.K + c.M > len(c.Servers) {
		log.Fatalf("need k + m <= servers, but %d + %d > %d", c.K, c.M, len(c.Servers))
	}

	if c.Tolerance > c.M {
		log.Fatalf("need tolerrance <= M, but %d > %d", c.Tolerance, c.M)
	}

	c.maps = make(map[string]*Server)
	c.slots = make([]string, c.K + c.M)

	log.Info("- init servers")
	for i := range c.Servers {
		server := c.Servers[i]

		_, ok := c.maps[server.Id]
		if ok {
			log.Fatalf("Duplicated id: %s", server.Id)
		}

		if c.dryrun {
			server.Type = "dryrun"
		}
		
		server.Init(c)

		c.maps[server.Id] = server
	}

	log.Info("- schedule slots")
	c.ScheduleSlots()
	log.Info("Init Done")
}



func(c *Config) ScheduleSlots() {
	log.Info("- start to schedule shard slots")

	freeSlots := c.M / c.Tolerance - 1
	var servers Servers = c.Servers
	sort.Sort(servers)

	// n := c.K + c.M - freeSlots

	log.Debugf("- free slots: %d", freeSlots)
	
	nextSlotIndex := 0
	for i := 0; i < servers.Len() && nextSlotIndex < len(c.slots); i++ {
		if !servers[i].reachable {
			continue
		}
		c.slots[nextSlotIndex] = servers[i].Id
		nextSlotIndex++
		for ;freeSlots > 0; freeSlots-- {
			c.slots[nextSlotIndex] = servers[i].Id
			nextSlotIndex++
		}
	}

	if nextSlotIndex != len(c.slots) {
		log.Fatalf("unenough servers. At least K + M - freeSlots = %d", c.K + c.M - (c.M / c.Tolerance - 1))
	}

	log.Infof("- slots alloc: %v", c.slots)

}

func(c *Config) TestAll() {
	for i := range c.Servers {
		server := c.Servers[i]

		log.Infof("Start to test speed for %s", server.Id)
		err := server.TestSpeed()
		if err != nil {
			log.Infof("failed to test speed for %s, because: %v", server.Id, err)
			continue
		}
		log.Infof("- Result: ↑ %.2fB/s ↓ %.2fB/s", server.UploadBandwidth, server.DownloadBandwidth)
	}
	c.ScheduleSlots()
	SaveConfigToDB(c)
}

