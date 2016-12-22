package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/johnnywidth/synopsis"
)

// RegisterRepoController register handlers
func RegisterRepoController(router *mux.Router) {
	router.Methods("GET").Path("/repo").HandlerFunc(GetAllRepoHandler)
	router.Methods("POST").Path("/repo").HandlerFunc(AddRepoHandler)
	router.Methods("DELETE").Path("/repo").HandlerFunc(DeleteRepoHandler)
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
