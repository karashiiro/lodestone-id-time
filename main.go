package main

import (
	"fmt"
	"time"

	"github.com/jszwec/csvutil"
	"github.com/karashiiro/bingode"
	"github.com/xivapi/godestone/v2"
)

type IDCreationInfo struct {
	ID        uint64    `csv:"id"`
	CreatedAt time.Time `csv:"createdAt"`
}

func main() {
	bin := bingode.New()
	scraper := godestone.NewScraper(bin, godestone.EN)

	creationInfo := make([]*IDCreationInfo, 0)

	for i := uint32(1); i <= 400; i++ {
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
					ID:        1,
					CreatedAt: oldest,
				})
			}
		}
	}

	b, err := csvutil.Marshal(creationInfo)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Println(string(b))
}
