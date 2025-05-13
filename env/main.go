package env

import (
	"os"
	"strings"
	"todorist/utils"
)

var (
	Port                       uint64
	GinMode, JwtScretKey       string
	AllowOrigins, AllowMethods []string

	// DATABASE
	PgHost,
	PgUser,
	PgPassword,
	PgDatabase,
	DbString string
	PgPort uint64
)

func GetEnv() {
	AllowOrigins = strings.Split(os.Getenv("ALLOW_ORIGINS"), ",")
	AllowMethods = strings.Split(os.Getenv("ALLOW_METHODS"), ",")
	GinMode = os.Getenv("GIN_MODE")
	Port = utils.ParseToUint(os.Getenv("PORT"), 8003)

	// Database Configuration
	PgHost = os.Getenv("PG_HOST")
	PgPort = utils.ParseToUint(os.Getenv("PG_PORT"), 5432)
	PgUser = os.Getenv("PG_USER")
	PgPassword = os.Getenv("PG_PASSWORD")
	PgDatabase = os.Getenv("PG_DATABASE")
	DbString = os.Getenv("DB_STRING")

	// JWT and other secrets
	JwtScretKey = os.Getenv("JWT_SECRET_KEY")
}
