package main

import (
	"bytes"
	"fmt"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"net/http"
	"os"
)

func Home(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Great!\n")
}

func Upload(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Upload\n")
	buf := new(bytes.Buffer)
	reader, err := r.MultipartReader()
	// Part1: Chunk Number
	// Part4: Total Size (bytes)
	// Part6: File Name
	// Part8: Total Chunks
	// Part9: Chunk Data
	if err != nil {
		return err
	}
}

func ChunkRecieved(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Valid\n")
	chunkDirPath := "./tmp/chunks/" + r.FormValue("flowIdentifier") + "/" + r.FormValue("flowFilename") + ".part" + r.FormValue("flowChunkNumber")
	if _, err := os.Stat(chunkDirPath); err != nil {
		w.WriteHeader(204)
		return
	}
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/", Home)
	router.HandleFunc("/upload", Upload).Methods("POST")
	router.HandleFunc("/upload", ChunkRecieved).Methods("GET")

	n := negroni.Classic()
	//n.User(NewMiddleware)
	n.UseHandler(router)
	n.Run(":3000")
}
