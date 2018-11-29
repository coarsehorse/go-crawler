package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	CRAWLING_TASK_TABLE     = "crawling_task"
	DB_CREDENTIALS_FILENAME = "db_credentials.json"
)

// Crawling statuses representation
const (
	IN_QUEUE    = "in_queue"
	IN_PROGRESS = "in_progress"
	DONE        = "done"
)

type CrawlingTask struct {
	Id           int            `json:"id"`
	IdEstimator  int            `json:"id_estimator"`
	Url          string         `json:"url"`
	Status       string         `json:"status"`
	Hidden       bool           `json:"hidden"`
	CrawledLinks sql.NullString `json:"crawled_links"`
}

type DBCredentials struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	HostAddress string `json:"host_address"`
	Port        int    `json:"port"`
	DbName      string `json:"db_name"`
}

func checkErr(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func getConnection() (conn *sql.DB, err error) {
	// Open json file with credentials
	jsonFile, err := os.Open(DB_CREDENTIALS_FILENAME)
	if err != nil {
		return nil, err
	}

	byteValue, _ := ioutil.ReadAll(jsonFile)

	// Unmarshal data
	var cred DBCredentials
	err = json.Unmarshal(byteValue, &cred)
	if err != nil {
		return nil, err
	}

	// Close opened credentials file
	err = jsonFile.Close()
	if err != nil {
		return nil, err
	}

	// Open connection
	conn, err = sql.Open("mysql",
		cred.Username+":"+cred.Password+"@tcp("+cred.HostAddress+":"+
			strconv.Itoa(cred.Port)+")/"+cred.DbName+"?charset=utf8")
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func getActiveTasks(conn *sql.DB) (activeTasks []CrawlingTask, err error) {
	// Select all active tasks
	activeTasks = make([]CrawlingTask, 0)
	tasks, err := conn.Query("SELECT * FROM " + CRAWLING_TASK_TABLE + " WHERE `hidden` IS FALSE")
	if err != nil {
		return nil, err
	}
	// Map data to CrawlingTask objects
	for tasks.Next() {
		task := CrawlingTask{}
		err = tasks.Scan(&task.Id, &task.IdEstimator, &task.Url, &task.Status, &task.Hidden, &task.CrawledLinks)
		if err != nil {
			return nil, err
		}
		activeTasks = append(activeTasks, task)
	}

	return activeTasks, nil
}

func updateById(task CrawlingTask, conn *sql.DB) (err error) {
	stmt, err := conn.Prepare("UPDATE " + CRAWLING_TASK_TABLE + " SET " +
		"id_estimator=?," +
		"url=?," +
		"status=?," +
		"hidden=?," +
		"crawled_links=? " +
		"WHERE id=?")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(task.IdEstimator, task.Url, task.Status, task.Hidden, task.CrawledLinks, task.Id)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	fmt.Println("Starting...")
	connection, err := getConnection()
	checkErr(err)

	for {
		// Get current tasks
		activeTasks, err := getActiveTasks(connection)
		checkErr(err)

		// Sort by id(less id - added earlier)
		sort.Slice(activeTasks[:], func(i, j int) bool {
			return activeTasks[i].Id < activeTasks[j].Id
		})
		for _, task := range activeTasks {
			if task.Status == IN_QUEUE {
				// Update status
				task.Status = IN_PROGRESS
				err = updateById(task, connection)
				checkErr(err)
				fmt.Println("Updated: ", task.Id, " with status: ", task.Status)

				// Perform a task
				crawledLevels := crawl([]string{task.Url}, []string{}, []CrawledLevel{})
				crawledLinks := extractUniqueLinks(crawledLevels)

				// Update task with results
				task.Status = DONE
				task.CrawledLinks.Valid = true
				task.CrawledLinks.String = strings.Join(crawledLinks, "\n")
				err = updateById(task, connection)
				checkErr(err)
				fmt.Println("Updated: ", task.Id, " with status: ", task.Status)
			}
		}

		time.Sleep(3 * time.Second)
	}
}
