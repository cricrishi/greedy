
package main

import (
	"log"
	"net/http"
	"github.com/gorilla/mux"
	"encoding/json"
	"io/ioutil"
	"net/url"
	"bytes"
	"time"
	"os"

)

func main() {

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/search", fetchData).Methods("GET")
	log.Fatal(http.ListenAndServe(GetPort(), router))

}

func fetchData(w http.ResponseWriter, r *http.Request) {
	queryString :=  r.URL.Query().Get("q")
	
	data := make(map[string]string,0)

	ch := make(chan map[string]string)

  	go fetchGoogle(queryString,ch)
	go fetchDuckDuckGo(queryString,ch)
	go fetchTwitter(queryString,ch)
	
	for i:=0; i<3; i++ {
		 x := <-ch
		 for k,v :=range x {
		 	data[k] = v
		 }
	}

	jData, err := json.Marshal(data)
	if err != nil {
    	panic(err)
    	return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jData)
}

func fetchGoogle(queryString string,ch chan<-map[string]string) {
	url := "https://www.googleapis.com/customsearch/v1?key=AIzaSyCco2D4R07YhUpNMEmV3q2oGcNb18xP7aw&cx=017576662512468239146:omuauf_lfve&q=" + queryString
	timeout := time.Duration(900 * time.Millisecond)
	client := http.Client{
    	Timeout: timeout,
	}
	responseMap := make(map[string]string)
	resp, e := client.Get(url)
	if e!= nil {
		responseMap["google"] = "Error"
		ch <- responseMap
		return
	} 
	defer resp.Body.Close()
	 
	 body, _ := ioutil.ReadAll(resp.Body)
	 responseMap["google"] = string(body)
	 ch<- responseMap
}

func fetchDuckDuckGo(queryString string,ch chan<-map[string]string) {
	url := "http://api.duckduckgo.com/?format=json&q=" + queryString
	timeout := time.Duration(900 * time.Millisecond)
	client := http.Client{
    	Timeout: timeout,
	}
	responseMap := make(map[string]string)
		
	resp, e := client.Get(url)
	if e!= nil {
		responseMap["duduckgo"] = "Error"
		ch <- responseMap
		return
	}
	defer resp.Body.Close()
	 body, _ := ioutil.ReadAll(resp.Body)
	 responseMap["duckduckgo"] = string(body)
	 ch<- responseMap
}

func fetchTwitter(queryString string,ch chan<-map[string]string)  {
	bToken := getBearerToken()
	url := "https://api.twitter.com/1.1/search/tweets.json?q=" + queryString
	timeout := time.Duration(900 * time.Millisecond)
	client := &http.Client{
		Timeout: timeout,
	}
	responseMap := make(map[string]string)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+ bToken)
	resp, e := client.Do(req)
	if e!= nil {
		responseMap["twitter"] = "Error"
		ch <- responseMap
		return
	}
	defer resp.Body.Close()
	 body, _ := ioutil.ReadAll(resp.Body)
	 responseMap["twitter"] = string(body)
	 ch<- responseMap
}

func getBearerToken() string{
	type response struct {
		TokenType  string `json:"token_type"`
		AccessToken string `json:"access_token"`
	}
	var res response
	postUrl := "https://api.twitter.com/oauth2/token"
	data := url.Values{}
    data.Set("grant_type", "client_credentials")
   //var jsonStr = []byte(`{"grant_type":"client_credentials"}`)

	req, _ := http.NewRequest("POST", postUrl, bytes.NewBuffer([]byte(data.Encode())))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Basic TWI4MTdlVklpYmFUeTdJODVXMjB4eFd4SDp5S3ptQ1k5dE4zVVA4Vlk5aTVyelFhQVpLNDJtZXVKeXpJM1U1ZHc0Y29OWGt0TlZMdw==")
	
	client := &http.Client{}
    resp, err := client.Do(req)
     if err != nil {
        panic(err)
    }
    defer resp.Body.Close()
    body, _ := ioutil.ReadAll(resp.Body)
    json.Unmarshal(body, &res)
    return res.AccessToken
}

func GetPort() string {
        var port = os.Getenv("PORT")
        if port == "" {
                port = "4747"
        }
        return ":" + port
}