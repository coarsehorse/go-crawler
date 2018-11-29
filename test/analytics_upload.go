package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
)

/*
create table analitycs
(
	id int not null autoincrement,
	status int null,
	time_ms int null,
	url varchar(100) null,
	constraint analitycs_pk
		primary key (id)
);
*/

const (
	DB_TABLE_NAME = "analytics"
)

type CrawlingStats struct {
	Status int
	TimeMs int
	Url    string
}

type DBCredentials struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	Host_address string `json:"host_address"`
	Port         int    `json:"port"`
	Db_name      string `json:"db_name"`
}

func main() {
	start := time.Now()
	// Open json file with credentials
	jsonFile, err := os.Open("db_credentials.json")
	if err != nil {
		panic(err.Error())
	}

	byteValue, _ := ioutil.ReadAll(jsonFile)

	// Unmarshal data
	var cred DBCredentials
	err = json.Unmarshal(byteValue, &cred)
	if err != nil {
		panic(err.Error())
	}

	// Close opened credentials file
	err = jsonFile.Close()
	if err != nil {
		panic(err.Error())
	}

	// Open connection
	conn, err := sql.Open("mysql",
		cred.Username+":"+cred.Password+"@tcp("+cred.Host_address+":"+
			strconv.Itoa(cred.Port)+")/"+cred.Db_name+"?charset=utf8")
	if err != nil {
		panic(err.Error())
	}

	// Read logs and upload data
	logsFile, err := os.Open("log.txt")
	if err != nil {
		panic(err.Error())
	}

	byteValue, _ = ioutil.ReadAll(logsFile)
	logsStr := string(byteValue)

	err = logsFile.Close()
	if err != nil {
		panic(err.Error())
	}

	logs := strings.Split(logsStr, "\n")

	for _, s := range logs {
		if s == "" {
			continue
		}
		stat := CrawlingStats{}
		stat.Status, err = strconv.Atoi(strings.Split(s, " ")[0])
		if err != nil {
			panic(err.Error())
		}
		stat.TimeMs, err = strconv.Atoi(strings.Split(strings.Split(s, "Parsed by ")[1], " ")[0])
		if err != nil {
			panic(err.Error())
		}
		spl := strings.Split(s, "url: ")
		stat.Url = spl[len(spl)-1]

		_, err = conn.Exec("INSERT INTO " + DB_TABLE_NAME +
			" (status, time_ms, url) VALUE (" + strconv.Itoa(stat.Status) + ", " + strconv.Itoa(stat.TimeMs) + ", '" + stat.Url + "')")
		if err != nil {
			panic(err.Error())
		}

		fmt.Println("Uploaded " + stat.Url)
	}
	fmt.Print("Done by ")
	fmt.Print(time.Now().Sub(start).Nanoseconds() / 1E+6)
	fmt.Println(" ms")
}
