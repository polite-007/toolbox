package fofa

import "context"

// Count 执行搜索查询并返回结果数量
func (c *Client) Count(ctx context.Context, query string) (int, error) {
	req := &SearchRequest{
		Ctx:   ctx,
		Query: query,
		Size:  1,
		Page:  1,
	}

	resp, err := c.Search(req)
	if err != nil {
		return 0, err
	}
	return resp.Size, nil
}
