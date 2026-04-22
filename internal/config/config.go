package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Log      LogConfig      `mapstructure:"log"`
	Wechat   WechatConfig   `mapstructure:"wechat"`
	Storage  StorageConfig  `mapstructure:"storage"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type DatabaseConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Username     string `mapstructure:"username"`
	Password     string `mapstructure:"password"`
	Database     string `mapstructure:"database"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

type RedisConfig struct {
	Enabled            bool   `mapstructure:"enabled"`
	Addr               string `mapstructure:"addr"`
	Username           string `mapstructure:"username"`
	Password           string `mapstructure:"password"`
	DB                 int    `mapstructure:"db"`
	KeyPrefix          string `mapstructure:"key_prefix"`
	DefaultTTLSeconds  int    `mapstructure:"default_ttl_seconds"`
	PlayerTTLSeconds   int    `mapstructure:"player_ttl_seconds"`
	GameListTTLSeconds int    `mapstructure:"game_list_ttl_seconds"`
	UserTTLSeconds     int    `mapstructure:"user_ttl_seconds"`
	RankTTLSeconds     int    `mapstructure:"rank_ttl_seconds"`
}

type WechatConfig struct {
	LoginURL string `mapstructure:"login_url"`
}

type StorageConfig struct {
	UploadDir  string `mapstructure:"upload_dir"`
	PublicPath string `mapstructure:"public_path"`
}

func Load(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.Username, c.Password, c.Host, c.Port, c.Database)
}

func (c RedisConfig) Prefix() string {
	if c.KeyPrefix == "" {
		return "mango-crew"
	}
	return c.KeyPrefix
}

func (c RedisConfig) DefaultTTL() time.Duration {
	if c.DefaultTTLSeconds <= 0 {
		return 5 * time.Minute
	}
	return time.Duration(c.DefaultTTLSeconds) * time.Second
}

func (c RedisConfig) PlayerTTL() time.Duration {
	if c.PlayerTTLSeconds <= 0 {
		return c.DefaultTTL()
	}
	return time.Duration(c.PlayerTTLSeconds) * time.Second
}

func (c RedisConfig) GameListTTL() time.Duration {
	if c.GameListTTLSeconds <= 0 {
		return c.DefaultTTL()
	}
	return time.Duration(c.GameListTTLSeconds) * time.Second
}

func (c RedisConfig) UserTTL() time.Duration {
	if c.UserTTLSeconds <= 0 {
		return c.DefaultTTL()
	}
	return time.Duration(c.UserTTLSeconds) * time.Second
}

func (c RedisConfig) RankTTL() time.Duration {
	if c.RankTTLSeconds <= 0 {
		return c.DefaultTTL()
	}
	return time.Duration(c.RankTTLSeconds) * time.Second
}
