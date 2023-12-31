package scraper

import (
	"fmt"
	"strings"
	"time"

	"github.com/Virees/goru/flags"
	"github.com/fatih/color"
	"github.com/gocolly/colly/v2"
	"github.com/google/uuid"
)

func SetupTagsCollector(flags flags.GoruFlags) *colly.Collector {
	tagsColly := colly.NewCollector(
		colly.AllowedDomains("safebooru.org", "danbooru.donmai.us", "gelbooru.com"),
		colly.Async(true),
	)

	// Set config
	tagsColly.SetRequestTimeout(30 * time.Second)
	tagsColly.Limit(&colly.LimitRule{Parallelism: 4, DomainGlob: "*"})

	// Handlers
	tagsColly.OnResponse(onResponse)
	tagsColly.OnError(onError)
	tagsColly.OnScraped(onScraped)

	switch flags.Site.Name {
	case "safe", "safebooru":
		tagsColly.OnHTML(SAFE_TAGS_CONTAINER, setupOnTags(SAFE_TAGS_LOCATION, flags.TagsTypeList))
	case "dan", "danbooru":
		tagsColly.OnHTML(DAN_TAGS_CONTAINER, setupOnTags(DAN_TAGS_LOCATION, flags.TagsTypeList))
	case "gel", "gelbooru":
		tagsColly.OnHTML(GEL_TAGS_CONTAINER, setupOnTags(GEL_TAGS_LOCATION, flags.TagsTypeList))
	}

	return tagsColly
}

// Html handler
func setupOnTags(tagsLocation map[string]string, tagsToDownload flags.TagsType) func(*colly.HTMLElement) {
	return func(e *colly.HTMLElement) {
		if len(e.Request.Headers.Values("Referer")) != 0 {
			return
		}

		// Get output folder path
		outDirPath := string(e.Request.Ctx.Get("outFolder"))

		// Open/Create file
		file, err := CreateFile(outDirPath, uuid.New().String())
		if err != nil {
			fmt.Printf(color.YellowString("File create error: %s"), err)
			return
		}
		defer file.Close()

		// Extract tags
		extractedTags, err := ParseTags(tagsLocation, tagsToDownload, e)
		if err != nil {
			fmt.Printf(color.YellowString("Parse tags error: %s\n"), err)
			return
		}

		// Save tags to file
		file.WriteString(strings.Join(extractedTags, ", "))
	}
}

// Response handles
func onResponse(res *colly.Response) {
	referer := res.Request.Headers.Values("Referer")
	if len(referer) != 0 {
		fmt.Printf(color.RedString("TagsCollector: Response URL is redirected or not found. Tags will not be downloaded for URL: %s\n"), referer[0])
		return
	}

	fmt.Printf(color.GreenString("TagsCollector: Got a response from: %s (HTTP Code: %d)\n"), res.Request.URL, res.StatusCode)
}

func onError(res *colly.Response, e error) {
	fmt.Printf(color.RedString("TagsCollector: %s entering site %s (HTTP Code: %d)\n"), e, res.Request.URL, res.StatusCode)
}

func onScraped(res *colly.Response) {

}
