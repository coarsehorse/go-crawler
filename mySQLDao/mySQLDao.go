package mySQLDao

import (
	"database/sql"
	"encoding/json"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"os"
	"strconv"
)

const (
	CRAWLING_TASK_TABLE     = "crawling_task"
	ESTIMATOR_TABLE         = "estimator"
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

type Estimation struct {
	Id              int            `json:"id"`
	Url             string         `json:"url"`
	CrawledPagesNum sql.NullInt64  `json:"crawled_pages_num"`
	StartDate       string         `json:"start_date"`
	EndDate         sql.NullString `json:"end_date"`
	CrawlingTime    sql.NullInt64  `json:"crawling_time"`
	ResultsLink     string         `json:"results_link"`
}

type DBCredentials struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	HostAddress string `json:"host_address"`
	Port        int    `json:"port"`
	DbName      string `json:"db_name"`
}

func GetConnection() (conn *sql.DB, err error) {
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

func GetActiveTasks(conn *sql.DB) (activeTasks []CrawlingTask, err error) {
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

func UpdateCrawlingTaskById(task CrawlingTask, conn *sql.DB) (err error) {
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

func UpdateEstimatorById(id int, crawledPagesNum sql.NullInt64, endDate sql.NullString,
	crawlingTime sql.NullInt64, conn *sql.DB) (err error) {
	stmt, err := conn.Prepare("UPDATE " + ESTIMATOR_TABLE + " SET " +
		"crawled_pages_num=?, " +
		"end_date=?, " +
		"crawling_time=? " +
		"WHERE id=?")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(crawledPagesNum, endDate, crawlingTime, id)
	if err != nil {
		return err
	}

	return nil
}
