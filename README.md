# Go Tang

> Cache Rules Everything Around Me

Go Tang is a dead simple golang + redis cache implementation that prevents the dog pile effect. Inspired by [this blog post](http://kovyrin.net/2008/03/10/dog-pile-effect-and-how-to-avoid-it-with-ruby-on-rails-memcache-client-patch/) and the Rails caching layer.

### `fetch`

```go

// make a function that the Fetch call should use to 
// fill the cache. The function returns a string value
// for the cache, a ttl (so you can use the header max-age
// from http calls, etc), and an error if something failed.
block := func() (string, intn, error) {
  return "myvalue", 1, nil // value, ttl, error
}

// This call tells the cache to fetch, and that the approximate
// fetch time will be 1 second.
fetchedValue, err1 := gotang.Fetch("mykey", block, gotang.Options{FetchTime: 1})

// This call will hit the cache, as we just fetched, and the ttl
// of 5 seconds hasn't expired.
cachedValue, err2 := gotang.Fetch("mykey", block, gotang.Options{FetchTime: 1})
```

### Disabling

The lib also supports disabling the cache layer, so methods like `fetch` jsut immediately returns the value from the block. This is helpful for development mode, or if you have logic that needs to disable caching.

You can create a new disabled gotang instance like this.

```go
cache := gotang.NewDisabled()
```

You can also bypass caching on a request basis, by passing in a boolen in the fetch options:

```go
notCached, err := gotang.Fetch("mykey", block, gotang.Options{
  FetchTime: 1
  Disabled: true
})
```

Name credit: [Steve Klise](http://sklise.com/)
