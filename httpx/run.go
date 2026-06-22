package httpx

import (
	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/gologger/levels"
	"github.com/projectdiscovery/httpx/runner"
)

// Run 执行 HTTP 探测
//   - targets: 目标列表（域名或 IP）
//   - onResult: 每条结果的回调函数
//
// 注意：会将全局 gologger 日志级别设为 Info 以输出探测详情
func (c *HttpxClient) Run(targets []string, onResult func(r runner.Result)) error {
	gologger.DefaultLogger.SetMaxLevel(levels.LevelInfo)

	options := c.toRunnerOptions(targets, onResult)

	if err := options.ValidateOptions(); err != nil {
		return err
	}

	httpxRunner, err := runner.New(options)
	if err != nil {
		return err
	}
	defer httpxRunner.Close()

	httpxRunner.RunEnumeration()
	return nil
}