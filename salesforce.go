/**
 *	Salesforce
 *	Copyright (C) 2025  hannjosh
 *
 *	This program is free software: you can redistribute it and/or modify
 *	it under the terms of the GNU General Public License as published by
 *	the Free Software Foundation, either version 3 of the License, or
 *	(at your option) any later version.
 *
 *	This program is distributed in the hope that it will be useful,
 *	but WITHOUT ANY WARRANTY; without even the implied warranty of
 *	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *	GNU General Public License for more details.
 *
 *	You should have received a copy of the GNU General Public License
 *	along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */
package salesforce

// Import standard packages.
import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

/*
 *	Version of the Salesforce REST API to use.
 *	@since	1.0.0
 */
const ApiVersion string = "v61.0"

/*
 *	My Domain
 *	The subdomain of the Salesforce org.
 *	@since	1.0.0
 */
var MyDomain string

/*
 *	OAuth 2.0 access token used to authorise the client.
 *	The Salesforce REST API supports the Bearer authentication type.
 *	@since	1.0.0
 */
var OAuth2AccessToken string

/*
 *	GetAuthorizationToken
 *	Obtains an OAuth 2.0 access token to authorise calls to the Salesforce REST API.
 *	@since	1.0.0
 */
func GetOAuth2AccessToken(client_id string, client_secret string) (string, error) {

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", client_id)
	data.Set("client_secret", client_secret)

	request, _ := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("https://%s.my.salesforce.com/services/oauth2/token", MyDomain),
		strings.NewReader(data.Encode()),
	)

	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	response, _ := (&http.Client{}).Do(request)

	var responseBody struct {
		// OK
		AccessToken string `json:"access_token"`
		InstanceUrl string `json:"instance_url"`
		Id          string `json:"id"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
		IssuedAt    string `json:"issued_at"`
		Signature   string `json:"signature"`

		// Error
		Error string `json:"error"`
	}

	json.NewDecoder(response.Body).Decode(&responseBody)

	response.Body.Close()

	if responseBody.TokenType == "" {
		return "", errors.New(responseBody.Error)
	}

	return responseBody.TokenType + " " + responseBody.AccessToken, nil

}

/*
 *	Query
 *	@since	1.0.0
 */
func Query(soql string) []byte {

	request, _ := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("https://%s.my.salesforce.com/services/data/%s/query/?q=%s", MyDomain, ApiVersion, url.QueryEscape(soql)),
		nil,
	)

	request.Header.Add("Accept", "application/json")
	request.Header.Add("Content-Type", "application/json; charset=UTF-8")
	request.Header.Add("Authorization", OAuth2AccessToken)

	response, _ := (&http.Client{}).Do(request)

	body, _ := io.ReadAll(response.Body)

	response.Body.Close()

	return body

}

/*
 *	Create
 *	@since	1.0.1
 */
func Create(object string, data map[string]interface{}) (string, error) {

	jsonData, _ := json.Marshal(data)

	request, _ := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("https://%s.my.salesforce.com/services/data/%s/sobjects/%s/", MyDomain, ApiVersion, object),
		bytes.NewBuffer(jsonData),
	)

	request.Header.Add("Authorization", OAuth2AccessToken)
	request.Header.Add("Content-Type", "application/json; charset=UTF-8")

	response, _ := (&http.Client{}).Do(request)

	body, _ := io.ReadAll(response.Body)

	var query struct {
		// 200 OK
		Id      string `json:"id"`
		Success bool
	}

	json.NewDecoder(bytes.NewReader(body)).Decode(&query)

	response.Body.Close()

	return query.Id, nil

}
