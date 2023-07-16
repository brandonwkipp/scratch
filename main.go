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

func main() {
	flag.StringVar(&API_KEY, "key", "", "Your itchi.io API key")
	flag.Parse()

	if API_KEY == "" {
		log.Fatal("Usage: main.go -key \"YOUR_KEY\"")
	}

	client := http.Client{}
	var p Page
	pageRaw, _ := MakeRequest("GET", client, "profile/owned-keys?page=1", nil)

	err := json.Unmarshal(pageRaw, &p)
	if err != nil {
		log.Fatalln(err)
	}

	for i, v := range p.OwnedKeys {
		if i == 0 {
			var files GameFiles
			filesRaw, err := MakeRequest("GET", client, "games/"+strconv.Itoa(v.GameId)+"/uploads?download_key_id="+strconv.Itoa(v.Id), nil)
			if err != nil {
				log.Fatalln(err)
			}
			err = json.Unmarshal(filesRaw, &files)
			if err != nil {
				log.Fatalln(err)
			}

			var u Uuid
			uuidRaw, err := MakeRequest("POST", client, "games/"+strconv.Itoa(files.Uploads[0].Id)+"/download-sessions", DownloadSession{DownloadKeyId: v.Id})
			if err != nil {
				log.Fatalln(err)
			}
			err = json.Unmarshal(uuidRaw, &u)
			if err != nil {
				log.Fatalln(err)
			}

			DownloadFile(files.Uploads[0].FileName, "uploads/"+strconv.Itoa(files.Uploads[0].Id)+"/download?api_key="+API_KEY+"&download_key_id="+strconv.Itoa(v.Id)+"&uuid="+u.Uuid)
		}
	}
}

func DownloadFile(name string, url string) {
	resp, err := grab.Get("./"+name, "https://api.itch.io/"+url)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Download saved to", resp.Filename)
}

func MakeRequest(method string, client http.Client, url string, bodyReq interface{}) ([]byte, error) {
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

	return body, nil
}
