package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"

	"github.com/gorilla/mux"
	"github.com/johnnywidth/synopsis"
	"github.com/johnnywidth/synopsis/composer"
)

// Lock is global lock for update handler
var Lock bool

// RegisterPackageController register handlers
func RegisterPackageController(router *mux.Router) {
	router.HandleFunc("/package/all", AllPackagesHandler)
	router.HandleFunc("/package/update", PackageUpdateHandler)
}

// AllPackagesHandler return all packages from `packages.json` file
func AllPackagesHandler(w http.ResponseWriter, req *http.Request) {
	config := PrepareConfig()

	file, err := ioutil.ReadFile(path.Join(config.OutputDir, "packages.json"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(file)
}

// PackageUpdateHandler run update packages
func PackageUpdateHandler(w http.ResponseWriter, req *http.Request) {
	if !Lock {
		Lock = true
		defer func() {
			Lock = false
		}()

		config := PrepareConfig()
		config.MakeOutputDir()

		flag := make(chan bool, config.ThreadNumber)
		ch := make(chan composer.PackageJSON)

		for _, repo := range config.File.Repositories {
			go func(r synopsis.Repository) {
				flag <- true
				defer func() {
					<-flag
				}()

				p, err := r.UpdateAll(config)
				if err != nil {
					log.Println(err)
				}

				ch <- p
			}(repo)
		}

		var useEventStream bool
		f, err := setEventStream(w)
		if err == nil {
			useEventStream = true
		}

		pm := make(composer.PackageJSON, len(config.File.Repositories))

		for i := 0; i < len(config.File.Repositories); i++ {
			p := <-ch
			for k, v := range p {
				pm[k] = v
			}
			// Flush event stream
			if useEventStream {
				fmt.Fprintf(w, "id: %s \n", "update")
				fmt.Fprintf(w, "data: %d \n\n", i+1)
				f.Flush()
			}
		}

		p := composer.Packages{Package: pm}
		if err := p.WriteToFile(config.OutputDir); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func setEventStream(w http.ResponseWriter) (http.Flusher, error) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return flusher, errors.New("Streaming unsupported")
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	return flusher, nil
}
