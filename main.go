package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"file-system.com/file-system/util"
)

var allFiles []string
var mtx sync.Mutex

func handleExceptions(sig os.Signal) {
	os.Chdir("..")
	os.RemoveAll("Static-Dir")
	os.Exit(0)
}

func main() {
	os.Mkdir("Static-Dir", os.ModePerm)
	os.Chdir("Static-Dir")

	// defer os.RemoveAll("Static-Dir")
	sigChan := make(chan os.Signal, 1)

	signal.Notify(sigChan, syscall.SIGINT)

	go handleRequests()

	sig := <-sigChan

	handleExceptions(sig)

	handleRequests()
}

func handleRequests() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			sendFiles(w, r)

		case http.MethodPost:
			writeFile(w, r)

		case http.MethodPut:
			writeFile(w, r)

		case http.MethodDelete:
			removeFiles(w, r)
		
		case "LIST":
			returnFiles(w, r)

		}
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func sendFiles(w http.ResponseWriter, r *http.Request) {
	fileName := r.Header.Get("File-Name")

	file, err := os.Open(fileName)

	if err != nil {
		w.Write([]byte("File Doesn't exist"))
		return
	}

	defer file.Close()

	for {
		buffer := make([]byte, 4096)
		_, err := file.Read(buffer)
		if err == io.EOF {
			break
		}
		w.Write(buffer)
	}
}

func writeFile(w http.ResponseWriter, r *http.Request) {
	fileName := r.Header.Get("File-Name")
	
	file, err := os.Create(fileName)

	if err != nil {
		os.Exit(1)
	}

	defer file.Close()

	for {
		buffer := make([]byte, 4096)
		_, err := io.ReadFull(r.Body, buffer)
		if err == io.EOF {
			break
		}
		file.Write(buffer)
	}

	mtx.Lock()
	if util.FileIndex(allFiles, fileName) == -1 {
		allFiles = append(allFiles, fileName)
	}
	mtx.Unlock()
}

func removeFiles(w http.ResponseWriter, r *http.Request) {
	fileName := r.Header.Get("File-Name")
	mtx.Lock()
	fileIdx := util.FileIndex(allFiles, fileName)
	if fileIdx == -1 {
		mtx.Unlock()
		w.Write([]byte("Doesn't exist"))
		return
	}

	os.Remove(fileName)

	if fileIdx != -1 {
		allFiles = append(allFiles[:fileIdx], allFiles[fileIdx+1:]...)
	}
	mtx.Unlock()
}

func returnFiles(w http.ResponseWriter, r *http.Request) {
	mtx.Lock()
	for i := 0; i < len(allFiles); i++ {
		_, err := fmt.Fprintln(w, allFiles[i])
		if err != nil {
			fmt.Println(err.Error())
		}
	}
	mtx.Unlock()
}