package config

import "os"

type Config struct {
	MongoURI string
	Database string
}

func Load() Config {
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		uri = "mongodb://localhost:27018"
	}

	db := "secure_actions"
	
	return Config{
		MongoURI: uri,
		Database: db,
	}
}
