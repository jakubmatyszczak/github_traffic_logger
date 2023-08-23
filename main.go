package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
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
    if len(os.Args) < 2 {
	fmt.Println("Too few arguments!")
	fmt.Println("Syntax github_traffic_logger <repo_owner/repo_name> (optional)<csv_filepath> (optional)<gh_token_path>")
	fmt.Println("Example: github_traffic_logger jakubmatyszczak/github_traffic_logger ~/log_path.csv ~/token")
	log.Fatal(os.ErrInvalid)
    }
    ghRepoOwner := os.Args[1]
    // csvFilePath := os.Args[2]
    // ghTokenFile := os.Args[3]
    
    cmd := "gh api   -H \"Accept: application/vnd.github+json\"   -H \"X-GitHub-Api-Version: 2022-11-28\"   /repos/" + ghRepoOwner + "/traffic/clones"
    exCmd := exec.Command("bash", "-c", cmd);
    exCmd.Env = os.Environ();
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

    file, err := os.OpenFile("log.csv", os.O_CREATE | os.O_RDWR, 0644)
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
