package cache

import (
	"log"
	"os"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/dgraph-io/badger/v4"
	"github.com/redis/go-redis/v9"
)

var testRedisCache RedisCache
var testBadgerCache BadgerCache

func TestMain(m *testing.M) {
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer s.Close()

	client := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})

	testRedisCache.Conn = client
	testRedisCache.Prefix = "test-celeritas"

	defer testRedisCache.Conn.Close()

	_ = os.RemoveAll("./testdata/tmp/badger")

	// create a badger database
	err = os.MkdirAll("./testdata/tmp/badger", 0755)
	if err != nil {
		log.Fatal(err)
	}

	db, _ := badger.Open(badger.DefaultOptions("./testdata/tmp/badger"))
	testBadgerCache.Conn = db

	os.Exit(m.Run())
}
