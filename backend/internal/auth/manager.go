package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"os"

	"backend/internal/storage"

	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

type contextKey string

const (
	usernameKey contextKey = "username"

	GITHUB_EXCHANGE_TOKEN_URL = "https://github.com/login/oauth/access_token"
	GITHUB_USER_INFO_URL      = "https://api.github.com/user"
)

var store *sessions.CookieStore

func InitStore(secret string) {
	store = sessions.NewCookieStore(
		[]byte(secret),
		nil,
	)

	store.Options = &sessions.Options{
		HttpOnly: true,
		Secure:   os.Getenv("ENV") == "production",
		SameSite: getSameSiteValue(),
		MaxAge:   60 * 60 * 24 * 7, // 7 days
		Path:     "/",
	}
}

func GetUsernameFromContext(ctx context.Context) string {
	return ctx.Value(usernameKey).(string)
}

func ProtectedRoute(w http.ResponseWriter, r *http.Request) {
	username := GetUsernameFromContext(r.Context())
	json.NewEncoder(w).Encode(map[string]string{
		"username": username,
	})
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := store.Get(r, "studyhub-session")
		if err != nil {
			http.Error(w, "not authorized", http.StatusUnauthorized)
			return
		}

		username := session.Values["username"]
		if username == nil {
			http.Error(w, "not authorized", http.StatusUnauthorized)
			return
		}

		usernameStr, ok := username.(string)
		if !ok {
			http.Error(w, "not authorized", http.StatusUnauthorized)
			return
		}

		log.Printf("Authenticated user: %s", usernameStr)
		ctx := context.WithValue(r.Context(), usernameKey, usernameStr)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func SignUpHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "body must contain username and password strings", http.StatusBadRequest)
		return
	}

	username := html.EscapeString(req.Username)
	password := req.Password

	if username == "" || password == "" {
		http.Error(w, "username and password must be non-empty", http.StatusBadRequest)
		return
	}

	dbUsername, _, err := storage.GetUser(username)

	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if dbUsername == username {
		http.Error(w, "this username is already in use", http.StatusConflict)
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("error hashing: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if err := storage.CreateUser(username, string(passwordHash)); err != nil {
		log.Printf("error storing user to database: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	session, err := store.Get(r, "studyhub-session")
	if err != nil {
		log.Printf("error getting session: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
	session.Values["username"] = username

	if err := session.Save(r, w); err != nil {
		log.Printf("Session save error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	cookie := http.Cookie{
		Name:     "username",
		Value:    username,
		MaxAge:   60 * 60 * 24 * 7,
		Secure:   os.Getenv("ENV") == "production",
		Path:     "/",
		SameSite: getSameSiteValue(),
	}
	http.SetCookie(w, &cookie)

	json.NewEncoder(w).Encode(map[string]string{
		"username": username,
	})
}

func SignInHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "body must contain non-empty username and password strings", http.StatusBadRequest)
		return
	}

	username := html.EscapeString(req.Username)
	password := req.Password

	if username == "" || password == "" {
		http.Error(w, "body must contain non-empty username and password strings", http.StatusBadRequest)
		return
	}

	dbUsername, dbPasswordHash, _ := storage.GetUser(username)
	if dbUsername == "" || dbPasswordHash == "" {
		http.Error(w, "access denied", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(dbPasswordHash), []byte(password)); err != nil {
		http.Error(w, "access denied", http.StatusUnauthorized)
		return
	}

	session, err := store.Get(r, "studyhub-session")
	if err != nil {
		log.Printf("rror getting session: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
	session.Values["username"] = username

	if err := session.Save(r, w); err != nil {
		log.Printf("Session save error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	cookie := http.Cookie{
		Name:     "username",
		Value:    username,
		MaxAge:   60 * 60 * 24 * 7,
		Secure:   os.Getenv("ENV") == "production",
		Path:     "/",
		SameSite: getSameSiteValue(),
	}
	http.SetCookie(w, &cookie)

	json.NewEncoder(w).Encode(map[string]string{
		"username": username,
	})
}

func SignOutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "studyhub-session")

	// delete the session by setting max age to -1
	session.Options.MaxAge = -1
	session.Save(r, w)

	cookie := http.Cookie{Name: "username", Value: "", MaxAge: 60 * 60 * 24 * 7, Path: "/"}
	http.SetCookie(w, &cookie)

	http.Redirect(w, r, getSignOutRedirectUrl(), http.StatusFound)
}

func GithubLoginHandler(w http.ResponseWriter, r *http.Request) {
	// first step is to get the code from request
	var req struct {
		Code string `json:"code"`
	}

	json.NewDecoder(r.Body).Decode(&req)
	code := req.Code

	if code == "" {
		http.Error(w, "github auth code is required", http.StatusBadRequest)
		return
	}

	// next step is to exchange code for token
	client := &http.Client{}

	exchangeReq, _ := http.NewRequest("POST", GITHUB_EXCHANGE_TOKEN_URL, nil)
	exchangeReq.Header.Set("Accept", "application/json")

	q := exchangeReq.URL.Query()
	q.Add("client_id", os.Getenv("GITHUB_CLIENT_ID"))
	q.Add("client_secret", os.Getenv("GITHUB_CLIENT_SECRET"))
	q.Add("code", code)
	exchangeReq.URL.RawQuery = q.Encode()

	resp, err := client.Do(exchangeReq)
	if err != nil {
		log.Printf("error getting github callback: %v", err)
		http.Error(w, "failed to exchange code for token", http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("github API call failed with status code: %d %s", resp.StatusCode, resp.Status)
		http.Error(w, fmt.Sprintf("github auth failed: %s", resp.Status), http.StatusBadRequest)
		return
	}

	// next step is to get the acutal access token that we just exchanged
	var responseStructure struct {
		AccessToken           string `json:"access_token"`
		ExpiresIn             int    `json:"expires_in"`
		RefreshToken          string `json:"refresh_token"`
		RefreshTokenExpiresIn int    `json:"refresh_token_expires_in"`
		TokenType             string `json:"token_type"`
	}

	json.NewDecoder(resp.Body).Decode(&responseStructure)

	accessTok := responseStructure.AccessToken

	if accessTok == "" {
		log.Printf("failed to get access token from GitHub")
		http.Error(w, "internal server error", http.StatusBadRequest)
		return
	}

	// next step is to use access token to get user info from github
	userInfoReq, err := http.NewRequest("GET", GITHUB_USER_INFO_URL, nil)
	if err != nil {
		log.Printf("error in second api call: %v", err)
		return
	}

	userInfoReq.Header.Add("Accept", "application/vnd.github+json")
	userInfoReq.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessTok))
	userInfoReq.Header.Add("X-Github-Api-Version", "2022-11-28")

	response, err := client.Do(userInfoReq)
	if err != nil {
		log.Printf("error at resp2 after adding headers: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		log.Printf("GitHub user info request failed with status code: %d", response.StatusCode)
		http.Error(w, "internal server error", http.StatusBadRequest)
		return
	}

	var userInfoResponse struct {
		Login string `json:"login"`
	}

	json.NewDecoder(response.Body).Decode(&userInfoResponse)

	username := userInfoResponse.Login

	if username == "" {
		log.Printf("GitHub username is empty")
		http.Error(w, "failed to get username from GitHub", http.StatusBadRequest)
		return
	}

	if err := storage.CreateOrUpdateGitHubUser(username, accessTok); err != nil {
		log.Printf("error storing github user: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// next step is to log the user in by starting a session
	session, err := store.Get(r, "studyhub-session")
	if err != nil {
		log.Printf("error getting session: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
	session.Values["username"] = username

	if err := session.Save(r, w); err != nil {
		log.Printf("session save error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	cookie := http.Cookie{
		Name:     "username",
		Value:    username,
		MaxAge:   60 * 60 * 24 * 7,
		Secure:   os.Getenv("ENV") == "production",
		Path:     "/",
		SameSite: getSameSiteValue(),
	}
	http.SetCookie(w, &cookie)

	json.NewEncoder(w).Encode(map[string]string{
		"username": username,
	})
}

func getSameSiteValue() http.SameSite {
	var sameSite http.SameSite

	if os.Getenv(("ENV")) == "production" {
		sameSite = http.SameSiteStrictMode
	} else {
		sameSite = http.SameSiteLaxMode
	}

	return sameSite
}

func getSignOutRedirectUrl() string {
	redirectUrl := "/"

	if os.Getenv("ENV") == "production" {
		redirectUrl = os.Getenv("PROD_APP_URL")
	} else {
		redirectUrl = os.Getenv("LOCAL_APP_URL")
	}

	return redirectUrl
}
