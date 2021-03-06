package config

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/BurntSushi/toml"

	"github.com/lomik/graphite-clickhouse/helper/rollup"
	"github.com/lomik/zapwriter"
)

// Duration wrapper time.Duration for TOML
type Duration struct {
	time.Duration
}

var _ toml.TextMarshaler = &Duration{}

// UnmarshalText from TOML
func (d *Duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}

// MarshalText encode text with TOML format
func (d *Duration) MarshalText() ([]byte, error) {
	return []byte(d.Duration.String()), nil
}

// Value return time.Duration value
func (d *Duration) Value() time.Duration {
	return d.Duration
}

type Common struct {
	Listen string `toml:"listen"`
	// MetricPrefix   string    `toml:"metric-prefix"`
	// MetricInterval *Duration `toml:"metric-interval"`
	// MetricEndpoint string    `toml:"metric-endpoint"`
	MaxCPU int `toml:"max-cpu"`
}

type ClickHouse struct {
	Url              string    `toml:"url"`
	DataTable        string    `toml:"data-table"`
	DataTimeout      *Duration `toml:"data-timeout"`
	TreeTable        string    `toml:"tree-table"`
	ReverseTreeTable string    `toml:"reverse-tree-table"`
	TreeTimeout      *Duration `toml:"tree-timeout"`
	TagTable         string    `toml:"tag-table"`
	RollupConf       string    `toml:"rollup-conf"`
	ExtraPrefix      string    `toml:"extra-prefix"`
}

type Tags struct {
	Rules      string `toml:"rules"`
	Date       string `toml:"date"`
	InputFile  string `toml:"input-file"`
	OutputFile string `toml:"output-file"`
}

// Config ...
type Config struct {
	Common     Common             `toml:"common"`
	ClickHouse ClickHouse         `toml:"clickhouse"`
	Tags       Tags               `toml:"tags"`
	Logging    []zapwriter.Config `toml:"logging"`
	Rollup     *rollup.Rollup     `toml:"-"`
}

// NewConfig ...
func New() *Config {
	cfg := &Config{
		Common: Common{
			Listen: ":9090",
			// MetricPrefix: "carbon.graphite-clickhouse.{host}",
			// MetricInterval: &Duration{
			// 	Duration: time.Minute,
			// },
			// MetricEndpoint: MetricEndpointLocal,
			MaxCPU: 1,
		},
		ClickHouse: ClickHouse{
			Url: "http://localhost:8123",

			DataTable: "graphite",
			DataTimeout: &Duration{
				Duration: time.Minute,
			},
			TreeTable: "graphite_tree",
			TreeTimeout: &Duration{
				Duration: time.Minute,
			},
			RollupConf: "/etc/graphite-clickhouse/rollup.xml",
			TagTable:   "",
		},
		Tags: Tags{
			Date:  "2016-11-01",
			Rules: "/etc/graphite-clickhouse/tag.d/*.conf",
		},
		Logging: nil,
	}

	return cfg
}

func NewLoggingConfig() zapwriter.Config {
	cfg := zapwriter.NewConfig()
	cfg.File = "/var/log/graphite-clickhouse/graphite-clickhouse.log"
	return cfg
}

// PrintConfig ...
func PrintDefaultConfig() error {
	cfg := New()
	buf := new(bytes.Buffer)

	if cfg.Logging == nil {
		cfg.Logging = make([]zapwriter.Config, 0)
	}

	if len(cfg.Logging) == 0 {
		cfg.Logging = append(cfg.Logging, NewLoggingConfig())
	}

	encoder := toml.NewEncoder(buf)
	encoder.Indent = ""

	if err := encoder.Encode(cfg); err != nil {
		return err
	}

	fmt.Print(buf.String())
	return nil
}

// ReadConfig ...
func ReadConfig(filename string) (*Config, error) {
	var err error

	cfg := New()
	if filename != "" {
		b, err := ioutil.ReadFile(filename)
		if err != nil {
			return nil, err
		}

		body := string(b)

		// @TODO: fix for config starts with [logging]
		body = strings.Replace(body, "\n[logging]\n", "\n[[logging]]\n", -1)

		if _, err := toml.Decode(body, cfg); err != nil {
			return nil, err
		}
	}

	if cfg.Logging == nil {
		cfg.Logging = make([]zapwriter.Config, 0)
	}

	if len(cfg.Logging) == 0 {
		cfg.Logging = append(cfg.Logging, NewLoggingConfig())
	}

	if err := zapwriter.CheckConfig(cfg.Logging, nil); err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	return cfg, nil
}
