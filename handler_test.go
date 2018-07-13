package main

import (
	"testing"
	"net/http"
	"net/http/httptest"
	"log"
	"github.com/go-redis/redis"
	"strconv"
)

func TestHandler(t *testing.T) {

	var err error

	config, err = GetConfiguration()
	if err != nil {
		log.Fatal("Configuration read error : " + err.Error())
	}

	client, err = RedisNewClient()
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.RemoteAddr = "8.8.8.8:8080"

	testRecorder := httptest.NewRecorder()
	handler := http.HandlerFunc(handler)

	handler.ServeHTTP(testRecorder, req)

	if status := testRecorder.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := `{"CountryName":"United States"}`
	if testRecorder.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			testRecorder.Body.String(), expected)
	}

	ipCountry, err := client.Get("8.8.8.8").Result()
	if err == redis.Nil {
		t.Error("don't persist in Redis")
	} else if err != nil {
		t.Error("Redis error %v", err)
	}

	if ipCountry != "United States" {
		t.Errorf("handler returned unexpected result from Redis: got %v want %v", ipCountry, "United States")
	}

	client.FlushDB()
	client.Close()

}

func TestHandlerMass(t *testing.T) {

	var err error

	config, err = GetConfiguration()
	if err != nil {
		log.Fatal("Configuration read error : " + err.Error())
	}

	config.Providers[0].MaxReqPerMinute = 2

	client, err = RedisNewClient()
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	var addr string

	for i := 0; i < 30; i++ {

		addr = "8.8.8."+strconv.Itoa(i)

		req.Header.Set("Origin", "http://"+addr+":8080")

		testRecorder := httptest.NewRecorder()
		handler := http.HandlerFunc(handler)

		handler.ServeHTTP(testRecorder, req)

		if status := testRecorder.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}

		expected := `{"CountryName":"United States"}`
		if testRecorder.Body.String() != expected {
			t.Errorf("handler returned unexpected body: got %v want %v",
				testRecorder.Body.String(), expected)
		}

		ipCountry, err := client.Get(addr).Result()
		if err == redis.Nil {
			t.Error("don't persist in Redis")
		} else if err != nil {
			t.Error("Redis error %v", err)
		}

		if ipCountry != "United States" {
			t.Errorf("handler returned unexpected result from Redis: got %v want %v", ipCountry, "United States")
		}
	}

	client.FlushDB()
	client.Close()
}
