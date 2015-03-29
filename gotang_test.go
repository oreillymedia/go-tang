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

var cli *redis.Client

func TestTest(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test Suite")
}

var _ = BeforeSuite(func() {
	rand.Seed(time.Now().UTC().UnixNano())
	godotenv.Load(".env")
	cli = redis.NewClient(&redis.Options{
		Addr:    os.Getenv("REDIS_URL"),
		Network: "tcp",
	})
	gotang.Client = cli
})

var _ = BeforeEach(func() {
	cli.FlushDb()
})

var _ = Describe("Fetch", func() {

	It("caches response in redis", func() {
		block := func() (string, int, error) {
			v := strconv.Itoa(rand.Intn(50000))
			return v, 5, nil
		}
		fetchedValue, err1 := gotang.Fetch("mykey", block, 1)
		cachedValue, err2 := gotang.Fetch("mykey", block, 1)
		Expect(err1).To(BeNil())
		Expect(err2).To(BeNil())
		Expect(fetchedValue).To(Equal(cachedValue))
	})

	It("expires cache ttl returned in block", func() {
		block := func() (string, int, error) {
			v := strconv.Itoa(rand.Intn(50000))
			return v, 1, nil
		}
		fetchedValue1, err1 := gotang.Fetch("mykey", block, 5)
		cachedValue1, err2 := gotang.Fetch("mykey", block, 5)
		time.Sleep(time.Duration(2) * time.Second)
		fetchedValue2, err3 := gotang.Fetch("mykey", block, 5)
		Expect(err1).To(BeNil())
		Expect(err2).To(BeNil())
		Expect(err3).To(BeNil())
		Expect(fetchedValue1).To(Equal(cachedValue1))
		Expect(fetchedValue1).ToNot(Equal(fetchedValue2))
	})

	// it prevents dog pile effect

	// it expires after certain time

})
