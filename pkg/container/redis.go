package container

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/ory/dockertest"
)

func (builder *Builder) RunRedis() (*redis.Client, error) {
	container, err := builder.FindContainer("redis")
	if err != nil {
		return nil, err
	}

	if container != nil {
		builder.containerIDs[container.ID] = true
		client := redis.NewClient(&redis.Options{
			Addr: fmt.Sprintf("localhost:%d", container.Ports[0].PublicPort),
		})
		err := client.Ping(context.Background()).Err()
		if err != nil {
			return nil, err
		}
		return client, nil
	}

	resource, err := builder.Run("redis", "6.0.9-alpine", nil)
	if err != nil {
		return nil, err
	}

	builder.containerIDs[resource.Container.ID] = true

	return builder.BuildRedisClient(resource)
}

func (builder *Builder) BuildRedisClient(resource *dockertest.Resource) (*redis.Client, error) {
	var (
		client *redis.Client
		ctx    = context.Background()
	)

	err := builder.Retry(func() error {
		client = redis.NewClient(&redis.Options{
			Addr: fmt.Sprintf("localhost:%s", resource.GetPort("6379/tcp")),
		})

		return client.Ping(ctx).Err()
	})
	if err != nil {
		return nil, err
	}

	return client, nil
}
