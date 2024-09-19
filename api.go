package rnas

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/klauspost/reedsolomon"
	log "github.com/sirupsen/logrus"
)


func (c *Config) Put(filepath string, _size int64, reader io.Reader) error {
	size := size_t(_size)
	log.Infof("Put object to %s with size %d", filepath, size)
	now := time.Now()
	enc, err := reedsolomon.New(c.K, c.M)
	if err != nil {
		return err
	} 

	fs := &FileStripe{Size: size, ConfigName: c.Name, Filepath: filepath, StripeConfig: c.StripeConfig }
	
	n := fs.K + fs.M
	// numDataShards := int((size + size_t(fs.StripeDepth) - 1) / size_t(fs.StripeDepth))
	numStripes := int((size + size_t(fs.StripeDepth * fs.K) - 1) / size_t(fs.StripeDepth * fs.K))
	// numShards := numStripes * n

	err = saveFileStripe(fs)
	if err != nil {
		return err
	}

	// log.Infof("- put object (%d shards) (%d stripes), size %d", numShards, numStripes, size)

	var done sync.WaitGroup
	done.Add(numStripes)

	log.Debug("numStripes:", numStripes)

	stripeIndex := 0
	for i := size_t(0); i < size; stripeIndex++{
		left := size - i
		shardSize := max(min(fs.StripeDepth, int(left / size_t(fs.K))), c.MinDepth)
		// stripeWidth := shardSize * n
		i += size_t(shardSize) * size_t(fs.K)
		log.Debugf("- handle stripe %d, shard size: %d", stripeIndex, shardSize)


		shards := make([]Shard, n)
		data := make([][]byte, n)

		// init shards and data
		for j := 0; j < n; j++ {
			data[j] = make([]byte, shardSize)
			shards[j] = Shard{
				fileID: fs.ID,
				shardIndex: stripeIndex * n + j,
				serverID: c.slots[j],
			}
		}

		// copy data
		for j := 0; j < fs.K; j++ {
			shards[j].dataShard = true
			n, err := reader.Read(data[j])
			if n > 0  && n < len(data[j]) {
				break
			}
			if err != nil {
				//todo goto err
				return fmt.Errorf("error while reading the data: %v", err)
			}
		}

		// encode and send stripe
		go func(data [][]byte, shards []Shard) {
			// encode
			log.Debug("- start to encode stripe")
			now := time.Now()
			err := enc.Encode(data)
			end := time.Since(now)
			
			if err != nil {
				log.Fatal("bad encoding:", err)
			}

			log.Debugf("- encode stripe cost %v", end)

			var wg sync.WaitGroup
			wg.Add(n)

			// send
			for i := 0; i < n; i++ {
				shard := &shards[i]
				shard.shardHashname = hashMD5(data[i])
				server := c.maps[shards[i].serverID]

				go func(shard *Shard, data []byte) {
					err := server.PutShard(shard, data);
					if err != nil {
						log.Errorf("error when put shard to server[%s]: %v", server.Id, err)
					} else {
						err := saveShard(shard)
						if err != nil {
							log.Errorf("saveShard error: %v", err)
						}
					}
					wg.Done()
					
				}(shard, data[i])
			}
			wg.Wait()
			done.Done()
		}(data, shards)

	}

	done.Wait()
	end := time.Since(now)
	fmt.Printf("%s has been put, took %v, speed %.2fB/s\n", filepath, end, float64(size) / float64(end.Seconds()))
	return nil
}



type StripeData struct {
	size int
	data [][]byte
	err error
}

type ShardData struct {
	data []byte
	shard *Shard
}

func (c *Config) ReadStream(filepath string) (io.Reader, error) {
	log.Infof("Get object from %s", filepath)

	fs, err := getFileStripe(filepath)
	if err != nil {
		return nil,err
	}

	allShards, err := getShards(fs.ID)
	if err != nil {
		return nil, err
	}

	n := fs.K + fs.M
	numShards := len(allShards)

	if numShards % n != 0 {
		return nil, fmt.Errorf("bad shards number: %d", numShards)
	}
	
	// numDataShards := int((fs.Size + size_t(fs.StripeDepth) - 1) / size_t(fs.StripeDepth))
	numStripes := numShards / n

	// stripeDataWidth := fs.K * int(fs.StripeDepth)


	log.Infof("- get object = (%d shards) = %d stripes, size %d", numStripes * n, numStripes, fs.Size)

	var data sync.Map
	pr, pw := io.Pipe()

	// check data and send them to receiver 
	go func() {
		for i := 0; i < numStripes; i++ {
			var stripe StripeData

			log.Debugf("- waiting for stripe %d", i)
			for {
				v, ok := data.Load(i)
				if ok {
					stripe = v.(StripeData)
					break
				}
				time.Sleep(time.Millisecond)
			}

			if stripe.err != nil {
				pw.CloseWithError(stripe.err)
				return
			}

			written := 0
			for _, data := range stripe.data {
				if written == stripe.size {
					// more data left, end
					break
				}
				size := min(len(data), stripe.size - written)
				pw.Write(data[:size])
				written += size
			}
			log.Debugf("- read %d bytes from stripe %d", written, i)
		}

		pw.Close()
	}()

	stripeIndex := 0
	// get stripes
	for i := size_t(0); i < fs.Size; stripeIndex++ {
		left := fs.Size - i
		shardSize := max(min(fs.StripeDepth, int(left / size_t(fs.K))), c.MinDepth)
		dataChan := make(chan ShardData, n)
		log.Debugf("- start to retrieve stripe %d, shard size: %d", stripeIndex, shardSize)
		i += size_t(shardSize) * size_t(fs.K)


		// get shards
		for j := 0; j < n; j++ {
			shard := &allShards[j + stripeIndex * n]
			server, ok := c.maps[shard.serverID]

			if !ok {
				// server not exists, may be removed?
				log.Warnf("server[%s] doesn't exist. skip", shard.serverID)
				dataChan <- ShardData{nil, shard}
				continue
			}

			// get shard
			go func(shard *Shard, shardSize int) {
				data := make([]byte, shardSize)
				_, err := server.GetShard(shard, data);

				if err != nil {
					log.Warnf("retrieve shard %d error: %v", shard.shardIndex, err)
					dataChan <- ShardData{nil, shard}
				} else {
					if hashMD5(data) != shard.shardHashname {
						log.Fatalf("bad data when verifying shard %d: %s", shard.shardIndex, shard.shardHashname)
					} else {
						log.Debugf("shard %d verified pass", shard.shardIndex)
					}
					dataChan <- ShardData{data, shard}
				}
			}(shard, shardSize)
		}

		go func(stripeIndex int, shardSize int, left size_t) {
			stripe := make([][]byte, n)
			received := 0
			dataReceived := 0
			validReceived := 0
			resultChan := make(chan [][]byte, 3)
			done := false
			var fixStatus chan bool

			for !done {
				// cannot fix
				if received == n && (fixStatus == nil || !<- fixStatus) {
					break
				}

				select {
				case v := <- dataChan:
					received++
					if v.data != nil {
						validReceived++
						if v.shard.dataShard {
							dataReceived++
						}
						stripe[v.shard.shardIndex - stripeIndex * n] = v.data
					}
					// receive all data
					if dataReceived == fs.K {
						resultChan <- stripe
					}
					// only triggered once
					if validReceived == fs.K && dataReceived < fs.K && fixStatus == nil {
						fixStatus = make(chan bool)
						copyStripe := make([][]byte, n)
						copy(copyStripe, stripe)
						// fix goroutine
						go func(data [][]byte) {
							log.Info("- got enough shards, begin to restore data in parallel")
							enc, err := reedsolomon.New(c.K, c.M)
							if err != nil {
								log.Error("failed to init decoder")
								fixStatus <- false
								return
							} 
							err = enc.ReconstructData(data)
							if err != nil {
								log.Error("failed to restore data")
								fixStatus <- false
								return
							} 
							log.Info("- fix done")
							fixStatus <- true
							resultChan <- data
						}(copyStripe)
					}
				case validStripe := <-resultChan:
					log.Infof("- got all shards for stripe %d", stripeIndex)
					done = true
					size :=  min(left, size_t(shardSize * fs.K))
					data.Store(stripeIndex, StripeData{size:int(size), data: validStripe[:fs.K]})
				}
			}

			if !done {
				data.Store(stripeIndex, StripeData{err: fmt.Errorf("stripe %d has left us permanently", stripeIndex)})
			}
		}(stripeIndex, shardSize, left)
	}

	return pr,nil
}