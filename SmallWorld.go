package main

import (
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/flowg/go_experiences/server/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/flowg/go_experiences/server/Godeps/_workspace/src/github.com/gorilla/mux"
	"github.com/flowg/go_experiences/server/Godeps/_workspace/src/github.com/spf13/viper"
	"github.com/flowg/go_experiences/server/Godeps/_workspace/src/gopkg.in/fsnotify.v1"
	"github.com/flowg/go_experiences/server/Godeps/_workspace/src/gopkg.in/redis.v3"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

/*
 * ToDo
 * 4) Read Docker doc and deal with it
 * 6) Write a readme file
 */

var redisClient *redis.Client

/*
 * Helper functions
 */

func sendJSON(w http.ResponseWriter, m map[string]string) {
	msg, err := json.Marshal(m)
	if err != nil {
		log.WithFields(log.Fields{
			"data": m,
		}).Error("Error while trying to encode JSON")

		w.Write([]byte("Error while trying to encode JSON"))
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Write(msg)
	}
}

func generateRandomString(n int, safety int) (string, error) {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, n)
	for i := 0; i < n; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	res := string(result)
	safety++

	// We try again until we're sure we have a unique token or we hit the limit of possible combinations
	if redisClient.Exists(res).Val() && safety <= n*len(chars) {
		res, _ = generateRandomString(n, safety)
	} else if safety > n*len(chars) {
		errMsg := "All combinations exhausted. Please increase the token_size parameter in config file"
		log.Error(errMsg)
		err := errors.New(errMsg)
		return "", err
	}

	return res, nil
}

func generateCustomRandToken(custom string, safety int) (string, error) {
	size := viper.GetInt("custom_rand_size")
	token := custom + strconv.Itoa(rand.Intn(size))
	safety++

	// We try again until we're sure we have a unique token
	if redisClient.Exists(token).Val() && safety <= size {
		token, _ = generateCustomRandToken(custom, safety)
	} else if safety > size {
		errMsg := "All combinations exhausted for this customization value"
		log.WithFields(log.Fields{
			"custom": custom,
		}).Error(errMsg)
		err := errors.New(errMsg)
		return "", err
	}

	return token, nil
}

func createShortLink(w http.ResponseWriter, url string, custom string) (string, error) {
	var (
		token = ""
		err   error
	)

	if custom == "" {
		token, err = generateRandomString(viper.GetInt("token_size"), 0)
	} else {
		// Checking if custom value is already a key
		if redisClient.Exists(custom).Val() {
			token, err = generateCustomRandToken(custom, 0)
		} else {
			token, err = custom, nil
		}
	}

	if err != nil {
		return "", err
	}

	// Now we have a unique token, let's store it with expiration in 3 months
	redisClient.HMSet(token, "creation_timestamp", strconv.Itoa(int(time.Now().Unix())), "origin", url, "token", token, "count", strconv.Itoa(0))
	redisClient.Expire(token, time.Hour*24*90)
	log.WithFields(log.Fields{
		"creation_timestamp": time.Now().Unix(),
		"origin":             url,
		"token":              token,
		"count":              0,
	}).Info("A new shortlink has been created")

	return token, nil
}

func factorizedHandler(w http.ResponseWriter, r *http.Request, f func(token string)) {
	// Get back token from request
	vars := mux.Vars(r)
	token := vars["token"]

	// Check if it exists in Redis
	if redisClient.Exists(token).Val() {
		f(token)
	} else {
		log.WithFields(log.Fields{
			"token": token,
		}).Error("This shortlink doesn't exist")

		var m = map[string]string{
			"Error": "true",
			"Msg":   "This shortlink doesn't exist",
			"token": token,
		}

		sendJSON(w, m)
	}
}

/*
 * Handlers for each route
 */

func shortLinkCreationHandler(w http.ResponseWriter, r *http.Request) {
	// Dealing with payload
	url := r.PostFormValue("url")
	custom := r.PostFormValue("custom")

	// Checking if URL exists
	res, err := http.Head(url)
	if err != nil || res.StatusCode == 404 {
		// URL doesn't exist, we log it and return a JSON explaining that
		log.WithFields(log.Fields{
			"url": url,
		}).Error("Invalid URL was provided to the API")

		var m = map[string]string{
			"Error": "true",
			"Msg":   "Invalid URL was provided to the API",
			"URL":   url,
		}

		sendJSON(w, m)
	} else {
		// URL exists so we generate a token and return a JSON
		token, err := createShortLink(w, url, custom)
		var m map[string]string

		if err != nil {
			m = map[string]string{
				"Error": "true",
				"Msg":   err.Error(),
			}
		} else {
			m = map[string]string{
				"origin": url,
				"token":  token,
			}
		}

		sendJSON(w, m)
	}
}

func redirectionHandler(w http.ResponseWriter, r *http.Request) {
	factorizedHandler(w, r, func(token string) {
		// Increment, log and redirect
		redisClient.HIncrBy(token, "count", 1)

		redirectTo := redisClient.HGet(token, "origin").Val()
		log.WithFields(log.Fields{
			"token":        token,
			"redirectedTo": redirectTo,
			"count":        redisClient.HGet(token, "count").Val(),
		}).Info("This shortlink has been used for a redirection")

		http.Redirect(w, r, redirectTo, 303)
	})
}

func monitoringHandler(w http.ResponseWriter, r *http.Request) {
	factorizedHandler(w, r, func(token string) {
		m := redisClient.HGetAllMap(token).Val()

		log.WithFields(log.Fields{
			"token": m,
		}).Info("A monitoring has been asked for this shortlink")

		sendJSON(w, m)
	})
}

func main() {
	// Setting Viper to access config file
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	// Setting logrus
	f, err := os.OpenFile(viper.GetString("log_file"), os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.SetOutput(os.Stderr)
	} else {
		defer f.Close()

		log.SetOutput(f)
	}
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.InfoLevel)

	// Setting Viper to watch config file so that we never have to shutdown the server
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.WithFields(log.Fields{
			"event": e.Name,
		}).Info("Config file changed")
	})

	// Setting Redis client
	redisClient = redis.NewClient(&redis.Options{
		Addr:     viper.GetString("datastore.host") + ":" + viper.GetString("datastore.port"),
		Password: viper.GetString("datastore.pwd"),
		DB:       int64(viper.GetInt("datastore.db_name")),
	})

	// Checking Redis server is available
	pingError := redisClient.Ping().Err()
	if pingError != nil {
		log.Fatal("Redis server is unavailable")
	}

	// Setting router
	r := mux.NewRouter()

	// Route for short link creation
	r.HandleFunc("/shortlink", shortLinkCreationHandler).Methods("POST")

	// Route for redirection
	r.HandleFunc("/{token}", redirectionHandler).Methods("GET")

	// Route to read monitoring
	r.HandleFunc("/admin/{token}", monitoringHandler).Methods("GET")

	// Bind to a port and pass our router in
	http.ListenAndServe(":"+viper.GetString("port"), r)
}
