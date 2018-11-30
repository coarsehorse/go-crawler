package main

import (
	_ "github.com/go-sql-driver/mysql"
	"goCrawler/mySQLDao"
	"goCrawler/utils"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

/*
create table analitycs
(
	id int not null AUTO_INCREMENT,
	status int null,
	time_ms int null,
	url varchar(1000) null,
	constraint analitycs_pk
		primary key (id)
);
*/

const (
	DB_TABLE_NAME = "analytics1"
	LOG_FILENAME  = "log_part.log"
)

type CrawlingStats struct {
	Status int
	TimeMs int
	Url    string
}

func main() {
	start := time.Now()

	// Open connection
	conn, err := mySQLDao.GetConnection()
	utils.CheckError(err)

	// Read logs and upload data
	logsFile, err := os.Open(LOG_FILENAME)
	if err != nil {
		panic(err.Error())
	}

	byteValue, _ := ioutil.ReadAll(logsFile)
	logsStr := string(byteValue)

	err = logsFile.Close()
	if err != nil {
		panic(err.Error())
	}

	logs := strings.Split(logsStr, "\n")

	temp1 := logs[len(logs)-20 : len(logs)]
	log.Print(strings.Join(temp1[:2], "\n"))

	batch := make([]string, 0, len(logs))

	// Construct batch of SQL requests
	for _, s := range logs {
		if s == "" || // skip not valid strings
			strings.Contains(s, "Starting crawl ") ||
			strings.Contains(s, "Crawled with error ") ||
			strings.Contains(s, "Crawled with error ") {
			continue
		}

		stat := CrawlingStats{}
		tabs := strings.Split(s, "\t")

		// Treat error as 0 status
		if strings.Contains(tabs[1], "ERROR") {
			tabs[1] = "0"
		}

		stat.Status, err = strconv.Atoi(strings.Split(tabs[1], " ")[0])
		if err != nil {
			panic(err.Error())
		}
		stat.TimeMs, err = strconv.Atoi(strings.Split(tabs[2], " ")[2])
		if err != nil {
			panic(err.Error())
		}

		stat.Url = strings.Split(tabs[3], "url: ")[1]

		batch = append(batch, "("+strconv.Itoa(stat.Status)+
			", "+strconv.Itoa(stat.TimeMs)+", '"+stat.Url+"')")
	}

	batchHeader := "INSERT INTO " + DB_TABLE_NAME +
		" (status, time_ms, url) VALUES "

	chunkSize := 500
	// Perform batch inserts
	for i := 0; i < len(batch); i += chunkSize {
		end := i + chunkSize
		if end > len(batch) {
			end = len(batch)
		}

		fooBatch := strings.Join(batch[i:end], ", ")
		_, err = conn.Exec(batchHeader + fooBatch)
		if err != nil {
			panic(err.Error())
		}
		log.Print("Uploaded data[", i, ":", end, "]")
	}

	log.Print("Done by ", time.Now().Sub(start).Nanoseconds()/1E+6, " ms")
}
