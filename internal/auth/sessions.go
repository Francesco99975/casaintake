package auth

import (
	"encoding/base64"
	"log"
	"net/http"

	"github.com/Francesco99975/casaintake/cmd/boot"
	"github.com/Francesco99975/casaintake/internal/enums"
	"github.com/gorilla/sessions"
)

var SessionStore *sessions.CookieStore

func InitSessionStore() {
	log.Print("Initializing SessionStore...")
	authKey, err := base64.StdEncoding.DecodeString(boot.Environment.SessionAuthKey)
	if err != nil {
		log.Fatal("Invalid SESSION_AUTH_KEY:", err)
	}

	encKey, err := base64.StdEncoding.DecodeString(boot.Environment.SessionEncryptionKey)
	if err != nil {
		log.Fatal("Invalid SESSION_ENCRYPTION_KEY:", err)
	}

	// Validate AES key length
	switch len(encKey) {
	case 16, 24, 32:
		// ok
	default:
		log.Fatalf("SESSION_ENCRYPTION_KEY decoded length must be 16, 24, or 32 bytes, got %d", len(encKey))
	}

	SessionStore = sessions.NewCookieStore(authKey, encKey)
}
func getSessionOptions(remember bool) *sessions.Options {
	domain := ""
	sameSite := http.SameSiteLaxMode
	maxAge := 86400 * 7 // One Week
	if remember {
		maxAge = maxAge * 52 // One Year
	}

	if boot.Environment.GoEnv == enums.Environments.DEVELOPMENT {

		if remember {
			maxAge = 0 //Session Only (Closing browser deletes session)
		} else {
			maxAge = 86400 / 24 / 60 * 5 // 5 minutes
		}

	}

	return &sessions.Options{
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   boot.Environment.GoEnv != enums.Environments.DEVELOPMENT,
		Domain:   domain,
		SameSite: sameSite,
	}

}

func SetSessionUser(w http.ResponseWriter, r *http.Request) error {
	session, err := SessionStore.Get(r, "session")
	if err != nil {
		return err
	}
	session.Values["authenticated"] = true
	session.Options = getSessionOptions(true)
	return session.Save(r, w)
}

func GetSessionUser(r *http.Request) bool {
	session, _ := SessionStore.Get(r, "session")
	authenticated, ok_authenticated := session.Values["authenticated"].(bool)
	if !ok_authenticated {
		log.Printf("No user_id in session: %v", session.Values)
	}
	return authenticated
}

func ClearSession(w http.ResponseWriter, r *http.Request) error {
	session, err := SessionStore.Get(r, "session")
	if err != nil {
		return err
	}
	session.Options.MaxAge = -1 // Delete cookie
	session.Options.Path = "/"
	return session.Save(r, w)
}
