package main

import (
	"encoding/json"
	"fmt"
	driveService "gdrive/driveservice"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	// fmt.Printf("Go to the following link in your browser then type the "+
	// 	"authorization code: \n%v\n", authURL)

	err := exec.Command("xdg-open", authURL).Start()
	if err != nil {
		log.Fatal("could not retrieve the token")
	}

	fmt.Println("Enter the authorization code here once done :")

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

var wg sync.WaitGroup

func main() {
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, drive.DriveScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)
	srv, err := (*driveService.FileService).GetInstance(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Drive client: %v", err)
	}

	r, err := srv.Files.List().Fields("nextPageToken, files(id, name)").Do()
	if err != nil {
		log.Fatalf("Unable to retrieve files: %v", err)
	}
	fmt.Printf("Length of the files are:%v \n", r.NextPageToken)
	var filenames []string
	err = filepath.Walk("/home/anubhav/Projects/backup/MoviesShows/",
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {

				filenames = append(filenames, path)
			}

			return nil
		})
	if err != nil {
		log.Println(err)
	}
	//uploadFile(srv, "1conPh8WTgBqzn5BSHJAWrmuv2Sv4cRg8", "/home/anubhav/Projects/backup/MoviesShows/Silicon Valley - Season 1 - 720p BluRay - x265 HEVC - ShAaNiG/Silicon.Valley.S01E01.720p.BluRay.x265.ShAaNiG.mkv")
	//uploadFileBatched(srv, "1Dhg2f9vtrG9PPsKvugNoTvq2vckXamjx", 5, &filenames)
	fmt.Println("Files:")
repeat:
	if len(r.Files) == 0 {
		fmt.Println("No files found.")
	} else {
		for _, i := range r.Files {

			fmt.Printf("name is: %s The ID is: %s\n ", i.Name, i.Id)
			if len(i.Name) > 5 {
				if i.Name[:5] == "/home" {
					fmt.Printf("Deleting file %v \n", i.Name)
					err := srv.Files.Delete(i.Id).Do()
					fmt.Println("Deleted file")
					if err != nil {
						log.Fatalf("cannot delete file the error is: %v\n", err)
					}
				}
			}

		}
	}
	if r.NextPageToken != "" {
		r, err = srv.Files.List().PageToken(r.NextPageToken).Do()
		goto repeat
	}
}

func deleteFileBatched(srv *drive.Service, prefix string, batchSize int) {
	panic("not implemented")

}

func uploadFileBatched(srv *drive.Service, parentID string, batchSize int, fileNames *[]string) {
	for i, file := range *fileNames {
		wg.Add(1)
		go uploadFile(srv, parentID, file)
		if (i)%batchSize == 0 {
			wg.Wait()
		}
	}

	wg.Wait()
}

func uploadFile(srv *drive.Service, parentID string, fileName string) {

	up, err := os.Open(fileName)
	if err != nil {
		log.Printf("Unable to read file: %v", err)
	}
	defer up.Close()
	ssz, _ := up.Stat()
	fmt.Printf("uploading file %v , Size is  %v \n", ssz.Name(), ssz.Size())
	file := drive.File{Name: up.Name(), Parents: []string{parentID}}

	_, err = srv.Files.Create(&file).ResumableMedia(nil, up, ssz.Size(), "").ProgressUpdater(call).Do()
	if err != nil {
		log.Fatalf("Unable to upload file: %v", err)
	}
	wg.Done()
}
func call(current, total int64) {
	fmt.Printf("Percent uploaded: %v \n", ((current * 100) / total))
}
