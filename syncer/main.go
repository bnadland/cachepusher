package main

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"gopkg.in/redis.v3"
	"log"
	"os"
	"time"
)

func main() {
	config_dsn := os.Getenv("CP_DSN")
	if config_dsn == "" {
		config_dsn = "postgres://cachepusher:cachepusher@10.10.42.23:5432/cachepusher?sslmode=disable"
	}

	config_redis := os.Getenv("CP_REDIS")
	if config_redis == "" {
		config_redis = "10.10.42.23:6379"
	}

	config_cacheprefix := os.Getenv("CP_PREFIX")
	if config_cacheprefix == "" {
		config_cacheprefix = "customer"
	}

	db, err := sqlx.Connect("postgres", config_dsn)
	if err != nil {
		log.Print("[Postgresql] ", err)
		return
	}

	r := redis.NewClient(&redis.Options{
		Addr: config_redis,
	})
	_, err = r.Ping().Result()
	if err != nil {
		log.Print("[Redis] ", err)
		return
	}

	log.Print("Clearing cache")
	keys, err := r.Keys(fmt.Sprintf("%s:*", config_cacheprefix)).Result()
	if err != nil {
		log.Print(err)
	}
	r.Pipelined(func(r *redis.Pipeline) error {
		for _, key := range keys {
			err = r.Del(key).Err()
			if err != nil {
				log.Print(err)
			}
		}
		return nil
	})

	listener := pq.NewListener(config_dsn, 10*time.Second, time.Minute, func(ev pq.ListenerEventType, err error) {
		if err != nil {
			log.Print(err)
		}
	})

	err = listener.Listen("customer_updated")
	if err != nil {
		log.Print(err)
		return
	}
	err = listener.Listen("customer_deleted")
	if err != nil {
		log.Print(err)
		return
	}

	/** Wait until we have set up the listener to get notifications before we trigger the warmup **/
	log.Print("Triggering cache warmup")
	_, err = db.Exec("select customer_warmup()")
	if err != nil {
		log.Print(err)
	}

	log.Printf("Listening for updates")
	for {
		select {
		case n := <-listener.Notify:
			cachekey := fmt.Sprintf("%s:%s", config_cacheprefix, n.Extra)
			switch n.Channel {
			case "customer_deleted":
				log.Printf("DEL %s", cachekey)
				err = r.Del(cachekey).Err()
				if err != nil {
					log.Print(err)
				}
			case "customer_updated":
				var customerJson string
				err = db.Get(&customerJson, "select customer_get($1)", n.Extra)
				if err != nil {
					log.Print(err)
				}
				log.Printf("SET %s %s", cachekey, customerJson)
				err = r.Set(cachekey, customerJson, 0).Err()
				if err != nil {
					log.Print(err)
				}
			}
		// Make sure our connection stays up
		case <-time.After(90 * time.Second):
			log.Print("LISTEN PING")
			go func() {
				err = listener.Ping()
				if err != nil {
					log.Print(err)
				}
			}()
		}
	}
}
