package config

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/CTO2BPublic/passage-server/pkg/models"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	Swagger       SwaggerConfig
	Auth          AuthConfig
	Tracing       TracingConfig
	Events        EventsConfig
	Log           LogConfig
	Db            DbConfig
	Creds         map[string]models.Credential `json:"-"`
	Roles         []models.AccessRole
	ApprovalRules []models.ApprovalRule
	SharedSecret  string `json:"-"`
}

type SwaggerConfig struct {
	Host string
}
type AuthConfig struct {
	OIDC AuthOIDC
	JWT  AuthJWT
}

type AuthOIDC struct {
	Enabled   bool
	IssuerURL string
	ClientID  string
}

type AuthJWT struct {
	Enabled                bool
	TokenHeader            string
	UsernameClaim          string
	GroupsClaim            string
	ProviderUsernamesClaim string
	HeaderPrefix           string
	JWKSURL                string
	Issuer                 string
}

type TracingConfig struct {
	Enabled         bool
	URL             string
	ConnectionType  string
	ServiceName     string
	EnvironmentName string
}

type EventsConfig struct {
	Kafka    EventsKafkaConfig
	Console  EventsConsoleConfig
	Database EventsDatabaseConfig
	Data     EventsData
}

type EventsKafkaConfig struct {
	Enabled           bool
	Hostname          string
	Topic             string
	NumPartitions     int
	ReplicationFactor int
}

type EventsConsoleConfig struct {
	Enabled bool
}

type EventsDatabaseConfig struct {
	Enabled bool
}

type EventsData struct {
	Tenant     string
	TypePrefix string
}

type LogConfig struct {
	Level  string
	Pretty bool
	Caller bool
}

type DbConfig struct {
	Engine string
	Psql   PsqlConfig
	Mysql  MysqlConfig
	Sqlite SqliteConfig
}

type PsqlConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	Database string
	Schema   string
	SSLMode  string
}

type MysqlConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	Database string
}

type SqliteConfig struct {
	Filename string
}

var k = koanf.New(".")
var configData Config

func InitConfig() error {

	// File config provider
	if err := k.Load(file.Provider("configs/config.yml"), yaml.Parser()); err != nil {
		return fmt.Errorf("error loading config file: %v", err)
	}
	if err := k.Load(file.Provider("configs/.secret.yml"), yaml.Parser()); err != nil {
		return fmt.Errorf("error loading secret config file: %v", err)
	}

	// ENV provider
	err := k.Load(env.Provider("PASSAGE_", ".", func(s string) string {
		return strings.ReplaceAll(strings.ToLower(
			strings.TrimPrefix(s, "PASSAGE_")), "_", ".")
	}), nil)
	if err != nil {
		return fmt.Errorf("error loading config from ENV: %v", err)
	}

	if err := k.Unmarshal("", &configData); err != nil {
		return fmt.Errorf("error unmarshaling config: %v", err)
	}

	if configData.SharedSecret == "" {
		configData.SharedSecret = generateRandomSecret()
	}

	return nil
}

func PrintConfig(configData *Config) {

	// Marshal the struct into JSON and omit sensitive fields based on json:"-" tag
	data, err := json.MarshalIndent(configData, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling config:", err)
		return
	}

	// Pretty-print the marshaled JSON
	fmt.Println("Config (omitting sensitive fields):")
	fmt.Println(string(data))
}

func GetConfig() *Config {
	return &configData
}

func (c *Config) GetCredentials(provider string) models.Credential {
	if credential, ok := configData.Creds[provider]; ok {
		return credential
	}
	return models.Credential{}
}

func generateRandomSecret() string {
	bytes := make([]byte, 32) // 256-bit secret
	_, err := rand.Read(bytes)
	if err != nil {
		panic("failed to generate random secret")
	}
	return base64.URLEncoding.EncodeToString(bytes)
}
