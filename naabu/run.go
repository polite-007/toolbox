package naabu

import (
	"context"
	"errors"
	"fmt"
	"sync"

	naaburesult "github.com/projectdiscovery/naabu/v2/pkg/result"
	"github.com/projectdiscovery/naabu/v2/pkg/runner"
)

// Run 执行端口扫描，并按开放端口逐条回调结果。
//   - targets: 目标列表，支持域名、IP、CIDR
//   - onResult: 每个开放端口的回调函数
func (c *NaabuClient) Run(ctx context.Context, targets []string, onResult func(r Result)) error {
	if err := c.validate(targets, onResult); err != nil {
		return err
	}

	options := c.toRunnerOptions(targets, func(hostResult *naaburesult.HostResult) {
		for _, r := range toResults(hostResult) {
			onResult(r)
		}
		if c.config.OnHostDone != nil {
			c.config.OnHostDone(toHostResult(hostResult))
		}
	})

	// 复用上游参数校验，错误通过返回值暴露（原 CLI 使用 gologger.Fatal 退出进程）。
	if err := options.ValidateOptions(); err != nil {
		return err
	}

	naabuRunner, err := runner.NewRunner(options)
	if err != nil {
		return err
	}
	defer naabuRunner.Close()

	return naabuRunner.RunEnumeration(ctx)
}

// RunHostPorts 执行 host:port 列表扫描，并按输入顺序回调已开放端口。
//
// 输入支持 "host:port"、"u:host:port"、"udp:host:port" 和 "tcp:host:port"。
// IPv6 host:port 第一版暂不支持。相同 host 的端口会聚合后并发扫描，避免交叉扫描额外端口。
func (c *NaabuClient) RunHostPorts(ctx context.Context, hostPorts []string, onResult func(r Result)) error {
	if len(hostPorts) == 0 {
		return errors.New("host ports are required")
	}
	if onResult == nil {
		return errors.New("onResult callback is required")
	}

	targets := make([]hostPortTarget, 0, len(hostPorts))
	for _, raw := range hostPorts {
		target, err := parseHostPortTarget(raw)
		if err != nil {
			return err
		}
		targets = append(targets, target)
	}

	if err := c.validate([]string{targets[0].Host}, onResult); err != nil {
		return err
	}

	groups := groupHostPortTargets(targets)
	found := make(map[string]Result)
	var foundMu sync.Mutex
	var wg sync.WaitGroup
	errCh := make(chan error, len(groups))
	sem := make(chan struct{}, c.config.HostPortConcurrency)

	for host, groupedTargets := range groups {
		wg.Add(1)
		go func(host string, groupedTargets []hostPortTarget) {
			defer wg.Done()

			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				errCh <- fmt.Errorf("scan host %s: %w", host, ctx.Err())
				return
			}

			client := &NaabuClient{config: cloneConfig(c.config)}
			client.config.Ports = hostPortExpression(groupedTargets)
			client.config.TopPorts = ""

			err := client.Run(ctx, []string{host}, func(result Result) {
				collectHostPortResults(&foundMu, found, host, result)
			})
			if err != nil {
				errCh <- fmt.Errorf("scan host %s: %w", host, err)
			}
		}(host, groupedTargets)
	}

	wg.Wait()
	close(errCh)

	var errs []error
	for err := range errCh {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	for _, result := range orderedHostPortResults(targets, found) {
		onResult(result)
	}
	return nil
}
