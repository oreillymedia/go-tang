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
		value1, err1 := gotang.Fetch("mykey", block, 1)
		value2, err2 := gotang.Fetch("mykey", block, 1)
		Expect(err1).To(BeNil())
		Expect(err2).To(BeNil())
		Expect(value1).To(Equal(value2))
	})

	// it prevents dog pile effect

	// it expires after certain time

})
