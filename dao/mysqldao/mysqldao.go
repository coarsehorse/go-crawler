package mysqldao

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	CRAWLING_TASK_TABLE      = "crawling_task"
	ESTIMATOR_TABLE          = "estimator"
	ESTIMATOR_SETTINGS_TABLE = "estimator_settings"
	CRAWLED_LINK_EST_TABLE   = "crawled_link_estimation"
	DB_CREDENTIALS_FILENAME  = "db_credentials.json"
	CONNECTION_TIMEOUT       = 5
	MAX_CONNECTIONS          = 5
)

// Crawling statuses representation
const (
	IN_QUEUE    = "in_queue"
	IN_PROGRESS = "in_progress"
	DONE        = "done"
)

type CrawlingTask struct {
	Id                int    `json:"id"`
	IdEstimator       int    `json:"idEstimator"`
	Url               string `json:"url"`
	IncludeSubdomains bool   `json:"IncludeSubdomains"`
	Status            string `json:"status"`
	Hidden            bool   `json:"hidden"`
}

type Estimation struct {
	Id              int            `json:"id"`
	Url             string         `json:"url"`
	CrawledPagesNum sql.NullInt64  `json:"crawledPagesNum"`
	StartDate       string         `json:"startDate"`
	EndDate         sql.NullString `json:"endDate"`
	CrawlingTime    sql.NullInt64  `json:"crawlingTime"`
	ResultsLink     string         `json:"resultsLink"`
}

type DBCredentials struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	HostAddress string `json:"hostAddress"`
	Port        int    `json:"port"`
	DbName      string `json:"dbName"`
}

type EstimatorSetting struct {
	Id          int             `json:"id"`
	ServiceName string          `json:"serviceName"`
	Design      sql.NullFloat64 `json:"design"`
	Markup      sql.NullFloat64 `json:"markup"`
	Development sql.NullFloat64 `json:"development"`
	ContentM    sql.NullFloat64 `json:"contentM"`
	Testing     sql.NullFloat64 `json:"testing"`
	Management  sql.NullFloat64 `json:"management"`
	Hidden      bool            `json:"hidden"`
}

type CrawledLinkEstimation struct {
	Id             int             `json:"id"`
	CrawlingTaskId int             `json:"crawlingTaskId"`
	Link           sql.NullString  `json:"link"`
	TypeId         sql.NullInt64   `json:"typeId"`
	Design         sql.NullFloat64 `json:"design"`
	Markup         sql.NullFloat64 `json:"markup"`
	Development    sql.NullFloat64 `json:"development"`
	ContentM       sql.NullFloat64 `json:"contentM"`
	Testing        sql.NullFloat64 `json:"testing"`
	Management     sql.NullFloat64 `json:"management"`
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

	// https://github.com/go-sql-driver/mysql/issues/461
	conn.SetConnMaxLifetime(time.Minute * CONNECTION_TIMEOUT)
	conn.SetMaxIdleConns(MAX_CONNECTIONS)
	conn.SetMaxOpenConns(MAX_CONNECTIONS)

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
		err = tasks.Scan(&task.Id, &task.IdEstimator, &task.Url, &task.IncludeSubdomains, &task.Status, &task.Hidden)
		if err != nil {
			return nil, err
		}
		activeTasks = append(activeTasks, task)
	}

	err = tasks.Close()
	if err != nil {
		return nil, err
	}

	return activeTasks, nil
}

func UpdateCrawlingTaskById(task CrawlingTask, conn *sql.DB) (err error) {
	stmt, err := conn.Prepare("UPDATE " + CRAWLING_TASK_TABLE + " SET " +
		"id_estimator=?, " +
		"url=?, " +
		"include_subdomains=?, " +
		"status=?, " +
		"hidden=? " +
		"WHERE id=?")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(task.IdEstimator, task.Url, task.IncludeSubdomains, task.Status, task.Hidden, task.Id)
	if err != nil {
		return err
	}

	err = stmt.Close()
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

	err = stmt.Close()
	if err != nil {
		return err
	}

	return nil
}

// Returns first row id from EstimatorSetting table as default
func GetDefaultEstimatorSetting(conn *sql.DB) (setting EstimatorSetting, err error) {
	firstSetting, err := conn.Query("SELECT * " +
		"FROM " + ESTIMATOR_SETTINGS_TABLE + " " +
		"WHERE `hidden` IS FALSE " +
		"LIMIT 1")
	if err != nil {
		return EstimatorSetting{}, err
	}

	defSetting := EstimatorSetting{}

	if !firstSetting.Next() {
		return EstimatorSetting{}, errors.New("not able to scan next estimator setting")
	}
	err = firstSetting.Scan(&defSetting.Id, &defSetting.ServiceName, &defSetting.Design, &defSetting.Markup,
		&defSetting.Development, &defSetting.ContentM, &defSetting.Testing, &defSetting.Management, &defSetting.Hidden)
	if err != nil {
		return EstimatorSetting{}, err
	}

	err = firstSetting.Close()
	if err != nil {
		return EstimatorSetting{}, err
	}

	return defSetting, nil
}

func nullableStringOrNull(nullable sql.NullString) string {
	if nullable.Valid {
		return nullable.String
	} else {
		return "NULL"
	}
}

func nullableIntOrNull(nullable sql.NullInt64) string {
	if nullable.Valid {
		return fmt.Sprintf("%d", nullable.Int64)
	} else {
		return "NULL"
	}
}

func nullableFloatOrNull(nullable sql.NullFloat64) string {
	if nullable.Valid {
		return fmt.Sprintf("%8.3f", nullable.Float64)
	} else {
		return "NULL"
	}
}

func InsertIntoCrawledLinkEstimation(linkEstimations []CrawledLinkEstimation, conn *sql.DB) (err error) {
	batchInsertHeader := "INSERT INTO " + CRAWLED_LINK_EST_TABLE +
		" (`crawling_task_id`, `link`, `type_id`, `design`, `markup`, " +
		"`development`, `content_m`, `testing`, `management`) VALUES "
	batch := make([]string, 0, len(linkEstimations))

	for _, e := range linkEstimations {
		batch = append(batch, "("+
			strconv.Itoa(e.CrawlingTaskId)+", "+
			"'"+nullableStringOrNull(e.Link)+"', "+
			nullableIntOrNull(e.TypeId)+", "+
			nullableFloatOrNull(e.Design)+", "+
			nullableFloatOrNull(e.Markup)+", "+
			nullableFloatOrNull(e.Development)+", "+
			nullableFloatOrNull(e.ContentM)+", "+
			nullableFloatOrNull(e.Testing)+", "+
			nullableFloatOrNull(e.Management)+")")
	}

	stmt, err := conn.Prepare(batchInsertHeader + strings.Join(batch, ", "))
	if err != nil {
		return err
	}

	_, err = stmt.Exec()
	if err != nil {
		return err
	}

	err = stmt.Close()
	if err != nil {
		return err
	}

	return nil
}
