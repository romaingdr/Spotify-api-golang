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

// SoundInfo est une structure de données d'un morceau
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

// AlbumInfo est une structure de donnée d'un album
type AlbumInfo struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Images []struct {
		URL string `json:"url"`
	} `json:"images"`
	ReleaseDate string `json:"release_date"`
	TotalTracks int    `json:"total_tracks"`
}

// TrackInfo est une structure qui contient le nom et l'id d'un morceau d'un album
type TrackInfo struct {
	Name string `json:"name"`
	Id   string `json:"id"`
}

// AlbumTracks est une structure d'un album qui contient une liste de tracks
type AlbumTracks struct {
	Tracks struct {
		Items []TrackInfo `json:"items"`
	} `json:"tracks"`
}

// getAlbumDetails recupère les tracks d'un album grâce à son id et l'api puis execute la template "album"
func getAlbumDetails(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseGlob("./templates/*.gohtml")

	albumID := r.URL.Path[len("/album/"):] // Récupération de l'id album passé dans l'url

	url := fmt.Sprintf("https://api.spotify.com/v1/albums/%s", albumID) // Création de la requête api
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println("Error lors de la création requête : ", err)
		return
	}

	req.Header.Set("Authorization", "Bearer "+accessToken) // Ajout du token d'autorisation dans le header

	client := &http.Client{} // Envoi de la requête
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Erreur lors de l'envoi requête :", err)
		return
	}
	defer resp.Body.Close()

	var albumTracks AlbumTracks // Création d'une variable AlbumTracks pour stocker les informations de l'album
	err = json.NewDecoder(resp.Body).Decode(&albumTracks)
	if err != nil {
		log.Println("Erreur lors du décodage JSON :", err)
		return
	}

	trackNames := make([]string, len(albumTracks.Tracks.Items))
	for i, track := range albumTracks.Tracks.Items {
		trackNames[i] = track.Name
	}

	tmpl.ExecuteTemplate(w, "album", albumTracks.Tracks.Items) // Execution du template "album" avec les données de l'album
}

// getSoundInfo recupère les informations d'un morceau grâce à l'api puis execute la template "sdm"
func getSoundInfo(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseGlob("./templates/*.gohtml")

	soundID := "0EzNyXyU7gHzj2TN8qYThj" // ID du morceau "bolide allemand" de SDM

	url := fmt.Sprintf("https://api.spotify.com/v1/tracks/%s", soundID) // Création de la requête api
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println("Error lors de la requête : ", err)
		return
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Erreur lors de l'envoi de la requête :", err)
		return
	}
	defer resp.Body.Close()

	var soundInfo SoundInfo
	err = json.NewDecoder(resp.Body).Decode(&soundInfo)
	if err != nil {
		log.Println("Erreur lors du décodage JSON :", err)
		return
	}

	err = tmpl.ExecuteTemplate(w, "sdm", soundInfo)
	if err != nil {
		log.Println("Erreur lors de l'exécution du template :", err)
		return
	}
}

// getJulAlbums recupère les albums de Jul avec l'api puis execute la template "jul"
func getJulAlbums(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseGlob("./templates/*.gohtml")

	artistID := "3IW7ScrzXmPvZhB27hmfgy" // Id de l'artiste Jul

	url := fmt.Sprintf("https://api.spotify.com/v1/artists/%s/albums", artistID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println("Erreur avec la requête api:", err)
		return
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Erreur lors de l'envoi de la requête :", err)
		return
	}
	defer resp.Body.Close()

	var albums struct {
		Items []AlbumInfo `json:"items"`
	}

	err = json.NewDecoder(resp.Body).Decode(&albums)
	if err != nil {
		log.Println("Error decoding response:", err)
		return
	}

	tmpl.ExecuteTemplate(w, "jul", albums.Items)
}

// getAccessToken recupère un token d'accès api avec les identifiants : clientID et clientSecret
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

// main est la fonction principale du programme, elle lance le serveur sur le port 8080
func main() {

	// Obtention d'un token (pour éviter l'expiration)
	token, err := getAccessToken()
	if err != nil {
		log.Fatal("Erreur lors de l'obtention du token d'accès:", err)
	}
	fmt.Println("[🟡 ] >>", token)
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

	//Route d'album
	http.HandleFunc("/album/", getAlbumDetails)

	// Démarrage du serveur
	log.Println("[✅ ] >> Serveur lancé !")
	fmt.Println("[🌐] >> http://localhost:8080")
	fmt.Println("[👽] >> http://localhost:8080/album/jul")
	fmt.Println("[🏎️] >> http://localhost:8080/track/sdm")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
