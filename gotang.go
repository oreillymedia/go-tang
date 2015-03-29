package gotang

import (
	"time"

	redis "gopkg.in/redis.v2"
)

// all cache methods must adhere to this type:
type CacheBlock func() (string, int, error) // (value, ttl, error)

// Global redis client
var Client *redis.Client

func Fetch(key string, block CacheBlock, generation_time int) (string, error) {

	// compute stale key
	stale_key := key + ".stale"

	// get both cache keys: The real one and the stale key
	value, _ := Client.Get(key).Result()
	stale, _ := Client.Get(stale_key).Result()

	// if stale expired, it's time to regenerate our data
	if stale == "" {

		// update the stale key to make sure other requests server old cache
		err := Client.PSetEx(stale_key, time.Duration(generation_time)*time.Second, "refreshing").Err()
		if err != nil {
			return "", err
		}

		// force this request to regenerate cache
		value = ""
	}

	// if nothing in cache, or cache regeneration was forced above,
	// cache the data
	if value == "" {

		// get the data from the passed in block
		blockValue, blockTtl, blockErr := block()

		// if fetch failed, return error
		if blockErr != nil {
			return "", blockErr
		}

		value = blockValue

		// set real cache to ttl plus fetch time
		real_ttl := blockTtl + (generation_time * 2)
		realErr := Client.PSetEx(key, time.Duration(real_ttl)*time.Second, value).Err()
		if realErr != nil {
			return "", realErr
		}

		// set stale cache to just ttl so it triggers before cache
		staleErr := Client.PSetEx(stale_key, time.Duration(blockTtl)*time.Second, "good").Err()
		if staleErr != nil {
			return "", staleErr
		}
	}

	return value, nil
}
