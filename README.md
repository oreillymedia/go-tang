# Go Tang

> Cache Rules Everything Around Me

Go Tang is a dead simple golang + redis cache implementation that prevents the dog pile effect. Inspired by [this blog post](http://kovyrin.net/2008/03/10/dog-pile-effect-and-how-to-avoid-it-with-ruby-on-rails-memcache-client-patch/) and the Rails caching layer.

### `fetch`

```go

// make a function that the Fetch call should use to 
// fill the cache. The function returns a string value
// for the cache, a ttl (so you can use the header max-age
// from http calls, etc), and an error if something failed.
block := func() (string, time.Duration, error) {
  return "myvalue", time.Duration(1) * time.Second, nil // value, ttl, error
}

// This call tells the cache to fetch, and that the approximate
// fetch time will be 1 second.
fetchedValue, err1 := gotang.Fetch("mykey", block, time.Duration(1) * time.Second)

// This call will hit the cache, as we just fetched, and the ttl
// of 5 seconds hasn't expired.
cachedValue, err2 := gotang.Fetch("mykey", block, time.Duration(1) * time.Second)
```

Name credit: [Steve Klise](http://sklise.com/)
