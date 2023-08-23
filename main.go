package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)
type ClonesEntry struct
{
    Timestamp time.Time `json:"timestamp"`
    Count     int       `json:"count"` 
    Uniques   int       `json:"uniques"`
}
type JsonContents struct {
    Count   int `json:"count"`
    Uniques int `json:"uniques"`
    Clones  []ClonesEntry `json:"clones"`
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

    cmd := "gh api   -H \"Accept: application/vnd.github+json\"   -H \"X-GitHub-Api-Version: 2022-11-28\"   /repos/" + ghRepoOwner + "/traffic/clones"
    exCmd := exec.Command("bash", "-c", cmd);
    if os.Getenv("GH_TOKEN") == "" && ghTokenPath == ""{
	log.Fatal("GH_TOKEN env variable not set and path to token not specified!")
    }
    if _, err := os.Stat(ghTokenPath); err == nil {
	token, err := exec.Command("bash", "-c", "cat " + ghTokenPath).Output()
	if err != nil {
	    log.Fatal(err)
	}
	os.Setenv("GH_TOKEN", string(token))
	exCmd.Env = append(os.Environ())
    }
    out, err := exCmd.Output()
    if err != nil {
	log.Fatal("Error getting response from GitHub server. Ensure there are no syntax errors. ",
	"Command error: ",err)
    }

    jsonFile := JsonContents{}
    err = json.Unmarshal(out, &jsonFile)
    if err != nil {
	log.Fatal(err)
    }

    if csvFilePath == "" {
	filename := strings.Split(ghRepoOwner, "/")[0] + "_" + strings.Split(ghRepoOwner, "/")[1]
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
	    log.Fatal(err)
	}
	csvFilePath = userHomeDir + "/" + filename + "_traffic.csv"
	fmt.Println("Creating new file: " + csvFilePath)
    }
    file, err := os.OpenFile(csvFilePath, os.O_CREATE | os.O_RDWR, 0644)
    if err != nil{
	log.Fatal(err)
    }
    csvReader := csv.NewReader(file)
    var lastRecord  []string 
    for {
	record, err := csvReader.Read()
	if(err != nil){
	    break;
	}
	lastRecord = record;
    }
    csvWriter := csv.NewWriter(file);
    var lastRecordTimestamp time.Time;
    if len(lastRecord) > 0{
	lastRecordTimestamp, err = time.Parse("2006-01-02 15:04:05 +0000 MST", lastRecord[0])
    } else{
	fmt.Println("No records found. Createing header")
	file.WriteString("Date, Clones, Uniques\n")
    }
    newEntires := 0
    for _,value := range jsonFile.Clones{
	if value.Timestamp.After(lastRecordTimestamp){
	    var record []string
	    record = append(record, value.Timestamp.String())
	    record = append(record, fmt.Sprint(value.Count))
	    record = append(record, fmt.Sprint(value.Uniques))
	    err = csvWriter.Write(record)
	    if err != nil{
		log.Fatal(err)
	    }
	    newEntires += 1
	}
    }
    csvWriter.Flush()
    defer file.Close()
    if newEntires == 0 {
	fmt.Println("No new records to add.")
	return
    }
    fmt.Println("Succesfully wrote ", newEntires, "records to .csv file")
}


