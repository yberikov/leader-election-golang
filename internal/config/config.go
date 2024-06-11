package config

import "time"

type Config struct {
	ZookeeperServers []string
	LeaderTimeout    time.Duration
	AttempterTimeout time.Duration
	FileDir          string
	StorageCapacity  int
}
