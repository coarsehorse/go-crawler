package main

import (
	"database/sql"
	"go-crawler/crawler"
	"go-crawler/dao/mysqldao"
	"go-crawler/utils"
	"go-crawler/validator"
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

		// Get estimator settings table first row id as default id
		defSett, err := mysqldao.GetDefaultEstimatorSetting(connection)
		utils.CheckError(err)

		// Sort by id(less id - added earlier)
		sort.Slice(activeTasks[:], func(i, j int) bool {
			return activeTasks[i].Id < activeTasks[j].Id
		})
		for _, task := range activeTasks {
			if task.Status == mysqldao.IN_QUEUE {
				log.Print("[task_tracker]\tFound new crawling task in queue with id: ", task.Id)

				// Update status
				task.Status = mysqldao.IN_PROGRESS
				err = mysqldao.UpdateCrawlingTaskById(task, connection)
				utils.CheckError(err)
				log.Print("[task_tracker]\tCrawling task status has been updated to: '", task.Status,
					"', task id: ", task.Id)

				taskUrl := utils.AddFollowingSlashToUrl(task.Url)

				// Check the url to crawl
				if !utils.IsUrl(taskUrl) {
					// Update skipped crawling task to DONE
					task.Status = mysqldao.DONE
					err = mysqldao.UpdateCrawlingTaskById(task, connection)
					utils.CheckError(err)
					log.Print("[task_tracker]\tCrawling task has been skipped because of not valid url: \"",
						taskUrl, "\", task id: ", task.Id)

					continue
				}

				// Construct validator from task string rules
				var exceptions []string
				if task.Exceptions.Valid {
					exceptions = strings.Split(task.Exceptions.String, "\n")
					exceptions = utils.TrimArray(exceptions)
				}

				var allowances []string
				if task.Allowances.Valid {
					allowances = strings.Split(task.Allowances.String, "\n")
					allowances = utils.TrimArray(allowances)
				}

				taskValidator := validator.NewValidator(exceptions, allowances)
				log.Println("[task_tracker]\tValidation rules: ", taskValidator, ", task id: ", task.Id)

				// Perform a task
				start := time.Now() // get start time

				linksToCrawl := []string{taskUrl}
				// Read the sitemap
				sitemap, err := crawler.GetLinksFromSitemap(taskUrl)
				if err == nil {
					linksToCrawl = utils.UniqueStringSlice(append(sitemap, taskUrl))
				}

				// First time validation(there are random number of links in the sitemap.xml)
				// Validate linksToCrawl
				linksToCrawl = utils.FilterSlice(linksToCrawl, func(link string) bool {
					return taskValidator.IsValid(link)
				})
				// Validate with domain pattern, subdomains handled
				domain := utils.ExtractDomain(taskUrl)
				linksToCrawl = utils.FilterLinksNotInDomain(domain, linksToCrawl, task.IncludeSubdomains)
				// Filter out image links
				linksToCrawl = utils.FilterLinksToImages(linksToCrawl)

				crawledLevels := crawler.Crawl(linksToCrawl, []string{},
					[]crawler.CrawledLevel{}, task.IncludeSubdomains, taskValidator)
				end := time.Now()                                      // get end time
				executionTimeMs := end.Sub(start).Nanoseconds() / 1E+6 // evaluate execution time
				log.Print("[task_tracker]\tCrawling task was performed, task id: ", task.Id)

				// Update crawled link estimation table
				crawledLinks := crawler.ExtractUniqueLinks(crawledLevels)
				crawledLinks = utils.RemoveEmptyStrings(crawledLinks)
				sort.Slice(crawledLinks[:], func(i, j int) bool {
					return crawledLinks[i] < crawledLinks[j]
				})
				linkEstimations := make([]mysqldao.CrawledLinkEstimation, 0, len(crawledLinks))
				for _, link := range crawledLinks {
					linkEstimations = append(linkEstimations, mysqldao.CrawledLinkEstimation{
						CrawlingTaskId: task.Id,
						Link:           sql.NullString{Valid: true, String: link},
						TypeId:         sql.NullInt64{Valid: true, Int64: int64(defSett.Id)},
						Design:         defSett.Design,
						Markup:         defSett.Markup,
						Development:    defSett.Development,
						ContentM:       defSett.ContentM,
						Testing:        defSett.Testing,
						Management:     defSett.Management,
					})
				}
				err = mysqldao.InsertIntoCrawledLinkEstimation(linkEstimations, connection)
				utils.CheckError(err)
				log.Print("[task_tracker]\t'"+mysqldao.CRAWLED_LINK_EST_TABLE+"' table has been appended(", len(linkEstimations),
					" rows) with results of crawling task with id: ", task.Id)

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
				log.Print("[task_tracker]\t'"+mysqldao.ESTIMATOR_TABLE+"' table record with id: ", task.IdEstimator,
					" was updated with results by crawling task with id: ", task.Id)

				// Update crawling task status
				task.Status = mysqldao.DONE
				err = mysqldao.UpdateCrawlingTaskById(task, connection)
				utils.CheckError(err)
				log.Print("[task_tracker]\tCrawling task status has been updated to: '"+task.Status+
					"', task id: ", task.Id)
			}
		}

		time.Sleep(3 * time.Second)
	}
}
