package main

import (
	"log"
	"os"
	"time"

	"github.com/jszwec/csvutil"
	"github.com/karashiiro/bingode"
	"github.com/xivapi/godestone/v2"
)

var characterCount int = 2000
var parallelism int = 8

type IDCreationInfo struct {
	ID        uint32    `csv:"id"`
	CreatedAt time.Time `csv:"created_at"`
}

func getCreationInfos(scraper *godestone.Scraper, ids chan uint32, done chan []*IDCreationInfo) {
	creationInfo := make([]*IDCreationInfo, 0)
	for i := range ids {
		acc, _, err := scraper.FetchCharacterAchievements(i)
		if err == nil {
			oldest := time.Now()
			hasAny := false
			for _, a := range acc {
				if a.Date.Before(oldest) {
					oldest = a.Date
					hasAny = true
				}
			}

			if hasAny {
				creationInfo = append(creationInfo, &IDCreationInfo{
					ID:        i,
					CreatedAt: oldest,
				})
			}
		}
	}
	done <- creationInfo
}

func main() {
	bin := bingode.New()
	scraper := godestone.NewScraper(bin, godestone.EN)

	charsPerGoroutine := characterCount / parallelism

	creationInfo := make([]*IDCreationInfo, 0)
	creationInfoChans := make([]chan []*IDCreationInfo, parallelism)
	for i := 0; i < parallelism; i++ {
		idChan := make(chan uint32, charsPerGoroutine)
		creationInfoChans[i] = make(chan []*IDCreationInfo, 1)

		go getCreationInfos(scraper, idChan, creationInfoChans[i])

		for j := 1 + i*charsPerGoroutine; j <= (i+1)*charsPerGoroutine; j++ {
			idChan <- uint32(j)
		}
		close(idChan)
	}

	for i := 0; i < parallelism; i++ {
		curCreationInfo := <-creationInfoChans[i]
		close(creationInfoChans[i])
		creationInfo = append(creationInfo, curCreationInfo...)
	}

	b, err := csvutil.Marshal(creationInfo)
	if err != nil {
		log.Fatalln(err)
	}

	f, err := os.Create("characters.csv")
	if err != nil {
		log.Fatalln(err)
	}

	defer f.Close()

	_, err = f.Write(b)
	if err != nil {
		log.Fatal(err)
	}
}
