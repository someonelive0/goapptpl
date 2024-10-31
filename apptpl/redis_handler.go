package main

import (
	"context"
	"fmt"
	"net/url"
	"sort"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
)

type RedisHandler struct {
	Redisconfig *RedisConfig
	cli         *redis.Client
}

// r := app.Group("/redis")
func (p *RedisHandler) AddRouter(r fiber.Router) error {
	log.Info("RedisHandler AddRouter")

	// r.Get("/dbs", p.dbsHandler)
	r.Get("/db/:db/keys", p.keysHandler)
	r.Get("/db/:db/keys/:prefix", p.keysHandler)
	r.Get("/db/:db/key/:key", p.keyHandler)
	// r.Get("/table/:table/indexs", p.indexsHandler)
	// r.Get("/table/:table", p.tableHandler)

	return nil
}

// GET /dbs?mime=excel|json
// func (p *RedisHandler) dbsHandler(c fiber.Ctx) error {
// 	return nil
// }

// GET /db/:db/keys?mime=excel|json
func (p *RedisHandler) keysHandler(c fiber.Ctx) error {
	prefix, _ := url.QueryUnescape(c.Params("prefix"))
	if prefix == "" {
		prefix = "*"
	} else if prefix[len(prefix)-1] != '*' {
		prefix = prefix + "*"
	}
	log.Tracef("redis get key with prefix: %s", prefix)

	if p.cli == nil {
		err := p.openRedis()
		if err != nil {
			return err
		}
	}

	keys, err := p.cli.Keys(context.Background(), prefix).Result()
	if err != nil {
		log.Errorf("redis get keys prefix='%s' failed: %v", prefix, err)
		return err
	}
	sort.Strings(keys)

	return c.JSON(fiber.Map{
		"keys": keys,
	})

	// return nil
}

// GET /db/:db/key/:key?mime=excel|json
func (p *RedisHandler) keyHandler(c fiber.Ctx) error {
	key, _ := url.QueryUnescape(c.Params("key"))
	if p.cli == nil {
		err := p.openRedis()
		if err != nil {
			return err
		}
	}

	// 获取key的数据类型，例如string、list、hash等
	datatype, err := p.cli.Type(context.Background(), key).Result()
	if err != nil {
		log.Errorf("redis get key '%s' data type failed: %v", key, err)
		return err
	}
	fmt.Println("datatype:", datatype)

	// 根据数据类型获取key的值
	switch datatype {
	case "string":
		val, err := p.cli.Get(context.Background(), key).Result()
		if err != nil {
			log.Errorf("redis get key '%s' failed: %v", key, err)
			return err
		}
		log.Tracef("string key '%s' -> %#v\n", key, val)

		return c.JSON(val)

	case "list":
		val, err := p.cli.LRange(context.Background(), key, 0, -1).Result()
		if err != nil {
			log.Errorf("redis get key '%s' failed: %v", key, err)
			return err
		}
		log.Tracef("hash key '%s' -> %#v\n", key, val)

		return c.JSON(val)

	case "hash":
		val, err := p.cli.HGetAll(context.Background(), key).Result()
		if err != nil {
			log.Errorf("redis get key '%s' failed: %v", key, err)
			return err
		}
		log.Tracef("hash key '%s' -> %#v\n", key, val)

		return c.JSON(val)

	case "set":
		val, err := p.cli.SMembers(context.Background(), key).Result()
		if err != nil {
			log.Errorf("redis get key '%s' failed: %v", key, err)
			return err
		}
		log.Tracef("hash key '%s' -> %#v\n", key, val)

		return c.JSON(val)

	case "zset":
		val, err := p.cli.ZRange(context.Background(), key, 0, -1).Result()
		if err != nil {
			log.Errorf("redis get key '%s' failed: %v", key, err)
			return err
		}
		log.Tracef("hash key '%s' -> %#v\n", key, val)

		return c.JSON(val)

	default:
		return fmt.Errorf("unknown key type: %s", datatype)
	}

	// val, err := p.cli.LPop(context.Background(), key).Result()
	// if err != nil {
	// 	log.Errorf("redis get key '%s' failed: %v", key, err)
	// 	return err
	// }
	// fmt.Println("name:", val)

	// return c.JSON(fiber.Map{
	// 	key: val,
	// })

	return nil
}

func (p *RedisHandler) openRedis() error {
	cli := redis.NewClient(&redis.Options{
		Addr:     p.Redisconfig.Addr,
		Password: p.Redisconfig.Password,
		DB:       int(p.Redisconfig.Db),
		Protocol: 3, // specify 2 for RESP 2 or 3 for RESP 3
	})

	// 测试连接
	pong, err := cli.Ping(context.Background()).Result()
	if err != nil {
		log.Errorf("connect redis failed: %v", err)
		return err
	}
	log.Debugf("connect redis success: %s", pong)

	p.cli = cli
	return nil
}

func (p *RedisHandler) Close() error {
	if p.cli != nil {
		return p.cli.Close()
	}
	return nil
}
