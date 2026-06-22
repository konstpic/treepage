package config

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const DefaultConfigPath = "/opt/app/conf/config.yml"

type Loader struct {
	Path string
}

func (l *Loader) Load(dst any) error {
	path := l.Path
	if path == "" {
		path = os.Getenv("CONFIG_PATH")
		if path == "" {
			path = DefaultConfigPath
		}
	}

	if data, err := os.ReadFile(path); err == nil {
		if err := yaml.Unmarshal(data, dst); err != nil {
			return fmt.Errorf("parse config %s: %w", path, err)
		}
	}

	applyEnvOverrides(dst)
	applySecretEnv(dst)
	return validate(dst)
}

func applySecretEnv(v any) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return
	}
	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return
	}
	if pw := os.Getenv("DB_PASSWORD"); pw != "" {
		setNestedPassword(rv, "Postgres", pw)
	}
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		setNestedPassword(rv, "JWT", secret)
	}
}

func setNestedPassword(parent reflect.Value, fieldName, value string) {
	f := parent.FieldByName(fieldName)
	if !f.IsValid() || f.Kind() != reflect.Struct {
		return
	}
	pw := f.FieldByName("Password")
	if pw.IsValid() && pw.Kind() == reflect.String && pw.CanSet() {
		pw.SetString(value)
	}
	secret := f.FieldByName("Secret")
	if secret.IsValid() && secret.Kind() == reflect.String && secret.CanSet() && fieldName == "JWT" {
		secret.SetString(value)
	}
}

func applyEnvOverrides(v any) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return
	}
	applyEnvToStruct(rv.Elem(), "")
}

func applyEnvToStruct(v reflect.Value, prefix string) {
	if v.Kind() != reflect.Struct {
		return
	}
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		fv := v.Field(i)
		yamlTag := field.Tag.Get("yaml")
		if yamlTag == "" || yamlTag == "-" {
			continue
		}
		key := strings.Split(yamlTag, ",")[0]
		envKey := buildEnvKey(prefix, key)
		if fv.Kind() == reflect.Struct {
			applyEnvToStruct(fv, envKey)
			continue
		}
		if val, ok := os.LookupEnv(envKey); ok {
			setFieldFromString(fv, val)
		}
	}
}

func buildEnvKey(prefix, key string) string {
	part := strings.ToUpper(strings.ReplaceAll(key, "-", "_"))
	if prefix == "" {
		return part
	}
	return prefix + "_" + part
}

func setFieldFromString(fv reflect.Value, val string) {
	switch fv.Kind() {
	case reflect.String:
		fv.SetString(val)
	case reflect.Bool:
		b, _ := strconv.ParseBool(val)
		fv.SetBool(b)
	case reflect.Int, reflect.Int64:
		n, _ := strconv.ParseInt(val, 10, 64)
		fv.SetInt(n)
	case reflect.Uint, reflect.Uint64:
		n, _ := strconv.ParseUint(val, 10, 64)
		fv.SetUint(n)
	case reflect.Float64:
		f, _ := strconv.ParseFloat(val, 64)
		fv.SetFloat(f)
	default:
		if fv.Type() == reflect.TypeOf(time.Duration(0)) {
			d, err := time.ParseDuration(val)
			if err == nil {
				fv.Set(reflect.ValueOf(d))
			}
		}
	}
}

func validate(v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil
	}
	t := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("validate")
		if tag == "required" {
			fv := rv.Field(i)
			if isZero(fv) {
				return fmt.Errorf("config field %s is required", field.Name)
			}
		}
	}
	return nil
}

func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Int, reflect.Int64:
		return v.Int() == 0
	case reflect.Bool:
		return !v.Bool()
	default:
		return v.IsZero()
	}
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

func (s ServerConfig) Addr() string {
	host := s.Host
	if host == "" {
		host = "0.0.0.0"
	}
	port := s.Port
	if port == 0 {
		port = 8080
	}
	return fmt.Sprintf("%s:%d", host, port)
}

type PostgresConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	User     string `yaml:"user"`
	Password string `yaml:"-" validate:"required"`
	SSLMode  string `yaml:"ssl_mode"`
}

func (p PostgresConfig) DSN() string {
	host := p.Host
	if host == "" {
		host = "localhost"
	}
	port := p.Port
	if port == 0 {
		port = 5432
	}
	ssl := p.SSLMode
	if ssl == "" {
		ssl = "disable"
	}
	password := p.Password
	if password == "" {
		password = os.Getenv("DB_PASSWORD")
	}
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		host, port, p.User, password, p.Database, ssl,
	)
}

type LoggingConfig struct {
	Level string `yaml:"level"`
}

type RedisConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"-"`
	DB       int    `yaml:"db"`
}

func (r RedisConfig) Addr() string {
	host := r.Host
	if host == "" {
		host = "localhost"
	}
	port := r.Port
	if port == 0 {
		port = 6379
	}
	return fmt.Sprintf("%s:%d", host, port)
}

type JWTConfig struct {
	Secret           string        `yaml:"-" validate:"required"`
	AccessTTL        time.Duration `yaml:"access_ttl"`
	RefreshTTL       time.Duration `yaml:"refresh_ttl"`
	Issuer           string        `yaml:"issuer"`
	Audience         string        `yaml:"audience"`
	SigningAlgorithm string        `yaml:"signing_algorithm"`
}

func (j JWTConfig) AccessDuration() time.Duration {
	if j.AccessTTL == 0 {
		return 15 * time.Minute
	}
	return j.AccessTTL
}

func (j JWTConfig) RefreshDuration() time.Duration {
	if j.RefreshTTL == 0 {
		return 7 * 24 * time.Hour
	}
	return j.RefreshTTL
}

type OIDCConfig struct {
	Enabled      bool   `yaml:"enabled"`
	IssuerURL    string `yaml:"issuer_url"`
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"-"`
	RedirectURL  string `yaml:"redirect_url"`
	Scopes       string `yaml:"scopes"`
	RoleClaim    string `yaml:"role_claim"`
	GroupClaim   string `yaml:"group_claim"`
	SyncGroups   bool   `yaml:"sync_groups"`
}

type GitConfig struct {
	SyncInterval time.Duration `yaml:"sync_interval"`
	WorkDir      string        `yaml:"work_dir"`
}

type SecurityConfig struct {
	CSRFSecret     string `yaml:"-"`
	RateLimitRPS   int    `yaml:"rate_limit_rps"`
	AllowedOrigins string `yaml:"allowed_origins"`
	EnableAuditLog bool   `yaml:"enable_audit_log"`
}
