package config

import "os"

type Config struct {
	MongoURI       string
	Database       string
	MongoTLS       bool
	MongoTLSCAFile string
	MongoCertFile  string
	MongoKeyFile   string
	MongoAuthDB    string
	MongoUsername  string
	MongoPassword  string
}

func Load() Config {
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		uri = "mongodb://localhost:27018"
	}

	db := "secure_actions"

	return Config{
		MongoURI:       uri,
		Database:       db,
		MongoTLS:       getEnvBool("MONGO_TLS", false),
		MongoTLSCAFile: os.Getenv("MONGO_TLS_CA_FILE"),
		MongoCertFile:  os.Getenv("MONGO_CERT_FILE"),
		MongoKeyFile:   os.Getenv("MONGO_KEY_FILE"),
		MongoAuthDB:    getEnvDefault("MONGO_AUTH_DB", "admin"),
		MongoUsername:  os.Getenv("MONGO_USERNAME"),
		MongoPassword:  os.Getenv("MONGO_PASSWORD"),
	}
}

func getEnvBool(key string, defaultVal bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val == "true" || val == "1" || val == "yes"
}

func getEnvDefault(key, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val
}
