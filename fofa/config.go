package fofa

import "time"

// SetTimeout 设置请求超时时间
func (c *Client) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
	c.httpClient.Timeout = timeout
}

// SetBaseURL 设置 API 基础地址（用于自定义或测试）
func (c *Client) SetBaseURL(baseURL string) {
	c.baseURL = baseURL
}
