package main

import (
	"context"
	"flag"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

//go:generate sh -c "cp ../templates/consul.go ./consul.gen.go"

func main() {
	consul := flag.String("consul", "consul:8500", "Consul host")
	port := flag.Int("port", 8098, "this service port")
	flag.Parse()

	hostname, _ := os.Hostname()
	log.Println("Starting up... ", hostname, " consul host", *consul, " service  ", *port)

	// Register Service
	id := fmt.Sprintf("greeting-%v-%v", hostname, *port)
	consulClient, _ := NewConsulClient(*consul)
	health := fmt.Sprintf("http://%v:%v/api/greeting/v1/health", hostname, *port)
	consulClient.Register(id, "greeting-service", hostname, *port, "/api/greeting", health)

	router := mux.NewRouter().StrictSlash(true)

	// Define Health Endpoint
	router.Methods("GET").Path("/api/greeting/v1/health").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		str := fmt.Sprintf("{ 'status':'ok', 'host':'%v:%v' }", hostname, *port)
		fmt.Fprintf(w, str)
		log.Println("/api/greeting/v1/health called")
	})

	// The Hello endpoint for the greeting service
	router.Methods("GET").Path("/api/greeting/v1/hello/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		str := fmt.Sprintf("Hello, %q at %v:%v\n", html.EscapeString(r.URL.Path), hostname, *port)
		rt := rand.Intn(100)
		time.Sleep(time.Duration(rt) * time.Millisecond)
		fmt.Fprintf(w, str)
		log.Println(str)

		url, _ := consulClient.ServiceAddress("user-service")

		client := &http.Client{}
		resp, err := client.Get(url + "/api/users/v1/hello")
		if err != nil {
			fmt.Fprintf(w, "Error. %s", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			bodyBytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatal(err)
			}
			bodyString := string(bodyBytes)
			fmt.Fprintln(w, bodyString)
		}
	})

	server := &http.Server{
		Addr:         fmt.Sprintf(":%v", *port),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      router,
	}

	// De-register service at shutdown.
	c := make(chan os.Signal)
	signal.Notify(c, os.Kill, os.Interrupt, syscall.SIGTERM)
	go func() {
		for sig := range c {
			log.Println("Shutting Down...", sig)
			consulClient.DeRegister(id)
			server.Shutdown(context.Background())
			log.Println("Done...Bye")
			os.Exit(0)
		}
	}()

	log.Fatal(server.ListenAndServe())
}
