package main

import (
	"testing"
	"net/http"
	"net/http/httptest"

	"github.com/go-redis/redis"
	"strconv"

	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type MySuite struct {
	client *redis.Client
	config Configuration
}

var _ = Suite(&MySuite{})

func (s *MySuite) SetUpTest(c *C) {

	var err error

	s.config, err = GetConfiguration()
	if err != nil {
		c.Fatalf("Configuration read error : " + err.Error())
	}

	s.client, err = RedisNewClient()
	if err != nil {
		c.Fatalf(err.Error())
	}
}

func (s *MySuite) TearDownTest(c *C) {

	s.client.FlushDB()
	s.client.Close()

}

func (s *MySuite) TestHandler(c *C) {

	client = s.client
	config = s.config

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		c.Fatal(err)
	}

	req.RemoteAddr = "8.8.8.8:8080"

	testRecorder := httptest.NewRecorder()
	handler := http.HandlerFunc(handler)

	handler.ServeHTTP(testRecorder, req)

	c.Assert(testRecorder.Code,Equals,http.StatusOK)

	c.Assert(testRecorder.Body.String(),Equals,`{"CountryName":"United States"}`)

	ipCountry, err := client.Get("8.8.8.8").Result()
	if err != nil {
		c.Fatalf(err.Error())
	}

	c.Assert(ipCountry,Equals,"United States")

}

func (s *MySuite) TestHandlerMass(c *C) {

	client = s.client
	config = s.config

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		c.Fatal(err)
	}

	var addr string

	for i := 0; i < 30; i++ {

		addr = "8.8.8." + strconv.Itoa(i)

		req.Header.Set("Origin", "http://"+addr+":8080")

		testRecorder := httptest.NewRecorder()
		handler := http.HandlerFunc(handler)

		handler.ServeHTTP(testRecorder, req)

		c.Assert(testRecorder.Code,Equals,http.StatusOK)

		c.Assert(testRecorder.Body.String(),Equals,`{"CountryName":"United States"}`)

		ipCountry, err := client.Get(addr).Result()
		if err != nil {
			c.Fatalf(err.Error())
		}

		c.Assert(ipCountry,Equals,"United States")
	}

}
