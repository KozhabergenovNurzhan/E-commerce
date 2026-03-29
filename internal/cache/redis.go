package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/models"
	"github.com/redis/go-redis/v9"
)

const categoriesKey = "categories"

type RedisCache struct {
	client *redis.Client
}

func NewRedisClient(addr, password string, db int) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
}

func NewProductCache(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}

func productKey(id int64) string {
	return fmt.Sprintf("product:%d", id)
}

func (c *RedisCache) GetProduct(ctx context.Context, id int64) (*models.Product, error) {
	data, err := c.client.Get(ctx, productKey(id)).Bytes()
	if err != nil {
		return nil, err
	}
	var p models.Product
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (c *RedisCache) SetProduct(ctx context.Context, p *models.Product, ttl time.Duration) error {
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, productKey(p.ID), data, ttl).Err()
}

func (c *RedisCache) DeleteProduct(ctx context.Context, id int64) error {
	return c.client.Del(ctx, productKey(id)).Err()
}

func (c *RedisCache) GetCategories(ctx context.Context) ([]*models.Category, error) {
	data, err := c.client.Get(ctx, categoriesKey).Bytes()
	if err != nil {
		return nil, err
	}
	var cats []*models.Category
	if err := json.Unmarshal(data, &cats); err != nil {
		return nil, err
	}
	return cats, nil
}

func (c *RedisCache) SetCategories(ctx context.Context, cats []*models.Category, ttl time.Duration) error {
	data, err := json.Marshal(cats)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, categoriesKey, data, ttl).Err()
}

func (c *RedisCache) DeleteCategories(ctx context.Context) error {
	return c.client.Del(ctx, categoriesKey).Err()
}
