package main

//based on:
//curl -i -X POST -H "Authorization: Bearer SIGNED_JWT" -H "Accept: application/vnd.github.v3+json" https://api.github.com/app/installations/18109498/access_tokens

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	b64 "encoding/base64"

	"github.com/golang-jwt/jwt"
)

// not ideal but it works
var (
	client  = &http.Client{}
	pemKey  string // Created via the GH console & base64 encoded `cat key.pem | base64`
	appID   int    // hold the numeric id for the installed Github app
	orgName string // the name of the user handle or github org name. e.g: stephenh1991
)

// holds nested json map of account details
type Account struct {
	Login string `json:"login"`
}

// holds a list of installation objects from the installed app locations (repos, orgs)
type InstallResponseJSON struct {
	Id      int     `json:"id"`
	Account Account `json:"account"`
}

// hold the json response from the token generation endpoint
type tokenResponseJSON struct {
	Token string `json:"token"`
}

// wrapper to request from an api and add bearer jwt tokens
func requester(client http.Client, method string, url string, token string) ([]byte, error) {
	req, err := http.NewRequest(method, url, nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Add("Accept", "Accept: application/vnd.github.v3+json")

	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode > 400 {
		return nil, fmt.Errorf("error incorrect status: %d, status text: %s", resp.StatusCode, resp.Status)
	}

	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return responseData, nil

}

// signs jwt tokens with pem files provided by github
func jwtSigner(signBytes []byte) (string, error) {
	signKey, err := jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		log.Fatalf("signKey invalid, error: %v", err)
	}

	// create a signer for rsa 256
	t := jwt.New(jwt.GetSigningMethod("RS256"))
	iss := fmt.Sprint(appID) // we convert this so flags can have stronger typing to guard against error.

	t.Claims = &jwt.StandardClaims{
		IssuedAt:  time.Now().Add(time.Minute * -1).Unix(),
		ExpiresAt: time.Now().Add(time.Minute * 9).Unix(),
		Issuer:    iss,
	}

	signedString, err := t.SignedString(signKey)

	if err != nil {
		return "", err
	}

	return signedString, nil
}

func findInstallation(installations []InstallResponseJSON, account string) (int, error) {
	for _, installation := range installations {
		if installation.Account.Login == account {

			return installation.Id, nil
		}
	}

	return 0, fmt.Errorf("installation for this bot was not found in: %s", account)
}

// generates a pure access token from github for a specific app installation which is short lived for 60 minutes
func tokenGenerator(account string, baseUrl string, jwt string) (string, error) {
	installationUrl := fmt.Sprintf("%s/app/installations", baseUrl)

	installationResponseData, err := requester(*client, http.MethodGet, installationUrl, jwt)

	if err != nil {
		return "", err
	}

	installations := []InstallResponseJSON{}
	err = json.Unmarshal(installationResponseData, &installations)
	if err != nil {
		return "", err
	}

	installId, err := findInstallation(installations, account)

	if err != nil {
		return "", err
	}

	tokenResponeData, err := requester(*client, http.MethodPost, fmt.Sprintf("%s/%d/access_tokens", installationUrl, installId), jwt)
	if err != nil {
		return "", err
	}

	token := tokenResponseJSON{}
	err = json.Unmarshal(tokenResponeData, &token)
	if err != nil {
		return "", err
	}

	return token.Token, nil
}

func keyDecoder(key string) ([]byte, error) {
	decodedKey, err := b64.StdEncoding.DecodeString(key)
	if err != nil {
		return nil, err
	}

	return decodedKey, nil

}

func main() {
	flag.StringVar(&pemKey, "pem-key", "", "Required: base64 encoded string form key.pem generated from Github App page")
	flag.IntVar(&appID, "app-id", 0, "Required: int value of installed Github App ID from the Github console")
	flag.StringVar(&orgName, "org-name", "", "Required: name of the users github handle or organisation name where the Github app is installed")

	flag.Parse()

	if pemKey == "" || appID == 0 || orgName == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	url := "https://api.github.com"

	decodedFile, err := keyDecoder(pemKey)

	if err != nil {
		log.Fatalf("Decoded file invalid, error: %v", err)
	}

	jwt, err := jwtSigner(decodedFile)

	if err != nil {
		log.Fatal(err)
	}

	token, err := tokenGenerator(orgName, url, jwt)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(token)

}
