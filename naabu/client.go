package naabu

import (
	"errors"
	"fmt"

	naaburesult "github.com/projectdiscovery/naabu/v2/pkg/result"
	"github.com/projectdiscovery/naabu/v2/pkg/runner"
)

// NaabuClient 封装 projectdiscovery/naabu 的端口扫描客户端。
type NaabuClient struct {
	config *Config
}

// NewNaabuClient 创建一个新的 NaabuClient，支持通过 Option 函数自定义配置。
func NewNaabuClient(opts ...Option) *NaabuClient {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	return &NaabuClient{config: cfg}
}

// toRunnerOptions 将用户 Config 转换为上游 runner.Options。
// 集中管理与上游库的映射，上游 API 变动只需修改此函数。
func (c *NaabuClient) toRunnerOptions(targets []string, onResult naaburesult.ResultFn) *runner.Options {
	cfg := c.config

	// 显式指定 Ports 时清空 TopPorts：上游 ParsePorts 会把 Ports 与 TopPorts 取并集，
	// 若不清空，默认的 TopPorts=100 会叠加到用户显式端口上，导致多扫端口。
	topPorts := cfg.TopPorts
	if cfg.Ports != "" {
		topPorts = ""
	}

	return &runner.Options{
		// 目标配置
		Host:      targets,
		HostsFile: cfg.TargetsFile,

		// 端口配置
		Ports:    cfg.Ports,
		TopPorts: topPorts,

		// 扫描配置
		ScanType:   cfg.ScanType,
		Rate:       cfg.Rate,
		Retries:    cfg.Retries,
		Timeout:    cfg.Timeout,
		Threads:    cfg.Threads,
		WarmUpTime: cfg.WarmUpTime,
		Verify:     cfg.Verify,

		// 功能开关
		WithHostDiscovery: cfg.WithHostDiscovery,
		Passive:           cfg.Passive,
		ServiceDiscovery:  cfg.ServiceDiscovery,
		ServiceVersion:    cfg.ServiceVersion,
		ExcludeCDN:        cfg.ExcludeCDN,

		// 网络配置
		Proxy:          cfg.Proxy,
		ProxyAuth:      cfg.ProxyAuth,
		Resolvers:      cfg.Resolvers,
		SystemResolver: cfg.SystemResolver,
		DnsOrder:       "l", // 对齐上游 CLI 默认值，未在 Config 中暴露

		// 输出行为
		Silent:        cfg.Silent,
		DisableStdout: true,

		// 运行时注入
		OnResult: onResult,
	}
}

func (c *NaabuClient) validate(targets []string, onResult any) error {
	if c == nil || c.config == nil {
		return errors.New("naabu client is nil")
	}
	if len(targets) == 0 && c.config.TargetsFile == "" {
		return errors.New("targets or targets file is required")
	}
	if onResult == nil {
		return errors.New("onResult callback is required")
	}
	if c.config.ScanType != "c" && c.config.ScanType != "s" {
		return fmt.Errorf("unsupported scan type %q", c.config.ScanType)
	}
	if c.config.Rate <= 0 {
		return errors.New("rate must be greater than 0")
	}
	if c.config.Retries < 0 {
		return errors.New("retries must be greater than or equal to 0")
	}
	if c.config.Timeout <= 0 {
		return errors.New("timeout must be greater than 0")
	}
	if c.config.Threads <= 0 {
		return errors.New("threads must be greater than 0")
	}
	if c.config.WarmUpTime < 0 {
		return errors.New("warm up time must be greater than or equal to 0")
	}
	if c.config.HostPortConcurrency <= 0 {
		return errors.New("host port concurrency must be greater than 0")
	}
	return nil
}
