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
	ip string
	country string
}

var _ = Suite(&MySuite{
	ip : "8.8.1.",
	country : "United States",
})

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

	req.RemoteAddr = s.ip+"8:8080"

	testRecorder := httptest.NewRecorder()
	handler := http.HandlerFunc(handler)

	handler.ServeHTTP(testRecorder, req)

	c.Assert(testRecorder.Code,Equals,http.StatusOK)

	c.Assert(testRecorder.Body.String(),Equals,`{"CountryName":"`+s.country+`"}`)

	ipCountry, err := client.Get(s.ip+"8").Result()
	if err != nil {
		c.Fatalf(err.Error())
	}

	c.Assert(ipCountry,Equals,s.country)

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

		addr = s.ip + strconv.Itoa(i)

		req.Header.Set("Origin", "http://"+addr+":8080")

		testRecorder := httptest.NewRecorder()
		handler := http.HandlerFunc(handler)

		handler.ServeHTTP(testRecorder, req)

		c.Assert(testRecorder.Code,Equals,http.StatusOK)

		c.Assert(testRecorder.Body.String(),Equals,`{"CountryName":"`+s.country+`"}`)

		ipCountry, err := client.Get(addr).Result()
		if err != nil {
			c.Fatalf(err.Error())
		}

		c.Assert(ipCountry,Equals,s.country)
	}

}
