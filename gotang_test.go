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
})

var _ = BeforeEach(func() {
	cache.Client.FlushDb()
})

var _ = Describe("Fetch", func() {

	// main time variable to adjust the speed of the test
	t := time.Duration(50) * time.Millisecond

	It("caches response in redis", func() {
		block := func() (string, time.Duration, error) {
			v := strconv.Itoa(rand.Intn(50000))
			return v, t * 5, nil
		}
		fetchedValue, err1 := cache.Fetch("mykey", block, t)
		cachedValue, err2 := cache.Fetch("mykey", block, t)
		Expect(err1).To(BeNil())
		Expect(err2).To(BeNil())
		Expect(fetchedValue).To(Equal(cachedValue))
	})

	It("expires cache after ttl specified in block", func() {
		block := func() (string, time.Duration, error) {
			v := strconv.Itoa(rand.Intn(50000))
			return v, t, nil
		}
		fetchedValue1, err1 := cache.Fetch("mykey", block, t*5)
		cachedValue1, err2 := cache.Fetch("mykey", block, t*5)
		time.Sleep(t * 2)
		fetchedValue2, err3 := cache.Fetch("mykey", block, t*5)
		Expect(err1).To(BeNil())
		Expect(err2).To(BeNil())
		Expect(err3).To(BeNil())
		Expect(fetchedValue1).To(Equal(cachedValue1))
		Expect(fetchedValue1).ToNot(Equal(fetchedValue2))
	})

	It("prevents dog pile effect", func() {

		// make fetch block that takes 2 seconds and caches for 10 seconds
		block := func() (string, time.Duration, error) {
			v := strconv.Itoa(rand.Intn(50000))
			time.Sleep(t * 2)
			return v, t * 10, nil
		}

		// set cache to old value for a second and fetchtime of 2 seconds
		cache.Set("mykey", "oldvalue", t, t*2)

		// setup for 3 concurrent processes
		messages := make(chan string, 2)
		fun := func() {
			value, err := cache.Fetch("mykey", block, t*2)
			Expect(err).To(BeNil())
			messages <- value
		}

		// first one should trigger cache
		go func() {
			time.Sleep(t)
			fun()
		}()

		// second one should use old cache and finish before
		go func() {
			time.Sleep(t * 3)
			fun()
		}()

		Expect(<-messages).To(Equal("oldvalue"))    // second one
		Expect(<-messages).ToNot(Equal("oldvalue")) // first one
	})

	It("ignores caching if disabled is true", func() {
		disabledCache := gotang.NewDisabled()
		block := func() (string, time.Duration, error) {
			return "myvalue", t, nil
		}
		fetchedValue, err := disabledCache.Fetch("mykey", block, t*5)
		Expect(err).To(BeNil())
		Expect(fetchedValue).To(Equal("myvalue"))
	})

})
