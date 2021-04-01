package main

import (
	"log"
	"os"
	"time"

	"github.com/jszwec/csvutil"
	"github.com/karashiiro/bingode"
	"github.com/xivapi/godestone/v2"
)

// The number of characters to attempt to fetch. The program will
// usually get much less than this, since many people keep their
// achievements private.
var characterCount uint32 = 10

// Number of goroutines to execute at once. Setting this too high will
// get you IP-blocked for a couple of days (can still log into the game).
var parallelism uint32 = 3

// Number of characters to skip in iteration. Multiply this by
// the character count to get the maximum ID the program will attempt
// to fetch.
var sampleRate uint32 = 2

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

	now := time.Now()
	for i := range ids {
		acc, _, err := scraper.FetchCharacterAchievements(i)
		if err == nil {
			oldest := now
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

	charsPerGoroutine := characterCount / parallelism

	creationInfo := make([]*IDCreationInfo, 0)
	creationInfoChans := make([]chan []*IDCreationInfo, parallelism)
	for i := uint32(0); i < parallelism; i++ {
		idChan := make(chan uint32, charsPerGoroutine)
		creationInfoChans[i] = make(chan []*IDCreationInfo, 1)

		go getCreationInfos(scraper, idChan, creationInfoChans[i])

		startID := uint32(1+i*charsPerGoroutine) * sampleRate
		endID := uint32((i+1)*charsPerGoroutine) * sampleRate

		for j := startID; j <= endID; j += sampleRate {
			idChan <- j
		}

		// Handle remainder
		if i == parallelism-1 {
			remainder := characterCount % parallelism
			endID += sampleRate
			for j := uint32(0); j < remainder; j++ {
				idChan <- endID + sampleRate*j
			}
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
