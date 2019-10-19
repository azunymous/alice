package dependencies

import (
	"github.com/alice-ws/alice/data"
	"github.com/alice-ws/alice/minioclient"
	"github.com/alice-ws/alice/redisclient"
	"github.com/spf13/viper"
	"log"
	"time"
)

func (d *Dependencies) GetDB() data.DB {
	quit := make(chan bool)
	done := make(chan *redisclient.RedisClient)
	go d.tryRedis(done, quit, viper.GetString("redis.addr"))
	select {
	case rc := <-done:
		d.SetHealthy("redis")
		return rc
	case <-time.After(viper.GetDuration("redis.timeout")):
		d.SetFallback("redis")
		quit <- true
		log.Printf("Failed connecting to redis in  %s. Falling back to in memory DB", viper.GetString("redis.timeout"))
		return data.NewMemoryDB()
	}
}

func (d *Dependencies) tryRedis(done chan *redisclient.RedisClient, quit chan bool, addr string) {
	for {
		select {
		case <-quit:
			return
		default:
			log.Printf("Trying connection to redis")
			client, err := redisclient.ConnectToRedis(addr)
			if err == nil {
				log.Printf("Successfully connected to redis on %s", addr)
				done <- client
				return
			}
			log.Printf("Error connecting to redis %v", err)
			time.Sleep(1 * time.Second)
		}
	}
}

func (d *Dependencies) GetImageRepository() data.MediaRepo {
	quit := make(chan bool)
	done := make(chan data.MediaRepo)
	go d.tryMinio(done, quit, viper.GetString("minio.addr"))
	select {
	case mc := <-done:
		d.SetHealthy("minio")
		return mc
	case <-time.After(viper.GetDuration("minio.timeout")):
		quit <- true
		log.Printf("Failed connecting to minio in  %s. falling back to local filesystem", viper.GetString("minio.timeout"))
		d.SetFallback("minio")
		return data.NewLocalRepo(viper.GetString("board.images.dir"))
	}
}

func (d *Dependencies) tryMinio(done chan data.MediaRepo, quit chan bool, addr string) {
	for {
		select {
		case <-quit:
			return
		default:
			log.Printf("Trying connection to minio")
			mc, minioErr := minioclient.ConnectToMinioWithTimeout(addr, ImageGroup, viper.GetString("minio.access"), viper.GetString("minio.secret"), viper.GetDuration("minio.timeout")/2)
			if minioErr == nil {
				log.Printf("Successfully connected to minio on %s", addr)
				d.SetHealthy("redis")
				done <- mc
				return
			}
			log.Printf("Error connecting to minio %v", minioErr)
			time.Sleep(1 * time.Second)
		}
	}
}
