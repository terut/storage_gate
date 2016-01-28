package main

import (
	"./store"

	"bytes"
	"fmt"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

func Home(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Great!\n")
}

func Upload(w http.ResponseWriter, r *http.Request) {
	buf := new(bytes.Buffer)
	reader, err := r.MultipartReader()
	// Part1: Chunk Number
	// Part4: Total Size (bytes)
	// Part6: File Name
	// Part8: Total Chunks
	// Part9: Chunk Data
	if err != nil {
		//return err
		return
	}

	part, err := reader.NextPart() // Part1
	if err != nil {
		//return err
		return
	}
	io.Copy(buf, part)
	token := buf.String()
	buf.Reset()

	part, err = reader.NextPart() // Part1
	if err != nil {
		//return err
		return
	}
	io.Copy(buf, part)
	dir := buf.String()
	buf.Reset()

	part, err = reader.NextPart() // Part1
	if err != nil {
		//return err
		return
	}
	io.Copy(buf, part)
	chunkNo := buf.String()
	buf.Reset()

	// Skip Part2, 3 and Use Part4
	for i := 0; i < 3; i++ {
		part, err = reader.NextPart()
		if err != nil {
			//return err
			return
		}
	}
	io.Copy(buf, part)
	flowTotalSize := buf.String()
	buf.Reset()

	// Skip Part5 and Use Part6
	for i := 0; i < 2; i++ {
		part, err = reader.NextPart()
		if err != nil {
			//return err
			return
		}
	}
	io.Copy(buf, part)
	fileName := buf.String()
	buf.Reset()

	// Skip Part7, 8 and Use Part9
	for i := 0; i < 3; i++ {
		part, err = reader.NextPart()
		if err != nil {
			//return err
			return
		}
	}
	chunkDirPath := "./tmp/chunks/" + fileName
	err = os.MkdirAll(chunkDirPath, 02750)
	if err != nil {
		//return err
		return
	}

	dst, err := os.Create(chunkDirPath + "/" + chunkNo)
	if err != nil {
		//return err
		return
	}
	defer dst.Close()
	io.Copy(dst, part)

	fileInfos, err := ioutil.ReadDir(chunkDirPath)
	if err != nil {
		//return err
		return
	}

	if flowTotalSize == strconv.Itoa(int(TotalSize(fileInfos))) {
		uploadFile := UploadFile{FilePath: chunkDirPath, Dir: dir, Token: token}
		completedFiles <- uploadFile
	}
	return
}

func TotalSize(fileInfos []os.FileInfo) int64 {
	var sum int64
	for _, fi := range fileInfos {
		sum += fi.Size()
	}
	return sum
}

func ChunkRecieved(w http.ResponseWriter, r *http.Request) {
	chunkDirPath := "./tmp/chunks/" + r.FormValue("flowIdentifier") + "/" + r.FormValue("flowFilename") + ".part" + r.FormValue("flowChunkNumber")
	if _, err := os.Stat(chunkDirPath); err != nil {
		w.WriteHeader(204)
		return
	}
}

type UploadFile struct {
	FilePath string
	Dir      string
	Token    string
}

//FIXME
var completedFiles = make(chan UploadFile, 100)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	for i := 0; i < 8; i++ {
		go assembleFile(completedFiles)
	}

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:8080"},
	})

	router := mux.NewRouter()
	router.HandleFunc("/", Home)
	router.HandleFunc("/upload", Upload).Methods("POST")
	router.HandleFunc("/upload", ChunkRecieved).Methods("GET")

	n := negroni.Classic()
	//n.User(NewMiddleware)
	n.Use(c)
	n.UseHandler(router)
	n.Run(":3000")
}

//For sort by filename(chunk number)
type ByChunk []os.FileInfo

func (a ByChunk) Len() int      { return len(a) }
func (a ByChunk) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByChunk) Less(i, j int) bool {
	ai, _ := strconv.Atoi(a[i].Name())
	aj, _ := strconv.Atoi(a[j].Name())
	return ai < aj
}

func assembleFile(jobs <-chan UploadFile) {
	for path := range jobs {
		assemble(path)
	}
}

func assemble(path UploadFile) {
	fileInfos, err := ioutil.ReadDir(path.FilePath)
	if err != nil {
		return
	}

	store := store.GetInstance()
	bucket := store.Buckets[path.Token]
	filename := strings.Split(path.FilePath, "/")[3]
	fileDirPath := "./public/static/" + bucket.Name + "/" + path.Dir

	err = os.MkdirAll(fileDirPath, 02750)
	if err != nil {
		//return err
		return
	}

	dst, err := os.Create(fileDirPath + "/" + filename)
	if err != nil {
		return
	}
	defer dst.Close()

	sort.Sort(ByChunk(fileInfos))
	for _, fs := range fileInfos {
		src, err := os.Open(path.FilePath + "/" + fs.Name())
		if err != nil {
			return
		}
		defer src.Close()
		io.Copy(dst, src)
	}
	os.RemoveAll(path.FilePath)
}
