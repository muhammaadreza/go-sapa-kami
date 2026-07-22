package config

import (
	"os"
	"log"
	"strconv"

	"github.com/joho/godotenv"
)

// Config Structs
type Config struct {
	App 	 *App
	Postgres *Postgres
	Defaults *Defaults
	JWT 	 *JWT
}

// App Config
type App struct {
	Port   string
	Domain string
}

// Postgres Config
type Postgres struct {
	Port     string
	Host     string
	User     string
	Password string
	DBName   string
}

// UserDefaults Config
type Defaults struct {
	AdminID  uint
	Password string
}

// JWT Config
type JWT struct {
	Secret        []byte
	Issuer        string
	AccessExpiry  int
	RefreshExpiry int

	AccessMaxAge  int
	RefreshMaxAge int
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil { log.Println("Error loading .env file:", err) }

	app := App{
		Port:   getEnv("APP_PORT"),
		Domain: getEnv("APP_DOMAIN"),
	}

	defaults := Defaults{
		AdminID:  getEnvUint("DEFAULT_ADMIN_ID"),
		Password: getEnv("DEFAULT_PASSWORD"),
	}

	postgres := Postgres{
		Port:     getEnv("POSTGRES_PORT"),
		Host:     getEnv("POSTGRES_HOST"),
		User:     getEnv("POSTGRES_USER"),
		Password: getEnv("POSTGRES_PASSWORD"),
		DBName:   getEnv("POSTGRES_DB_NAME"),
	}

	jwt := JWT{
		Secret:        []byte(getEnv("JWT_SECRET")),
		Issuer:        getEnv("JWT_ISSUER"),
		AccessExpiry:  getEnvInt("JWT_ACCESS_EXPIRY"), // In minutes
		RefreshExpiry: getEnvInt("JWT_REFRESH_EXPIRY"), // In hours

		AccessMaxAge:  getEnvInt("JWT_ACCESS_EXPIRY") * 60, // Convert minutes to seconds
		RefreshMaxAge: getEnvInt("JWT_REFRESH_EXPIRY") * 3600, // Convert hours to seconds
	}

	return &Config{
		App:       &app,
		Postgres:  &postgres,
		Defaults:  &defaults,
		JWT:       &jwt,
	}
}

// Helper: Get ENV (string)
func getEnv(key string) string {
	value := os.Getenv(key)
	if value == "" { log.Fatalf("Environment variable '%s' is missing or empty.", key) }

	return value
}

// Helper: Get ENV (int)
func getEnvInt(key string) int {
	strValue := getEnv(key)
	value, err := strconv.Atoi(strValue)
	if err != nil { log.Fatalf("Environment variable '%s' must be a valid integer. Got: '%s'", key, strValue) }

	return value
}

// Helper: Get ENV (uint)
func getEnvUint(key string) uint {
	strValue := getEnv(key)
	value, err := strconv.ParseUint(strValue, 10, 32)
	if err != nil { log.Fatalf("Environment variable '%s' must be a valid unsigned integer. Got: '%s'", key, strValue) }

	return uint(value)
}

// Helper: Get ENV (bool)
func getEnvBool(key string) bool {
	strValue := getEnv(key)
	value, err := strconv.ParseBool(strValue)
	if err != nil { log.Fatalf("Environment variable '%s' must be 'true' or 'false'. Got: '%s'", key, strValue) }

	return value
}
