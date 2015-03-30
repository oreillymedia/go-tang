package gotang

import (
	"time"

	redis "gopkg.in/redis.v2"
)

// Main package struct to hold redis conn
type Cache struct {
	Client   *redis.Client
	Disabled bool
}

// constructor to get a new *gotang.Cache
// receives options for the redis DB.
func New(opts *redis.Options) *Cache {
	c := Cache{
		Client: redis.NewClient(opts),
	}
	return &c
}

// constructor to get a disabled *gotang.Cache
// useful for development
func NewDisabled() *Cache {
	c := Cache{Disabled: true}
	return &c
}

// all cache methods must adhere to this type
// must return (cache value, ttl, error)
type FetchBlock func() (string, int, error)

// Options to be used with the cache methods
type Options struct {
	Ttl       int
	FetchTime int
	Disabled  bool
}

func (c *Cache) Fetch(key string, block FetchBlock, opts Options) (string, error) {

	// if disabled, just return block response
	if c.Disabled || opts.Disabled {
		blockValue, _, blockErr := block()
		return blockValue, blockErr
	}

	// get both cache keys: The real one and the stale key
	value, _ := c.Client.Get(key).Result()
	stale, _ := c.Client.Get(c.stalekey(key)).Result()

	// if stale expired, it's time to regenerate our data
	if stale == "" {

		// update the stale key to make sure other requests server old cache
		err := c.Client.PSetEx(c.stalekey(key), time.Duration(opts.FetchTime)*time.Second, "refreshing").Err()
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

		// use the block result as return value
		value = blockValue

		opts.Ttl = blockTtl
		setErr := c.Set(key, value, opts)
		if setErr != nil {
			return "", setErr
		}
	}

	return value, nil
}

func (c *Cache) Set(key string, value string, opts Options) error {

	// set stale cache to just ttl so it triggers before cache
	staleErr := c.Client.PSetEx(c.stalekey(key), time.Duration(opts.Ttl)*time.Second, "good").Err()
	if staleErr != nil {
		return staleErr
	}

	// set real cache to ttl plus fetch time
	realTtl := opts.Ttl + (opts.FetchTime * 2)
	realErr := c.Client.PSetEx(key, time.Duration(realTtl)*time.Second, value).Err()
	if realErr != nil {
		return realErr
	}

	return nil
}

func (c *Cache) stalekey(key string) string {
	return key + ".stale"
}
