package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/DenisAltruist/distsys/db"
	"github.com/DenisAltruist/distsys/utils"
	jwt "github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

func calcPassHash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 16)
	return string(bytes), err
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
	}, nil
}

func validateToken(token string, duration time.Duration) bool {
	decodedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return os.Getenv("JWT_HS256_SECRET"), nil
	})
	if claims, ok := decodedToken.Claims.(jwt.MapClaims); ok && decodedToken.Valid {
		curTime := time.Now().Unix()
		expirationTime := claims["exp"].(int64)
		if expirationTime >= curTime {
			return true
		}
	}
	return false
}

func signUp(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	passwordHash, err := calcPassHash(password)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Can't hash password, got an error: %s", err.Error())
		return
	}
	client, ok := db.GetDbClient(w)
	if !ok {
		return
	}
	err = db.AddNewUser(client, &db.ShopUser{PasswordHash: passwordHash, Email: email}, time.Duration*5)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Can't sign up new user, got an error: %s", err.Error())
		return
	}
	utils.SendBodyResponse(w, "Successfully signed up!", http.StatusOK)
}

func signIn(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	passwordHash, err := calcPassHash(password)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Can't hash password, got an error: %s", err.Error())
		return
	}
	client, ok := db.GetDbClient(w)
	if !ok {
		return
	}
	filter := bson.D{bson.E{Key: "email", Value: email}, bson.E{Key: "password_hash", Value: passwordHash}}
	isFound, err := db.FindUser(client, &filter, 5*time.Minute)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Can't find user with pair (email, password), got an error: %s", err.Error())
		return
	}
	if !isFound {
		utils.SendError(w, http.StatusNotFound, "Can't find user with pair (email, password)")
		return
	}
	tokens, err := issueTokens(email)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Can't issue tokens pair, got an error: %s", err.Error())
		return
	}
	encodedTokens, err := json.Marshal(&tokens)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Can't encode JSON tokens pair, got an error: %s", err.Error())
		return
	}
	utils.SendBodyResponse(w, string(encodedTokens), http.StatusOK)
}

func validateEncodedToken(w http.ResponseWriter, token string, durationName string) bool {
	durMinutes, err := strconv.Atoi(os.Getenv(durationName))
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Can't parse %s tokens durations from config", durationName)
		return false
	}
	isTokenValid := validateToken(token, time.Duration(durMinutes)*time.Minute)
	if !isTokenValid {
		utils.SendError(w, http.StatusBadRequest, "Refresh token is expired or not correct")
		return false
	}
	return true
}

func refresh(w http.ResponseWriter, r *http.Request) {
	ok := validateEncodedToken(w, r.FormValue("token"), "REFRESH_TOKENS_DURATION_MINUTES")
	if !ok {
		return
	}
	tokens, err := issueTokens(email)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Can't issue tokens pair, got an error: %s", err.Error())
		return
	}
	encodedTokens, err := json.Marshal(&tokens)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Can't encode JSON tokens pair, got an error: %s", err.Error())
		return
	}
	utils.SendBodyResponse(w, string(encodedTokens), http.StatusOK)
}

func validate(w http.ResponseWriter, r *http.Request) {
	ok := validateEncodedToken(w, r.FormValue("token", "ACCESS_TOKENS_DURATION_MINUTES"))
	if !ok {
		utils.SendBodyResponse(w, "Not authorized", http.StatusOK)
		return
	}
	utils.SendBodyResponse(w, "Authorized", http.StatusOK)
}

func main() {
	http.HandleFunc("/signup", signUp)
	http.HandleFunc("/signin", signIn)
	http.HandleFunc("/refresh", refresh)
	http.HandleFunc("/validate", validate)
}
