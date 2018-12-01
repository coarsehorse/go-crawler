package main

import (
	"database/sql"
	"go-crawler/crawler"
	"go-crawler/dao/mysqldao"
	"go-crawler/utils"
	"log"
	"sort"
	"strings"
	"time"
)

func main() {
	log.Print("Starting...")
	connection, err := mysqldao.GetConnection()
	utils.CheckError(err)

	for {
		// Get current tasks
		activeTasks, err := mysqldao.GetActiveTasks(connection)
		utils.CheckError(err)

		// Sort by id(less id - added earlier)
		sort.Slice(activeTasks[:], func(i, j int) bool {
			return activeTasks[i].Id < activeTasks[j].Id
		})
		for _, task := range activeTasks {
			if task.Status == mysqldao.IN_QUEUE {
				// Update status
				task.Status = mysqldao.IN_PROGRESS
				err = mysqldao.UpdateCrawlingTaskById(task, connection)
				utils.CheckError(err)
				log.Print("Found new task id: ", task.Id, ", updated with status: ", task.Status)

				// Perform a task
				start := time.Now() // get start time
				crawledLevels := crawler.Crawl([]string{task.Url}, []string{}, []crawler.CrawledLevel{})
				end := time.Now()                                      // get end time
				executionTimeMs := end.Sub(start).Nanoseconds() / 1E+6 // evaluate execution time

				// Extract crawled links
				crawledLinks := crawler.ExtractUniqueLinks(crawledLevels)

				// Update crawling task table
				task.Status = mysqldao.DONE
				task.CrawledLinks.Valid = true
				task.CrawledLinks.String = strings.Join(crawledLinks, "\n")
				err = mysqldao.UpdateCrawlingTaskById(task, connection)
				utils.CheckError(err)

				// Update estimator table
				nullCrawledLinksNum := sql.NullInt64{
					Valid: true,
					Int64: int64(len(crawledLinks)),
				}
				nullTime := sql.NullInt64{
					Valid: true,
					Int64: executionTimeMs,
				}
				nullEndTime := sql.NullString{
					Valid:  true,
					String: end.Format("2006-01-02 15:04:05"), // mySQL mask
				}
				err = mysqldao.UpdateEstimatorById(task.IdEstimator,
					nullCrawledLinksNum, nullEndTime, nullTime, connection)
				utils.CheckError(err)
				log.Print("Estimator table id: ", task.IdEstimator,
					"was updated with results by crawling task id: ", task.Id)

				log.Print("Task id: ", task.Id, " has been performed and updated with status: ", task.Status)
			}
		}

		time.Sleep(3 * time.Second)
	}
}
