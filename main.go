package main

import (
	"log"
	"net/http"

	"github.com/go-redis/redis"
	"encoding/json"
	"time"
)

var (
	client      *redis.Client
	config      Configuration
	numProvider int
)

func main() {

	var err error

	config, err = GetConfiguration()
	if err != nil {
		log.Fatal("Configuration read error : " + err.Error())
	}

	if len(config.Providers) < 1 {
		log.Fatal("No ip info providers")
	}

	client, err = RedisNewClient()
	if err != nil {
		log.Fatal(err)
	}

	defer client.Close()

	http.HandleFunc("/", handler)
	log.Println("server listen 8081")
	log.Fatal(http.ListenAndServe(":8081", nil))

}

func handler(w http.ResponseWriter, r *http.Request) {

	//grab IP from request
	ipString, err := GetClientIPHelper(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//check if ip is in DB
	ipCountry, err := client.Get(ipString).Result()
	if err == redis.Nil {
		//key does not exist
		//request providers till success or quota limit
		for {
			numProvider = -1
			for i := 0; i < len(config.Providers); i++ {
				if checkProvider(i) == nil {
					numProvider = i
					break
				}
			}

			log.Println("Provider # ", numProvider)

			if numProvider == -1 {
				http.Error(w, "quota exceeded on all providers", http.StatusInternalServerError)
				return
			}

			//make request to provider
			ipCountry, err = getInfo(ipString)
			if err == nil {
				log.Println("succefully got info from provider")
				break
			}
		}

		//save ip to DB
		err = client.Set(ipString, ipCountry, time.Duration(config.ExpireTime)*time.Second).Err()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}


	//return result
	payload, _ := json.Marshal(struct {
		CountryName string
	}{ipCountry})
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(payload)
	return

	}
