package main

import (
	"log"
	"errors"
	"net"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"encoding/json"
	"github.com/go-redis/redis"
	"sort"
	"time"
	"io/ioutil"
)

// GetClientIPHelper gets the client IP using a mixture of techniques.

func GetClientIPHelper(req *http.Request) (ipResult string, errResult error) {

	//  Try Request Header ("Origin")
	clientURL, err := url.Parse(req.Header.Get("Origin"))
	if err == nil {
		host := clientURL.Host
		ip, _, err := net.SplitHostPort(host)
		if err == nil {
			log.Printf("debug: Found IP using Header (Origin) sniffing. ip: %v", ip)
			return ip, nil
		}
	}

	// Try by Request
	ip, err := getClientIPByRequestRemoteAddr(req)
	if err == nil {
		log.Printf("debug: Found IP using Request sniffing. ip: %v", ip)
		return ip, nil
	}

	// Try Request Headers (X-Forwarder). Client could be behind a Proxy
	ip, err = getClientIPByHeaders(req)
	if err == nil {
		log.Printf("debug: Found IP using Request Headers sniffing. ip: %v", ip)
		return ip, nil
	}

	err = errors.New("error: Could not find clients IP address")
	return "", err
}

// getClientIPByRequest tries to get directly from the Request.
// https://blog.golang.org/context/userip/userip.go
func getClientIPByRequestRemoteAddr(req *http.Request) (ip string, err error) {

	// Try via request
	ip, port, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		log.Printf("debug: Getting req.RemoteAddr %v", err)
		return "", err
	} else {
		log.Printf("debug: With req.RemoteAddr found IP:%v; Port: %v", ip, port)
	}

	userIP := net.ParseIP(ip)
	if userIP == nil {
		message := fmt.Sprintf("debug: Parsing IP from Request.RemoteAddr got nothing.")
		log.Printf(message)
		return "", fmt.Errorf(message)

	}
	log.Printf("debug: Found IP: %v", userIP)
	return userIP.String(), nil

}

// getClientIPByHeaders tries to get directly from the Request Headers.
// This is only way when the client is behind a Proxy.
func getClientIPByHeaders(req *http.Request) (ip string, err error) {

	// Client could be behid a Proxy, so Try Request Headers (X-Forwarder)
	ipSlice := make([]string, 0)

	ipSlice = append(ipSlice, req.Header.Get("X-Forwarded-For"))
	ipSlice = append(ipSlice, req.Header.Get("x-forwarded-for"))
	ipSlice = append(ipSlice, req.Header.Get("X-FORWARDED-FOR"))

	for _, v := range ipSlice {
		log.Printf("debug: client request header check gives ip: %v", v)
		if v != "" {
			return v, nil
		}
	}
	err = errors.New("error: Could not find clients IP address from the Request Headers")
	return "", err

}

type Provider struct {
	PreReqURL       string
	PostReqURL      string
	KeysInResponce  []string
	MaxReqPerMinute int
	ReqTimes        []time.Time
}

type Configuration struct {
	ExpireTime int //seconds
	Providers  []Provider
}

func GetConfiguration() (config Configuration, err error) {

	file, err := os.Open("./configuration.json")
	if err != nil {
		return
	}

	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return
	}
	return
}

func RedisNewClient() (client *redis.Client, err error) {
	client = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	_, err = client.Ping().Result()
	return
}

func checkProvider(num int) (err error) {

	if len(config.Providers[num].ReqTimes) < config.Providers[num].MaxReqPerMinute {
		return nil
	}

	//get index for minute interval
	i := sort.Search(len(config.Providers[num].ReqTimes),
		func(i int) bool {
			return config.Providers[num].ReqTimes[i].After(time.Now().Add(time.Minute * (-1)))

		})

	config.Providers[num].ReqTimes = config.Providers[num].ReqTimes[i:]

	if len(config.Providers[num].ReqTimes) < config.Providers[num].MaxReqPerMinute {
		return nil
	}

	return fmt.Errorf("quota exceeded")
}

func parseResult(raw []byte, num int) (country string, err error) {

	var f interface{}
	err = json.Unmarshal(raw, &f)
	if err != nil {
		return
	}

	for i := 0; i < len(config.Providers[num].KeysInResponce); i++ {

		m := f.(map[string]interface{})
		r, ok := m[config.Providers[num].KeysInResponce[i]]
		if !ok {
			err = fmt.Errorf("no such key : %s at %s", config.Providers[num].KeysInResponce[i],config.Providers[num].PreReqURL)
			return
		}

		f = r

	}

	return f.(string), nil

}

func getInfo(ip string) (country string, err error) {

	rs, err := http.Get(config.Providers[numProvider].PreReqURL + ip + config.Providers[numProvider].PostReqURL)
	//process response
	if err != nil {
		return
	}
	defer rs.Body.Close()

	//commit successful take
	config.Providers[numProvider].ReqTimes = append(config.Providers[numProvider].ReqTimes, time.Now())

	bodyBytes, err := ioutil.ReadAll(rs.Body)
	if err != nil {
		return
	}

	country, err = parseResult(bodyBytes, numProvider)
	if err != nil {
		return
	}
	return
}