package httpx

import (
	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/gologger/levels"
	"github.com/projectdiscovery/httpx/runner"
)

func (c *HttpxClient) Run(i []string, s func(r runner.Result)) error {
	// 设置日志级别
	gologger.DefaultLogger.SetMaxLevel(levels.LevelInfo) // increase the verbosity (optional)

	options := c.Options
	options.InputTargetHost = i

	// 设置回调函数
	options.OnResult = s

	// 配置验证
	if err := options.ValidateOptions(); err != nil {
		return err
	}

	httpxRunner, err := runner.New(options)
	if err != nil {
		return err
	}
	defer httpxRunner.Close()

	// 执行扫描
	httpxRunner.RunEnumeration()

	return nil
}
