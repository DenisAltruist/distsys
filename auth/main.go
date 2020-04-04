package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/DenisAltruist/distsys/db"
	"github.com/DenisAltruist/distsys/utils"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

func calcPassHash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 16)
	return string(bytes), err
}

func comparePass(password string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		log.Printf("Got an compare pass error: %s", err.Error())
		return false
	}
	return err == nil
}

func issueTokens(email string) (*db.TokensPair, error) {
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": email,
		"exp":   time.Now().Add(time.Minute * 5).Unix(),
	})
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": email,
		"exp":   time.Now().Add(time.Minute * 10).Unix(),
	})
	at, err := accessToken.SignedString([]byte(os.Getenv("JWT_HS256_SECRET")))
	if err != nil {
		return nil, err
	}
	rt, err := refreshToken.SignedString([]byte(os.Getenv("JWT_HS256_SECRET")))
	if err != nil {
		return nil, err
	}
	return &db.TokensPair{
		AccessToken:  at,
		RefreshToken: rt,
		Email:        email,
	}, nil
}

func validateToken(token string, duration time.Duration) (*jwt.MapClaims, error) {
	decodedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return []byte(os.Getenv("JWT_HS256_SECRET")), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := decodedToken.Claims.(jwt.MapClaims); ok && decodedToken.Valid {
		curTime := time.Now().Unix()
		expirationTime := claims["exp"].(float64)
		if int64(expirationTime) >= curTime {
			return &claims, nil
		}
	}
	return nil, errors.New("Can't convert jwt claims to map")
}

func getShopUserFromReq(w http.ResponseWriter, r *http.Request) (*db.ShopUser, bool) {
	contents, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Can't parse request body, got error: %s", err.Error())
		return nil, false
	}
	var user db.ShopUser
	err = json.Unmarshal(contents, &user)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Can't unrmashal contents %s, expected valid JSON", string(contents))
		return nil, false
	}
	return &user, true
}

func signUp(w http.ResponseWriter, r *http.Request) {
	newUser, ok := getShopUserFromReq(w, r)
	if !ok {
		return
	}
	passwordHash, err := calcPassHash(newUser.Password)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Can't hash password, got an error: %s", err.Error())
		return
	}
	client, ok := db.GetDbClient(w)
	if !ok {
		return
	}
	err = db.AddNewUser(client, &db.ShopUser{PasswordHash: passwordHash, Email: newUser.Email}, time.Second*5)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Can't sign up new user, got an error: %s", err.Error())
		return
	}
	utils.SendBodyResponse(w, "Successfully signed up!", http.StatusOK)
}

func signIn(w http.ResponseWriter, r *http.Request) {
	user, ok := getShopUserFromReq(w, r)
	if !ok {
		return
	}
	client, ok := db.GetDbClient(w)
	if !ok {
		return
	}
	filter := bson.D{bson.E{Key: "email", Value: user.Email}}
	foundUser, err := db.FindUser(client, &filter, 5*time.Second)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Can't find user with pair (email, password), got an error: %s", err.Error())
		return
	}
	signedIn := comparePass(user.Password, foundUser.PasswordHash)
	if !signedIn {
		utils.SendError(w, http.StatusNotFound, "Can't find user with pair (email, password)")
		return
	}
	tokens, err := issueTokens(user.Email)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Can't issue tokens pair, got an error: %s", err.Error())
		return
	}
	encodedTokens, err := json.Marshal(&tokens)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Can't encode JSON tokens pair, got an error: %s", err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s\n", string(encodedTokens))
}

func validateEncodedToken(w http.ResponseWriter, token string, durationName string) *jwt.MapClaims {
	durMinutes, err := strconv.Atoi(os.Getenv(durationName))
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Can't parse %s tokens durations from config: %s", durationName, err.Error())
		return nil
	}
	claims, err := validateToken(token, time.Duration(durMinutes)*time.Minute)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Token is expired or not correct: %s", err.Error())
		return nil
	}
	return claims
}

func refresh(w http.ResponseWriter, r *http.Request) {
	claims := validateEncodedToken(w, r.FormValue("token"), "REFRESH_TOKENS_DURATION_MINUTES")
	if claims == nil {
		return
	}
	tokens, err := issueTokens((*claims)["email"].(string))
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Can't issue tokens pair, got an error: %s", err.Error())
		return
	}
	encodedTokens, err := json.Marshal(&tokens)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Can't encode JSON tokens pair, got an error: %s", err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s\n", string(encodedTokens))
}

func validate(w http.ResponseWriter, r *http.Request) {
	token := r.FormValue("token")
	log.Printf("Got token: %s\n", token)
	claims := validateEncodedToken(w, token, "ACCESS_TOKENS_DURATION_MINUTES")
	if claims == nil {
		utils.SendBodyResponse(w, "Not authorized", http.StatusOK)
		return
	}
	utils.SendBodyResponse(w, "Authorized", http.StatusOK)
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/signup", signUp).Methods("POST")
	router.HandleFunc("/signin", signIn).Methods("PUT")
	router.HandleFunc("/refresh", refresh).Methods("PUT")
	router.HandleFunc("/validate", validate).Methods("GET")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("INTERNAL_LISTEN_PORT")), router))
}
