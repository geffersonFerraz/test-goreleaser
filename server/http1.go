package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type AllAssets struct {
	Name    string `json:"name"`
	TagName string `json:"version"`
	URL     string `json:"url"`
}

type Release struct {
	URL       string `json:"url"`
	AssetsURL string `json:"assets_url"`
	UploadURL string `json:"upload_url"`
	HTMLURL   string `json:"html_url"`
	ID        int    `json:"id"`
	Author    struct {
		Login             string `json:"login"`
		ID                int    `json:"id"`
		NodeID            string `json:"node_id"`
		AvatarURL         string `json:"avatar_url"`
		GravatarID        string `json:"gravatar_id"`
		URL               string `json:"url"`
		HTMLURL           string `json:"html_url"`
		FollowersURL      string `json:"followers_url"`
		FollowingURL      string `json:"following_url"`
		GistsURL          string `json:"gists_url"`
		StarredURL        string `json:"starred_url"`
		SubscriptionsURL  string `json:"subscriptions_url"`
		OrganizationsURL  string `json:"organizations_url"`
		ReposURL          string `json:"repos_url"`
		EventsURL         string `json:"events_url"`
		ReceivedEventsURL string `json:"received_events_url"`
		Type              string `json:"type"`
		SiteAdmin         bool   `json:"site_admin"`
	} `json:"author"`
	NodeID          string    `json:"node_id"`
	TagName         string    `json:"tag_name"`
	TargetCommitish string    `json:"target_commitish"`
	Name            string    `json:"name"`
	Draft           bool      `json:"draft"`
	Prerelease      bool      `json:"prerelease"`
	CreatedAt       time.Time `json:"created_at"`
	PublishedAt     time.Time `json:"published_at"`
	Assets          []struct {
		URL      string `json:"url"`
		ID       int    `json:"id"`
		NodeID   string `json:"node_id"`
		Name     string `json:"name"`
		Label    string `json:"label"`
		Uploader struct {
			Login             string `json:"login"`
			ID                int    `json:"id"`
			NodeID            string `json:"node_id"`
			AvatarURL         string `json:"avatar_url"`
			GravatarID        string `json:"gravatar_id"`
			URL               string `json:"url"`
			HTMLURL           string `json:"html_url"`
			FollowersURL      string `json:"followers_url"`
			FollowingURL      string `json:"following_url"`
			GistsURL          string `json:"gists_url"`
			StarredURL        string `json:"starred_url"`
			SubscriptionsURL  string `json:"subscriptions_url"`
			OrganizationsURL  string `json:"organizations_url"`
			ReposURL          string `json:"repos_url"`
			EventsURL         string `json:"events_url"`
			ReceivedEventsURL string `json:"received_events_url"`
			Type              string `json:"type"`
			SiteAdmin         bool   `json:"site_admin"`
		} `json:"uploader"`
		ContentType        string    `json:"content_type"`
		State              string    `json:"state"`
		Size               int       `json:"size"`
		DownloadCount      int       `json:"download_count"`
		CreatedAt          time.Time `json:"created_at"`
		UpdatedAt          time.Time `json:"updated_at"`
		BrowserDownloadURL string    `json:"browser_download_url"`
	} `json:"assets"`
	TarballURL string `json:"tarball_url"`
	ZipballURL string `json:"zipball_url"`
	Body       string `json:"body"`
}

var (
	GIT_LATEST_URL     = ""
	GIT_TAG_URL        = ""
	GIT_PERSONAL_TOKEN = ""
	GIT_HEADER_VERSION = ""
	GIT_OWNER_PROJECT  = ""
	SERVER_URL         = ""
)

func init() {
	SERVER_URL = getEnvOrThrow("SERVER_URL")
	GIT_OWNER_PROJECT = getEnvOrThrow("GIT_OWNER_PROJECT")

	GIT_LATEST_URL = "https://api.github.com/repos/" + GIT_OWNER_PROJECT + "/releases/latest"
	GIT_TAG_URL = "https://api.github.com/repos/" + GIT_OWNER_PROJECT + "/releases/tags/%s"
	GIT_HEADER_VERSION = "X-GitHub-Api-Version: 2022-11-28"

	GIT_PERSONAL_TOKEN = "Bearer " + getEnvOrThrow("GIT_PERSONAL_TOKEN")
}

type (
	Server interface {
		Listen()
		prepare()
	}

	server struct {
		url string
		mux *http.ServeMux
	}
)

func NewHttpServer(host string, port string) Server {
	url := fmt.Sprintf("%s:%s", host, port)
	mux := http.NewServeMux()
	return &server{url: url, mux: mux}
}

func (s *server) Listen() {
	s.prepare()
	log.Printf("running...")
	http.ListenAndServe(s.url, s.mux)
}

func (s *server) prepare() {
	s.mux.HandleFunc("GET /download/cli/{version}/{so}", func(w http.ResponseWriter, r *http.Request) {
		r = injectTime(r)
		run(w, r)
		printTime(r)
	})

	s.mux.HandleFunc("GET /download/cli/{version}", func(w http.ResponseWriter, r *http.Request) {
		r = injectTime(r)
		run(w, r)
		printTime(r)
	})
}

func run(w http.ResponseWriter, r *http.Request) {
	soRequested := r.PathValue("so")
	versionRequested := r.PathValue("version")
	getLatest := versionRequested == "latest"
	versionWithoutV := strings.Replace(versionRequested, "v", "", 1)
	if string(versionRequested[0]) != "v" {
		versionRequested = "v" + versionRequested
	}

	client := &http.Client{}

	var req *http.Request
	var err error

	if getLatest {
		req, err = http.NewRequest("GET", GIT_LATEST_URL, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		req, err = http.NewRequest("GET", fmt.Sprintf(GIT_TAG_URL, versionRequested), nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	req.Header.Add("Authorization", GIT_PERSONAL_TOKEN)
	req.Header.Add("Accept", "application/vnd.github+json")
	req.Header.Add("X-GitHub-Api-Version", "2022-11-28")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("fail to download file.")
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Println("wrong status from get release")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var result Release
	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Println("fail to get release info")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	folderPath := "./downloads/" + result.TagName
	if _, err := os.Stat(folderPath); errors.Is(err, os.ErrNotExist) {
		os.MkdirAll(folderPath, os.ModePerm)
	}

	fullName := ""
	fmtUrl := "http://" + SERVER_URL + "/download/cli/" + versionWithoutV + "/%s"

	var allAssets []AllAssets
	for _, a := range result.Assets {

		if strings.Contains(a.Name, soRequested) || soRequested == "" {

			name := reconstructName(strings.Split(a.Name, "_"), 2)
			allAssets = append(allAssets, AllAssets{Name: a.Name, TagName: result.TagName, URL: fmt.Sprintf(fmtUrl, name)})
			fullName = filepath.Join(folderPath, a.Name)
			_, err := os.Stat(fullName)

			if err == nil {
				if soRequested == "" {
					continue
				} else {
					break
				}
			}

			reqFile, err := http.NewRequest("GET", a.URL, nil)
			if err != nil {
				fmt.Println("fail create http request")
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			reqFile.Header.Add("Authorization", GIT_PERSONAL_TOKEN)
			reqFile.Header.Add("Accept", "application/octet-stream")
			reqFile.Header.Add("X-GitHub-Api-Version", "2022-11-28")

			// Criar o arquivo onde o conteúdo será salvo
			out, err := os.Create(filepath.Join(folderPath, a.Name))
			if err != nil {
				fmt.Println("fail to create file")
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer out.Close()

			respFile, err := client.Do(reqFile)
			if err != nil {
				fmt.Println("fail to download file..")
				fmt.Println(err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			_, err = io.Copy(out, respFile.Body)
			if err != nil {
				fmt.Println("fail to save downloaded file")
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

		}
	}

	if soRequested == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response, err := json.Marshal(allAssets)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(response)
		return
	}

	fn := filepath.Base(fullName)

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Access-Control-Expose-Headers", "Content-Disposition")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", fn))

	file, err := os.OpenFile(fullName, os.O_RDONLY, 0644)
	if err != nil {
		fmt.Println("Fail to get open file")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()
	io.Copy(w, file)

	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Println("Fail to get file stats")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
	// w.WriteHeader(http.StatusOK) //-- superfluous response.WriteHeader
}

func reconstructName(name []string, startIn int) string {
	result := ""
	for y, x := range name {
		if y >= startIn {
			if result == "" {
				result = x
				continue
			}
			result = fmt.Sprintf("%s_%s", result, x)

		}
	}
	return result
}

func getEnvOrThrow(env string) string {
	result := os.Getenv(env)
	if result == "" {
		log.Fatalf("cannot found environment value for %s\n", env)
	}
	return result
}
