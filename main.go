package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path"

	"github.com/gorilla/mux"
	"github.com/johnnywidth/synopsis"
)

func main() {
	PrepareConfig()

	router := mux.NewRouter().StrictSlash(true)

	RegisterRepoController(router)
	RegisterPackageController(router)

	outputDir := os.Getenv("OUTPUT") + "/"
	router.HandleFunc("/packages.json", func(res http.ResponseWriter, req *http.Request) {
		http.ServeFile(res, req, path.Join(outputDir, "packages.json"))
	})
	router.PathPrefix("/dist/").Handler(http.StripPrefix("/dist/", http.FileServer(http.Dir(path.Join(outputDir, "dist")))))
	router.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("./public/assets/"))))
	router.PathPrefix("/admin").Handler(http.StripPrefix("/admin", http.FileServer(http.Dir("./public/view/admin/"))))
	router.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./public/view/package/"))))

	if err := http.ListenAndServe(os.Getenv("HOST")+":"+os.Getenv("PORT"), router); err != nil {
		log.Fatal(err)
	}
}

// PrepareConfig prepare synopsis config
func PrepareConfig() synopsis.Config {
	var config synopsis.Config
	config.PrepareConfig(os.Getenv("FILE"), os.Getenv("OUTPUT"), os.Getenv("THREAD"))

	return config
}
