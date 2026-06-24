package redis

import (
	"context"
	"fmt"

	"github.com/redis/rueidis"
)

type Client struct {
	c rueidis.Client
}

func Connect(ctx context.Context, addr string) (*Client, error) {
	if addr == "" {
		return nil, fmt.Errorf("redis addr is required")
	}

	c, err := rueidis.NewClient(rueidis.ClientOption{
		Addr: addr,
	})
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}

	if err := c.Do(ctx, c.B().Ping().Build()).Error(); err != nil {
		c.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}

	return &Client{c: c}, nil
}

func (c *Client) HealthCheck(ctx context.Context) error {
	if c == nil || c.c == nil {
		return fmt.Errorf("redis client not initialized")
	}
	return c.Do(ctx, c.B().Ping().Build()).Error()
}

func (c *Client) Do(ctx context.Context, cmd rueidis.Completed) rueidis.RedisResult {
	return c.c.Do(ctx, cmd)
}

func (c *Client) Close() {
	if c != nil && c.c != nil {
		c.c.Close()
	}
}
