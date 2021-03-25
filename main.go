package main

import (
	"log"
	"os"
	"time"

	"github.com/jszwec/csvutil"
	"github.com/karashiiro/bingode"
	"github.com/xivapi/godestone/v2"
)

var characterCount uint32 = 20000
var idOffset uint32 = 10000000
var parallelism uint32 = 50
var sampleRate uint32 = 100

type Time struct {
	time.Time
}

const format = "2006/01/02 15:04:05"

func (t Time) MarshalCSV() ([]byte, error) {
	var b [len(format)]byte
	return t.AppendFormat(b[:0], format), nil
}

type IDCreationInfo struct {
	ID        uint32 `csv:"id"`
	CreatedAt Time   `csv:"created_at"`
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
					CreatedAt: Time{oldest},
					ID:        i,
				})
			}
		}
	}
	done <- creationInfo
}

func main() {
	bin := bingode.New()
	scraper := godestone.NewScraper(bin, godestone.EN)

	charsPerGoroutine := characterCount / parallelism * sampleRate

	creationInfo := make([]*IDCreationInfo, 0)
	creationInfoChans := make([]chan []*IDCreationInfo, parallelism)
	for i := uint32(0); i < parallelism; i++ {
		idChan := make(chan uint32, charsPerGoroutine)
		creationInfoChans[i] = make(chan []*IDCreationInfo, 1)

		go getCreationInfos(scraper, idChan, creationInfoChans[i])

		for j := uint32(1+i*charsPerGoroutine) + idOffset; j <= uint32((i+1)*charsPerGoroutine)+idOffset; j += sampleRate {
			idChan <- j
		}
		close(idChan)
	}

	for i := uint32(0); i < parallelism; i++ {
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
