package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/cavaliergopher/grab/v3"
)

type DownloadSession struct {
	DownloadKeyId int `json:"download_key_id"`
}

type Game struct {
	Title string `json:"title"`
}

type GameFile struct {
	FileName string `json:"filename"`
	Id       int    `json:"id"`
}

type GameFiles struct {
	Uploads []GameFile `json:"uploads"`
}

type OwnedKey struct {
	Game   Game `json:"game"`
	GameId int  `json:"game_id"`
	Id     int  `json:"id"`
}

type Page struct {
	PerPage   int        `json:"per_page"`
	OwnedKeys []OwnedKey `json:"owned_keys"`
}

type Uuid struct {
	Uuid string `json:"uuid"`
}

var API_KEY string
var DownloadLocation string

func main() {
	flag.StringVar(&DownloadLocation, "downloadLocation", "", "The location to download your files")
	flag.StringVar(&API_KEY, "key", "", "Your itch.io API key")
	flag.Parse()

	if API_KEY == "" {
		log.Fatal("Usage: main.go -key \"YOUR_KEY\"")
	}

	client := http.Client{}
	var p Page
	pageRaw := MakeRequest("GET", client, "profile/owned-keys?page=1", nil)
	err := json.Unmarshal(pageRaw, &p)
	if err != nil {
		log.Fatalln(err)
	}

	for _, v := range p.OwnedKeys {
		var files GameFiles
		filesRaw := MakeRequest("GET", client, "games/"+strconv.Itoa(v.GameId)+"/uploads?download_key_id="+strconv.Itoa(v.Id), nil)
		err = json.Unmarshal(filesRaw, &files)
		if err != nil {
			log.Fatalln(err)
		}

		var location string
		if DownloadLocation != "" {
			location = filepath.Join(DownloadLocation, files.Uploads[0].FileName)
		} else {
			location = files.Uploads[0].FileName
		}

		if FileExists(location) {
			fmt.Printf("%s exists. Skipping download.\n", location)
		} else {
			var u Uuid
			uuidRaw := MakeRequest("POST", client, "games/"+strconv.Itoa(files.Uploads[0].Id)+"/download-sessions", DownloadSession{DownloadKeyId: v.Id})
			err = json.Unmarshal(uuidRaw, &u)
			if err != nil {
				log.Fatalln(err)
			}

			DownloadFile(location, "uploads/"+strconv.Itoa(files.Uploads[0].Id)+"/download?api_key="+API_KEY+"&download_key_id="+strconv.Itoa(v.Id)+"&uuid="+u.Uuid)
		}
	}
}

func DownloadFile(location string, url string) {
	resp, err := grab.Get(location, "https://api.itch.io/"+url)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Download saved to", resp.Filename)
}

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func MakeRequest(method string, client http.Client, url string, bodyReq interface{}) []byte {
	var requestBody io.Reader

	if bodyReq != nil {
		marshalled, err := json.Marshal(bodyReq)
		if err != nil {
			log.Fatalf("impossible to marshall body request: %s", err)
		}

		bytes.NewReader(marshalled)
	}

	req, err := http.NewRequest(method, "https://api.itch.io/"+url, requestBody)
	log.Println("https://api.itch.io/" + url)
	if err != nil {
		log.Fatalln(err)
	}

	req.Header = http.Header{
		"Authorization": {API_KEY},
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	return body
}
