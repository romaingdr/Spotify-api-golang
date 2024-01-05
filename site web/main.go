package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
)

var accessToken = ""

const (
	clientID     = "3e4c6edf57f841cb925046bee2808607"
	clientSecret = "836699a2060b460b88bafcd8a30930da"
)

type SoundInfo struct {
	Artists []struct {
		Name string `json:"name"`
	} `json:"artists"`
	Album struct {
		ReleaseDate string `json:"release_date"`
		Name        string `json:"name"`
		Images      []struct {
			URL string `json:"url"`
		} `json:"images"`
	} `json:"album"`
	ExternalURLs struct {
		Spotify string `json:"spotify"`
	} `json:"external_urls"`
}

type AlbumInfo struct {
	Name   string `json:"name"`
	Images []struct {
		URL string `json:"url"`
	} `json:"images"`
	ReleaseDate string `json:"release_date"`
	TotalTracks int    `json:"total_tracks"`
}

func getAccessToken() (string, error) {
	clientCreds := fmt.Sprintf("%s:%s", clientID, clientSecret)
	clientCredsB64 := base64.StdEncoding.EncodeToString([]byte(clientCreds))

	tokenURL := "https://accounts.spotify.com/api/token"

	tokenData := strings.NewReader("grant_type=client_credentials")

	tokenHeaders := map[string]string{
		"Authorization": "Basic " + clientCredsB64,
		"Content-Type":  "application/x-www-form-urlencoded",
	}

	req, err := http.NewRequest("POST", tokenURL, tokenData)
	if err != nil {
		return "", err
	}

	for key, value := range tokenHeaders {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Vérification de la réponse
	if resp.StatusCode == http.StatusOK {
		var tokenResponse map[string]interface{}
		err := json.NewDecoder(resp.Body).Decode(&tokenResponse)
		if err != nil {
			return "", err
		}

		accessToken, ok := tokenResponse["access_token"].(string)
		if !ok {
			return "", fmt.Errorf("token d'accès non trouvé")
		}
		return accessToken, nil
	}

	return "", fmt.Errorf("échec de l'obtention du token d'accès: %s", resp.Status)
}

func getSoundInfo(w http.ResponseWriter, r *http.Request) {
	soundID := "0EzNyXyU7gHzj2TN8qYThj"

	url := fmt.Sprintf("https://api.spotify.com/v1/tracks/%s", soundID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println("Error lors de la requête : ", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Erreur lors de l'envoi de la requête :", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var soundInfo SoundInfo
	err = json.NewDecoder(resp.Body).Decode(&soundInfo)
	if err != nil {
		log.Println("Erreur lors du décodage JSON :", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseGlob("./templates/*.gohtml")
	if err != nil {
		log.Println("Erreur lors du parsing du template :", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	fmt.Println("URL de l'image de l'album:", soundInfo.Album.Images[0].URL)

	err = tmpl.ExecuteTemplate(w, "sdm", soundInfo)
	if err != nil {
		log.Println("Erreur lors de l'exécution du template :", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func getJulAlbums(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseGlob("./templates/*.gohtml")
	artistID := "3IW7ScrzXmPvZhB27hmfgy" // Id de JUL

	url := fmt.Sprintf("https://api.spotify.com/v1/artists/%s/albums", artistID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println("Erreur avec la requête api:", err)
		return
	}

	// Token d'accès dans le header
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	var albums struct {
		Items []AlbumInfo `json:"items"`
	}

	// Décode la réponse en json
	err = json.NewDecoder(resp.Body).Decode(&albums)
	if err != nil {
		log.Println("Error decoding response:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	tmpl.ExecuteTemplate(w, "jul", albums.Items)
}

func main() {

	// Obtention d'un token (pour éviter l'expiration)
	token, err := getAccessToken()
	if err != nil {
		log.Fatal("Erreur lors de l'obtention du token d'accès:", err)
	}
	fmt.Println("[🪙] >> ", token)
	accessToken = token

	css := http.FileServer(http.Dir("./client/style"))
	http.Handle("/static/", http.StripPrefix("/static/", css))

	tmpl, _ := template.ParseGlob("./templates/*.gohtml")

	// Route d'accueil
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := tmpl.ExecuteTemplate(w, "main", nil)
		if err != nil {
			log.Println(err)
			http.Error(w, "Erreur : ", http.StatusInternalServerError)
		}
	})

	// Route des albums de jul
	http.HandleFunc("/album/jul", getJulAlbums)

	//Route de la musique de sdm
	http.HandleFunc("/track/sdm", getSoundInfo)

	// Démarrage du serveur
	log.Println("[✅] Serveur lancé !")
	fmt.Println("[🌐] http://localhost:8080")
	fmt.Println("[👽] http://localhost:8080/album/jul")
	fmt.Println("[🏎️] http://localhost:8080/track/sdm")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
