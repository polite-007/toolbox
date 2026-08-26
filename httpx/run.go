package httpx

import (
	"context"
	"errors"

	"github.com/projectdiscovery/httpx/runner"
)

// Run 执行 HTTP 探测，并按结果回调。
//   - ctx: 用于拦截初始化阶段；上游 RunEnumeration 不支持 context，扫描本身无法取消
//   - targets: 目标列表（host:port 或 URL）
//   - onResult: 每条探测结果的回调，使用本包 Result 类型
//
// 注意：不修改全局 gologger，库默认静默输出，结果通过回调返回。
func (c *HttpxClient) Run(ctx context.Context, targets []string, onResult func(r Result)) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if len(targets) == 0 {
		return errors.New("targets are required")
	}
	if onResult == nil {
		return errors.New("onResult callback is required")
	}

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
