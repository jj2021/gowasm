// +build js,wasm
package main

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"syscall/js"
)

//Response represents a git api call response
type Response struct {
	Encoding string `json:"encoding"`
	Name     string `json:"name"`
	Content  string `json:"content"`
}

var document js.Value

//Retreive latest covid data with call to github api
var confirmedContent = retrieveData("https://api.github.com/repos/CSSEGISandData/COVID-19/contents/csse_covid_19_data/csse_covid_19_time_series/time_series_covid19_confirmed_global.csv")
var deathContent = retrieveData("https://api.github.com/repos/CSSEGISandData/COVID-19/contents/csse_covid_19_data/csse_covid_19_time_series/time_series_covid19_deaths_global.csv")
var recoveredContent = retrieveData("https://api.github.com/repos/CSSEGISandData/COVID-19/contents/csse_covid_19_data/csse_covid_19_time_series/time_series_covid19_recovered_global.csv")

func main() {
	//make a channel to keep the main function from closing
	done := make(chan struct{}, 0)

	//get document for DOM manipulation
	document = js.Global().Get("document")

	//make update function available to js
	updateFunc := js.FuncOf(update)
	js.Global().Set("update", updateFunc)
	defer updateFunc.Release()

	<-done
}

func update(this js.Value, args []js.Value) interface{} {
	confirmed(confirmedContent)
	deaths(deathContent)
	recovered(recoveredContent)
	return js.ValueOf(true)
}

//call the github api to retrieve the csv file holding covid19 data
func retrieveData(url string) []byte {

	response, err := http.Get(url)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	//convert response into bytes
	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	//convert response bytes into a custom type called "Response"
	var gitResponse Response
	json.Unmarshal(responseData, &gitResponse)

	//decode the body (csv file content) from base64 into bytes slice
	contentBytes, err := base64.StdEncoding.DecodeString(gitResponse.Content)
	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	//return the content of the csv in a byte slice format
	return contentBytes

}

//retrieve data about confirmed cases
func confirmed(c []byte) {
	dataString := filterData(c)
	fmt.Print(dataString)

	para := document.Call("getElementById", "confdata")
	para.Set("innerText", dataString)
}

//retrieve data about deaths
func deaths(c []byte) {
	dataString := filterData(c)
	fmt.Print(dataString)

	para := document.Call("getElementById", "deathdata")
	para.Set("innerText", dataString)
}

//retrieve data about recoveries
func recovered(c []byte) {
	dataString := filterData(c)
	fmt.Print(dataString)

	para := document.Call("getElementById", "recdata")
	para.Set("innerText", dataString)
}

//filter data by coutry and date, return the requested result in a string
func filterData(content []byte) string {
	//read the csv file
	csvReader := csv.NewReader(bytes.NewReader(content))
	record, err := csvReader.ReadAll()

	if err != nil {
		fmt.Println(err.Error())
	}

	//place the information into a string
	country := "US"
	dateIndex := len(record[0]) - 1
	date := record[0][dateIndex]
	stat := string(record[226][dateIndex])
	return string(country + " on " + date + "\n" + stat + "\n")
}
