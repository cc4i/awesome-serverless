// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// [START eventarc_generic_handler]

// Sample generic is a Cloud Run service which logs and echos received requests.
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// GenericHandler receives and echos a HTTP request's headers and body.
func GenericHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Event received!")

	// Log all headers besides authorization header
	log.Println("HEADERS:")
	headerMap := make(map[string]string)
	for k, v := range r.Header {
		if k != "Authorization" {
			val := strings.Join(v, ",")
			headerMap[k] = val
			log.Println(fmt.Sprintf("%q: %q\n", k, val))
		}
	}

	// Log body
	log.Println("BODY:")
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("error parsing body: %v", err)
	}
	body := string(bodyBytes)
	log.Println(body)

	// Format and print full output
	type result struct {
		Headers map[string]string `json:"headers"`
		Body    string            `json:"body"`
	}
	res := &result{
		Headers: headerMap,
		Body:    body,
	}
	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Printf("error encoding response: %v", err)
		http.Error(w, "Could not marshal JSON output", 500)
		return
	}
	fmt.Fprintln(w)

	//Exec job
	InvokeRunJob()
}

// [END eventarc_generic_handler]
func GetMetadata(key string) (string, error) {
	baseUrl := "http://metadata.google.internal"
	url := baseUrl + key
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("Got error %s", err.Error())
	}
	req.Header.Add("Metadata-Flavor", "Google")
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Got error %s", err.Error())
	}
	defer resp.Body.Close()
	bt, _ := ioutil.ReadAll(resp.Body)

	return string(bt), nil
}

type TokenStruct struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

func InvokeRunJob() {

	// 1. Retrive access token-> http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/token
	atoken, _ := GetMetadata("/computeMetadata/v1/instance/service-accounts/default/token")
	log.Println(atoken)
	var data TokenStruct
	if err := json.Unmarshal([]byte(atoken), &data); err != nil {
		log.Println(err)
	}

	// 2. Get project id-> http://metadata.google.internal/computeMetadata/v1/project/project-id
	projectId, _ := GetMetadata("/computeMetadata/v1/project/project-id")
	log.Println(projectId)
	// 3. Get job name-> os.Getenv("JOB_NAME")
	jobName := os.Getenv("JOB_NAME")
	log.Println(jobName)

	// 4. Get region
	region, _ := GetMetadata("/computeMetadata/v1/instance/region")
	log.Println(region)
	// 5. Make up url with variables and then trigger the job
	log.Printf("Authorization: Bearer %s", data.AccessToken)
	jobUrl := fmt.Sprintf("https://%s-run.googleapis.com/apis/run.googleapis.com/v1/namespaces/%s/jobs/%s:run", strings.Split(region, "/")[3], projectId, jobName)
	log.Println(jobUrl)

	//	curl -H "Content-Type: application/json" \
	//	  -H "Authorization: Bearer ACCESS_TOKEN" \
	//	  -X POST \
	//	  -d '' \
	//	  https://REGION-run.googleapis.com/apis/run.googleapis.com/v1/namespaces/PROJECT-ID/jobs/JOB-NAME:run
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	req, err := http.NewRequest("POST", jobUrl, nil)
	if err != nil {
		log.Println(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", data.AccessToken))
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()
	bt, _ := ioutil.ReadAll(resp.Body)

	log.Println(string(bt))
}

// [START eventarc_generic_server]

func main() {
	http.HandleFunc("/", GenericHandler)
	// Determine port for HTTP service.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}
	// Start HTTP server.
	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

// [END eventarc_generic_server]
