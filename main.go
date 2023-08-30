package main

import (
	"io"
	"log"
	"net/http"
	"os"
)

const MaxChunkSize = 1024

var ResourceMap = make(map[string]string)

type resourceHandler struct {
}

func (r resourceHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	log.Println("Received request.")
	var err error
	if request.Method == http.MethodPut {
		err = r.putResource(writer, request)
	} else if request.Method == http.MethodGet {
		err = r.getResource(writer, request)
	}

	if err != nil {
		log.Printf("%s\n", err)
	}

}

func (r resourceHandler) getResource(writer http.ResponseWriter, request *http.Request) error {
	path := ResourceMap[request.URL.Path]
	if path == "" {
		http.NotFound(writer, request)
		return nil
	}
	file, err := os.Open(path)
	defer closeStream(file)
	if err != nil {
		return err
	}

	err = copyFile(file, writer)
	if err != nil {
		return err
	}

	return nil
}

//goland:noinspection GoUnusedParameter
func (r resourceHandler) putResource(writer http.ResponseWriter, request *http.Request) error {
	log.Println("Received PUT request")
	data := request.Body
	defer closeStream(data)

	file, err := os.CreateTemp("", "")
	defer closeStream(file)
	if err != nil {
		return err
	}
	err = copyFile(data, file)
	if err != nil {
		return err
	}
	log.Printf("Wrote data to %s\n", file.Name())

	ResourceMap[request.URL.Path] = file.Name()

	return nil
}

func copyFile(src io.Reader, dst io.Writer) error {
	buffer := make([]byte, MaxChunkSize)
	for {
		_, readErr := src.Read(buffer)
		_, writeErr := dst.Write(buffer)
		if writeErr != nil {
			return writeErr
		}
		if readErr == io.EOF {
			break
		}
	}
	return nil
}

func closeStream(data io.ReadCloser) {
	err := data.Close()
	if err != nil {
		log.Printf("Unexpected error when attempting to close stream: %s", err)
	}
}

func main() {
	http.Handle("/resources/", new(resourceHandler))
	addr := "localhost:8080"
	log.Printf("Hosting server at %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
