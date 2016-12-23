package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"strings"

	"github.com/johnnywidth/synopsis"
	"github.com/johnnywidth/synopsis/composer"
)

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

// Lock is global lock for update handler
var Lock bool

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
	F, ok := w.(http.Flusher)
	if !ok {
		return F, errors.New("Streaming unsupported")
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	return F, nil
}

// GetAllRepoHandler return list of all repositories
func GetAllRepoHandler(res http.ResponseWriter, req *http.Request) {
	config := PrepareConfig()

	j, err := json.Marshal(config.File.Repositories)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	res.Header().Set("Content-Type", "application/json")
	res.Write(j)
}

// AddRepoHandler add new repository to config file
func AddRepoHandler(res http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	repo := synopsis.Repository{}
	err = json.Unmarshal(body, &repo)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	urlRepo := strings.TrimSpace(repo.URL)
	typeRepo := strings.TrimSpace(repo.Type)
	originalURLRepo := req.PostForm.Get("original_url")

	config := PrepareConfig()

	isExist := false
	for key, value := range config.File.Repositories {
		if value.URL == originalURLRepo {
			isExist = true
			config.File.Repositories[key].URL = urlRepo
			config.File.Repositories[key].Type = typeRepo
		}
	}
	if !isExist {
		config.File.Repositories = append(config.File.Repositories, synopsis.Repository{Type: typeRepo, URL: urlRepo})
	}

	err = updateConfigFile(config)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
}

// DeleteRepoHandler delete repository from config file
func DeleteRepoHandler(res http.ResponseWriter, req *http.Request) {
	url := req.URL.Query().Get("url")

	config := PrepareConfig()

	isExist := false
	for key, value := range config.File.Repositories {
		if value.URL == url {
			isExist = true
			config.File.Repositories = append(config.File.Repositories[:key], config.File.Repositories[key+1:]...)
		}
	}
	if !isExist {
		http.Error(res, "Item not exist!", http.StatusNoContent)
		return
	}

	err := updateConfigFile(config)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
}

func updateConfigFile(config synopsis.Config) error {
	j, err := json.MarshalIndent(config.File, "", "  ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(config.FileName, j, 0755)
	if err != nil {
		return err
	}
	return nil
}
