package config

import (
	"fmt"
	"log"
	"os/exec"
	"path"
	"reflect"
	"strings"

	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
)

const (
	AppEnvDev     = "development"
	AppEnvStaging = "staging"
	AppEnvProd    = "production"
)

var (
	envMap = map[string]string{
		AppEnvDev:     "qa.",
		AppEnvStaging: "stage.",
		AppEnvProd:    "",
	}
)

// Config represents the global system wide configuration
type Config struct {
	AppEnv     string `env:"APP_ENV"`
	Port       string `env:"PORT"`
	Protocol   string `env:"PROTOCOL"`
	DomainName string `env:"DOMAIN_NAME"`

	LogDir      string `env:"LOG_DIR"`
	LogMaxSize  int    `env:"LOG_MAX_SIZE"`
	LogBackups  int    `env:"LOG_BACKUPS"`
	LogMaxAge   int    `env:"LOG_MAX_AGE"`
	LogCompress bool   `env:"LOG_COMPRESS"`

	RepositoryDriver string `env:"CUSTOMER_REPOSITORY_DRIVER"`
	RepositoryDSN    string `env:"CUSTOMER_REPOSITORY_DSN"`

	FrontendURL string

	SmsSenderAccountSid string `env:"SMS_SENDER_ACCOUNT_SID"`
	SmsSenderAuthToken  string `env:"SMS_SENDER_AUTH_TOKEN"`
	SmsSenderPhone      string `env:"SMS_SENDER_PHONE"`

	SESKey       string `env:"SES_AWS_KEY"`
	SESSecretKey string `env:"SES_AWS_SECRET_KEY"`
	SESRegion    string `env:"SES_AWS_REGION"`
	SESSender    string `env:"SES_SENDER"`

	TransportPublicKey  string `env:"CERT_TRANSPORT_PUBLIC_KEY"`
	TransportPrivateKey string `env:"CERT_TRANSPORT_PRIVATE_KEY"`
	TransportAlgo       string `env:"CERT_TRANSPORT_ALGO"`
	TransportKID        string `env:"CERT_TRANSPORT_KID"`

	SignaturePublicKey  string `env:"CERT_SIGNATURE_PUBLIC_KEY"`
	SignaturePrivateKey string `env:"CERT_SIGNATURE_PRIVATE_KEY"`
	SignatureAlgo       string `env:"CERT_SIGNATURE_ALGO"`
	SignatureKID        string `env:"CERT_SIGNATURE_KID"`

	EncryptionPublicKey  string `env:"CERT_ENCRYPTION_PUBLIC_KEY"`
	EncryptionPrivateKey string `env:"CERT_ENCRYPTION_PRIVATE_KEY"`
	EncryptionAlgo       string `env:"CERT_ENCRYPTION_ALGO"`

	HashIterations int    `env:"HASH_ITERATIONS"`
	HashKeyLength  int    `env:"HASH_KEY_LENGTH"`
	HashSaltString string `env:"HASH_SALT_STRING"`
}

func Load(filenames ...string) Config {
	if err := godotenv.Load(filenames...); err != nil {
		log.Printf("error loading environment files: %s\n", err) // todo: common logger
	}

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("failed to parse environment: %s\n", err)
	}

	// todo: make as absolute paths in '.env'
	rootDir, err := getRootDir()
	if err != nil {
		log.Fatalf(err.Error())
	}
	cfg.LogDir = path.Join(rootDir, cfg.LogDir)

	env, ok := envMap[cfg.AppEnv]
	if !ok {
		log.Fatal("config has invalid 'env' value")
	}
	cfg.FrontendURL = env + cfg.DomainName
	return cfg
}

// Print the contents of the config to the console
func (c Config) Print() {
	s := reflect.ValueOf(&c).Elem()
	typeOfT := s.Type()

	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		fmt.Printf("%s=%v\n", typeOfT.Field(i).Name, f.Interface())
	}
}

func getRootDir() (string, error) {
	out, err := exec.Command("pwd").Output()
	if err != nil {
		return "", errors.New("getting root dir")
	}
	rootDir := string(out)
	suffix := "\n"
	if len(rootDir) == 0 || !strings.HasSuffix(rootDir, suffix) {
		log.Fatalf("failed to get root dir")
	}
	rootDir = rootDir[:len(rootDir)-len(suffix)] // strip suffix
	return rootDir, nil
}
