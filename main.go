package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)
type TrafficEntry struct
{
    Timestamp time.Time `json:"timestamp"`
    Count     int       `json:"count"` 
    Uniques   int       `json:"uniques"`
}
type JsonClones struct {
    Count   int `json:"count"`
    Uniques int `json:"uniques"`
    Clones  []TrafficEntry `json:"clones"`
}
type JsonViews struct {
    Count   int `json:"count"`
    Uniques int `json:"uniques"`
    Views []TrafficEntry `json:"views"`
}
func getToken(filePath string) (string, error) {
    if os.Getenv("GH_TOKEN") == "" && filePath == ""{
	log.Fatal("GH_TOKEN env variable not set and path to token not specified!")
    }
    var err error
    if _, err := os.Stat(filePath); err == nil {
	tokenBytes, err := os.ReadFile(filePath)
	if err != nil {
	    return "", err
	}
	token := string(tokenBytes)
	if token[len(token) - 1:] == "\n" {
	    token = token[0:len(token)-1]
	}
	return token, nil
    }
    return "", err
}
func callGhApi(repo string, trafficTarget string, token string) ([]byte, error){
    url := "https://api.github.com/repos/" + repo + "/traffic/" + trafficTarget
    req, err := http.NewRequest("GET", url, nil)
    if err != nil{
	return nil, err
    }
    req.Header.Set("Accept", "application/vnd.github+json")
    req.Header.Set("Authorization", "Bearer " + token)
    req.Header.Set("X-Github-Api-Version", "2022-11-28")
    response, err := http.DefaultClient.Do(req)
    if err != nil {
	return nil, err
    }
    defer response.Body.Close()
    httpOutput, err := io.ReadAll(response.Body)
    if err != nil {
	return nil, err
    }
    return httpOutput, nil
}

func createCsv(target string) (string,error) {
    owner :=  strings.Split(target, "/")[0]
    repo := strings.Split(target, "/")[1]
    userHomeDir, err := os.UserHomeDir()
    if err != nil {
	return "", err
    }
    csvPath := userHomeDir + "/" + owner + "_" + repo + "_traffic.csv"
    return csvPath, nil
}
func getLastRecordFromCsv(reader csv.Reader) []string {
    var lastRecord  []string 
    for {
	record, err := reader.Read()
	if(err != nil){
	    break;
	}
	lastRecord = record;
    }
    if len(lastRecord) == 0{
	return nil
    }
    return lastRecord
}
func main(){
    csvFilePath := ""
    ghTokenPath := ""
    ghRepoOwner := ""

    if len(os.Args) < 2{
	log.Fatal("Not enought arguments!")
    }
    ghRepoOwner = os.Args[1]
    for i := 2; i < len(os.Args)-1; i++ {
	var arg string
	arg = os.Args[i]
	if arg[0:2] == "-c"{
	    csvFilePath = os.Args[i+1]
	    i += 1
	}
	if arg[0:2] == "-t"{
	    ghTokenPath = os.Args[i+1]
	    i += 1
	}
    }
    fmt.Println("tokenpath: ", ghTokenPath)
    fmt.Println("csvpath: ", csvFilePath)
    fmt.Println("ghrepo: ", ghRepoOwner)

    token, err := getToken(ghTokenPath)
    if err != nil{
	log.Fatal("Could not get token, neither from ENV for from file: ", ghTokenPath)
    }

    httpOutput, err := callGhApi(ghRepoOwner, "clones", token)
    if err != nil{
	log.Fatal(err)
    }
    jsonClones := JsonClones{}
    err = json.Unmarshal(httpOutput, &jsonClones)
    if err != nil {
	log.Fatal(err)
    }

    httpOutput, err = callGhApi(ghRepoOwner, "views", token)
    jsonViews := JsonViews{}
    err = json.Unmarshal(httpOutput, &jsonViews)
    if err != nil {
	log.Fatal(err)
    }
    if csvFilePath == "" {
	csvFilePath,err = createCsv(ghRepoOwner)
	if err != nil{
	    log.Fatal(err)
	}
    }
    file, err := os.OpenFile(csvFilePath, os.O_CREATE | os.O_RDWR, 0644)
    if err != nil{
	log.Fatal(err)
    }
    defer file.Close()

    csvReader := csv.NewReader(file)
    lastRecord := getLastRecordFromCsv(*csvReader)
    var lastRecordTimestamp time.Time;
    var lastRecordNo int
    if lastRecord == nil{
	fmt.Println("No records found. Createing header")
	file.WriteString("No, Date, Clones, Unique Clones, Views, Unique Views\n")
    } else {
	lastRecordNo,_ = strconv.Atoi(lastRecord[0])
	lastRecordTimestamp, err = time.Parse("2006-01-02 15:04:05 +0000 MST", lastRecord[1])
    }
    var firstNewRecordTimestamp time.Time
    var lastNewRecordTimestamp time.Time
    if jsonClones.Clones[0].Timestamp.Before(jsonViews.Views[0].Timestamp) {
	firstNewRecordTimestamp = jsonClones.Clones[0].Timestamp 
    } else {
	firstNewRecordTimestamp = jsonViews.Views[0].Timestamp
    }
    if jsonClones.Clones[len(jsonClones.Clones)-1].Timestamp.After(jsonViews.Views[len(jsonViews.Views)-1].Timestamp) {
	lastNewRecordTimestamp = jsonClones.Clones[len(jsonClones.Clones)-1].Timestamp 
    } else {
	lastNewRecordTimestamp = jsonViews.Views[len(jsonViews.Views)-1].Timestamp
    }
    newRecords := int(lastNewRecordTimestamp.Sub(firstNewRecordTimestamp).Abs().Hours() / 24)
    csvWriter := csv.NewWriter(file);
    newEntires := 0
    omittedClones := 0
    omittedViews := 0
    for i:=0; i < newRecords; i++ {
	timestamp := firstNewRecordTimestamp.AddDate(0,0,i)
	if timestamp.Compare(lastRecordTimestamp) <= 0{
	    continue
	}
	var record []string
	record = append(record, fmt.Sprint(lastRecordNo + newEntires + 1))
        record = append(record, timestamp.String())

	clonesIter := i - omittedClones
	if jsonClones.Clones[clonesIter].Timestamp == timestamp{
	    record = append(record, fmt.Sprint(jsonClones.Clones[clonesIter].Count))
	    record = append(record, fmt.Sprint(jsonClones.Clones[clonesIter].Uniques))
	} else {
	    omittedClones += 1
	    record = append(record, "0")
	    record = append(record, "0")
	}
	viewsIter := i - omittedViews
	if jsonViews.Views[viewsIter].Timestamp == timestamp{
	    record = append(record, fmt.Sprint(jsonViews.Views[viewsIter].Count))
	    record = append(record, fmt.Sprint(jsonViews.Views[viewsIter].Uniques))
	} else {
	    omittedViews += 1
	    record = append(record, "0")
	    record = append(record, "0")
	}
        
	err = csvWriter.Write(record)
	if err != nil{
	    log.Fatal(err)
	}
	newEntires += 1
    }
    csvWriter.Flush()
    if newEntires == 0 {
	fmt.Println("No new records to add.")
	return
    }
    fmt.Println("Succesfully wrote ", newEntires, "records to .csv file")
}


