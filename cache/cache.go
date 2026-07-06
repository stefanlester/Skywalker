package cache

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache interface {
	Has(string) (bool, error)
	Get(string) (interface{}, error)
	Set(string, interface{}, ...int) error
	Forget(string) error
	EmptyByMatch(string) error
	Empty() error
}

type RedisCache struct {
	Conn   *redis.Client
	Prefix string
}

type Entry map[string]interface{}

func (c *RedisCache) Has(str string) (bool, error) {
	key := fmt.Sprintf("%s:%s", c.Prefix, str)

	n, err := c.Conn.Exists(context.Background(), key).Result()
	if err != nil {
		return false, err
	}

	return n > 0, nil
}

func encode(item Entry) ([]byte, error) {
	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)
	err := e.Encode(item)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func decode(str string) (Entry, error) {
	item := Entry{}
	b := bytes.Buffer{}
	b.Write([]byte(str))
	d := gob.NewDecoder(&b)
	err := d.Decode(&item)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (c *RedisCache) Get(str string) (interface{}, error) {
	key := fmt.Sprintf("%s:%s", c.Prefix, str)

	cacheEntry, err := c.Conn.Get(context.Background(), key).Result()
	if err != nil {
		return nil, err
	}

	decoded, err := decode(cacheEntry)
	if err != nil {
		return nil, err
	}

	item := decoded[key]

	return item, nil
}

func (c *RedisCache) Set(str string, value interface{}, expires ...int) error {
	key := fmt.Sprintf("%s:%s", c.Prefix, str)

	entry := Entry{}
	entry[key] = value
	encoded, err := encode(entry)
	if err != nil {
		return err
	}

	var expiration time.Duration
	if len(expires) > 0 {
		expiration = time.Duration(expires[0]) * time.Second
	}

	return c.Conn.Set(context.Background(), key, string(encoded), expiration).Err()
}

func (c *RedisCache) Forget(str string) error {
	key := fmt.Sprintf("%s:%s", c.Prefix, str)

	return c.Conn.Del(context.Background(), key).Err()
}

func (c *RedisCache) EmptyByMatch(str string) error {
	key := fmt.Sprintf("%s:%s", c.Prefix, str)

	keys, err := c.getKeys(key)
	if err != nil {
		return err
	}

	for _, x := range keys {
		err := c.Conn.Del(context.Background(), x).Err()
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *RedisCache) Empty() error {
	key := fmt.Sprintf("%s:", c.Prefix)

	keys, err := c.getKeys(key)
	if err != nil {
		return err
	}

	for _, x := range keys {
		err := c.Conn.Del(context.Background(), x).Err()
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *RedisCache) getKeys(pattern string) ([]string, error) {
	ctx := context.Background()
	keys := []string{}

	iter := c.Conn.Scan(ctx, 0, fmt.Sprintf("%s*", pattern), 0).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return keys, err
	}

	return keys, nil
}
