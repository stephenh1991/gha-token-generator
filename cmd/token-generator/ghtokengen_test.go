package main

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestKeyFileDecoder(t *testing.T) {
	testCases := []struct {
		Name          string
		FilePath      string
		ExpectedError bool
	}{
		{"test successful b64 read", "testdata/key_64.pem", false},
		{"test failed base64 data", "testdata/key.pem", true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testData, err := os.ReadFile(testCase.FilePath)
			assert.Equal(t, err, nil)

			_, err = keyDecoder(string(testData))

			if testCase.ExpectedError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestTokenGenerator(t *testing.T) {
	defer gock.Off() // Flush pending mocks after test execution

	testCases := []struct {
		Name          string
		StatusCode    int
		ExpectedError bool
	}{
		{
			"test successful message processing", 200, false,
		},
		{
			"test failed message processing", 403, true,
		},
	}

	for _, testCase := range testCases {

		t.Run(testCase.Name, func(t *testing.T) {
			gock.New("https://api.github.com").
				Get("/app/installations").
				Reply(testCase.StatusCode).
				JSON(getInstallResponse(t))

			gock.New("https://api.github.com").
				Post("/app/installations/123456/access_tokens").
				Reply(testCase.StatusCode).
				JSON(getTokenResponse(t))

			_, err := tokenGenerator("org-name", "https://api.github.com", "jwttest")

			if testCase.ExpectedError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func getInstallResponse(t *testing.T) []map[string]interface{} {
	jsonBody := `[
		{
			"id": 123456,
			"account": {
				"login": "org-name"
			},
			"test": "ok"
		}
	]`
	var jsonData []map[string]interface{}
	err := json.Unmarshal([]byte(jsonBody), &jsonData)
	assert.Equal(t, err, nil)
	return jsonData
}

func getTokenResponse(t *testing.T) map[string]interface{} {
	// example key below isn't real
	jsonBody := `{
		"token": "ghs_j7YFxj88HsHeP25ynYDryI1XCb0soE0tNMB0"
	}`
	var jsonData map[string]interface{}
	err := json.Unmarshal([]byte(jsonBody), &jsonData)
	assert.Equal(t, err, nil)
	return jsonData
}
