package config

import "time"

type Config struct {
	ZookeeperServers []string
	LeaderTimeout    time.Duration
	AttempterTimeout time.Duration
	WriteInterval    time.Duration
	FileDir          string
	StorageCapacity  int
}
