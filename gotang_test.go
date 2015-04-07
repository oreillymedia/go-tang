package gotang_test

import (
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/joho/godotenv"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/oreillymedia/go-tang"
	"gopkg.in/redis.v2"
)

var cache *gotang.Cache
var disabledCache *gotang.Cache

func TestTest(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test Suite")
}

var _ = BeforeSuite(func() {
	rand.Seed(time.Now().UTC().UnixNano())
	godotenv.Load(".env")
	cache = gotang.New(&redis.Options{
		Addr:    os.Getenv("REDIS_URL"),
		Network: "tcp",
	})
	disabledCache = gotang.NewDisabled()
})

var _ = BeforeEach(func() {
	cache.Client.FlushDb()
})

func Sleep(secs int) {
	time.Sleep(time.Duration(secs) * time.Second)
}

var _ = Describe("Gotang", func() {

	Context("#Fetch", func() {

		// main time variable to adjust the speed of the test
		t := 1

		It("caches response in redis", func() {
			block := func() (string, int, error) {
				v := strconv.Itoa(rand.Intn(50000))
				return v, t * 5, nil
			}
			fetchedValue, err1 := cache.Fetch("mykey", block, gotang.Options{FetchTime: t})
			cachedValue, err2 := cache.Fetch("mykey", block, gotang.Options{FetchTime: t})
			Expect(err1).To(BeNil())
			Expect(err2).To(BeNil())
			Expect(fetchedValue).To(Equal(cachedValue))
		})

		It("expires cache after ttl specified in block", func() {
			block := func() (string, int, error) {
				v := strconv.Itoa(rand.Intn(50000))
				return v, t, nil
			}
			fetchedValue1, err1 := cache.Fetch("mykey", block, gotang.Options{FetchTime: t * 5})
			cachedValue1, err2 := cache.Fetch("mykey", block, gotang.Options{FetchTime: t * 5})
			Sleep(t * 2)
			fetchedValue2, err3 := cache.Fetch("mykey", block, gotang.Options{FetchTime: t * 5})
			Expect(err1).To(BeNil())
			Expect(err2).To(BeNil())
			Expect(err3).To(BeNil())
			Expect(fetchedValue1).To(Equal(cachedValue1))
			Expect(fetchedValue1).ToNot(Equal(fetchedValue2))
		})

		It("prevents dog pile effect", func() {

			// make fetch block that takes 2 seconds and caches for 10 seconds
			block := func() (string, int, error) {
				v := strconv.Itoa(rand.Intn(50000))
				Sleep(t * 2)
				return v, t * 10, nil
			}

			// set cache to old value for a second and fetchtime of 2 seconds
			cache.Set("mykey", "oldvalue", gotang.Options{
				Ttl:       t,
				FetchTime: t * 2,
			})

			// setup for 3 concurrent processes
			messages := make(chan string, 2)
			fun := func() {
				value, err := cache.Fetch("mykey", block, gotang.Options{FetchTime: t * 2})
				Expect(err).To(BeNil())
				messages <- value
			}

			// first one should trigger cache
			go func() {
				Sleep(t)
				fun()
			}()

			// second one should use old cache and finish before
			go func() {

				fun()
			}()

			Expect(<-messages).To(Equal("oldvalue"))    // second one
			Expect(<-messages).ToNot(Equal("oldvalue")) // first one
		})

		It("ignores caching if globally disabled", func() {
			block := func() (string, int, error) {
				v := strconv.Itoa(rand.Intn(50000))
				return v, t, nil
			}
			fetchedValue1, err1 := disabledCache.Fetch("mykey", block, gotang.Options{FetchTime: t * 3})
			fetchedValue2, err2 := disabledCache.Fetch("mykey", block, gotang.Options{FetchTime: t * 3})
			Expect(err1).To(BeNil())
			Expect(err2).To(BeNil())
			Expect(fetchedValue1).ToNot(Equal(fetchedValue2))
		})

		It("ignores caching if options disabled", func() {
			block := func() (string, int, error) {
				v := strconv.Itoa(rand.Intn(50000))
				return v, t, nil
			}
			opts := gotang.Options{FetchTime: t * 3, Disabled: true}
			fetchedValue1, err1 := cache.Fetch("mykey", block, opts)
			fetchedValue2, err2 := cache.Fetch("mykey", block, opts)
			Expect(err1).To(BeNil())
			Expect(err2).To(BeNil())
			Expect(fetchedValue1).ToNot(Equal(fetchedValue2))
		})
	})

	Context("#GetAll", func() {

		It("gets all keys", func() {
			cache.Set("mykey1", "myvalue1", gotang.Options{Ttl: 5, FetchTime: 1})
			cache.Set("mykey2", "myvalue2", gotang.Options{Ttl: 5, FetchTime: 1})
			cache.Set("mykey3", "myvalue3", gotang.Options{Ttl: 5, FetchTime: 1})
			vals, err := cache.GetAll("mykey1", "mykey2")
			Expect(err).To(BeNil())
			Expect(vals).To(ConsistOf("myvalue1", "myvalue2"))
		})

		It("retains nil values as empty strings", func() {
			cache.Set("mykey1", "myvalue1", gotang.Options{Ttl: 5, FetchTime: 1})
			cache.Set("mykey3", "myvalue3", gotang.Options{Ttl: 5, FetchTime: 1})
			vals, err := cache.GetAll("mykey1", "mykey2", "mykey3")
			Expect(err).To(BeNil())
			Expect(vals).To(ConsistOf("myvalue1", "", "myvalue3"))
		})

		It("ignores caching if globally disabled", func() {
			vals, err := disabledCache.GetAll("mykey1", "mykey2", "mykey3")
			Expect(err).To(BeNil())
			Expect(vals).To(ConsistOf("", "", ""))
		})

	})

	Context("#ParseRedisUrl", func() {

		It("handles real redis url", func() {
			host, pw, _ := gotang.ParseRedisUrl("redis://:password@domain:port")
			Expect(host).To(Equal("domain:port"))
			Expect(pw).To(Equal("password"))
		})

		It("handles localhost url", func() {
			host, pw, _ := gotang.ParseRedisUrl("redis://domain:port")
			Expect(host).To(Equal("domain:port"))
			Expect(pw).To(Equal(""))
		})

	})

})
