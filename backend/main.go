package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
)

type collection struct {
	HarvestsByDAU  map[string][]DAUHarvest
	HarvestsByUnit map[int][]UnitHarvest
}

type DAUHarvest struct {
	Season                      string
	DAU                         string
	TotalHunterEstimate         int
	TotalHarvestEstimate        int
	TotalRecreationDaysEstimate int
}

type UnitHarvest struct {
	Season  string `json:"season"`
	Bulls   int    `json:"bulls"`
	Cows    int    `json:"cows"`
	Calves  int    `json:"calves"`
	Harvest int    `json:"harvest"`
	Hunters int    `json:"hunters"`
	Success int    `json:"success"`
	RecDays int    `json:"recdays"`
}

const (
	ALLMANNERS      = "all manners of take"
	ALLRIFLESEASONS = "all rifle"
	ALLRFW          = "all ranching for wildlife"
	EARLYHICOUNTRY  = "early high country"
	LATESEASONS     = "late"
	PRIVATEONLY     = "private land only"
	FIRSTRIFLE      = "first rifle"
	SECONDRIFLE     = "second rifle"
	THIRDRIFLE      = "third rifle"
	FOURTHRIFLE     = "fourth rifle"
	ALLARCHERY      = "archery"
	MUZZLELOADER    = "muzzleloader"
)

//general table ending signatures:
// blank line[0] column or line[0] = total
var coll collection

//TODO: this could totally be regex, capturing elements of an entire string such as 2017 Elk Harvest, Hunters and Percent Success for All Muzzleloader Seasons
func getSeason(line []string, season, animal string, year int) string {
	//season defining lines have no second element
	if len(line) < 3 || !strings.HasPrefix(strings.ToLower(line[0]), fmt.Sprintf("%v %v harvest", year, animal)) {
		return season
	}

	columnsToCheck := []int{0, 2}

	for _, ctc := range columnsToCheck {
		log.Printf("%+v\n", line[ctc])
		seasonEntry := strings.ToLower(line[ctc])
		if strings.Contains(seasonEntry, ALLMANNERS) {
			return ALLMANNERS
		} else if strings.Contains(seasonEntry, ALLRIFLESEASONS) {
			return ALLRIFLESEASONS
		} else if strings.Contains(seasonEntry, ALLRFW) {
			return ALLRFW
		} else if strings.Contains(seasonEntry, EARLYHICOUNTRY) {
			return EARLYHICOUNTRY
		} else if strings.Contains(seasonEntry, LATESEASONS) {
			return LATESEASONS
		} else if strings.Contains(seasonEntry, PRIVATEONLY) {
			return PRIVATEONLY
		} else if strings.Contains(seasonEntry, FIRSTRIFLE) {
			return FIRSTRIFLE
		} else if strings.Contains(seasonEntry, SECONDRIFLE) {
			return SECONDRIFLE
		} else if strings.Contains(seasonEntry, THIRDRIFLE) {
			return THIRDRIFLE
		} else if strings.Contains(seasonEntry, FOURTHRIFLE) {
			return FOURTHRIFLE
		} else if strings.Contains(seasonEntry, ALLARCHERY) {
			return ALLARCHERY
		} else if strings.Contains(seasonEntry, MUZZLELOADER) {
			return MUZZLELOADER
		}
	}

	return "invalid"
}

func UnitRequest(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	unit, err := strconv.Atoi(ps.ByName("unit"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(coll.HarvestsByUnit[unit])
}

func CollRequest(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	json.NewEncoder(w).Encode(coll)
}

func main() {
	//serve
	router := httprouter.New()
	router.GET("/coll", CollRequest)
	router.GET("/unit/:unit", UnitRequest)

	// dir := "./huntData/CO2017.csv"

	// files, err := ioutil.ReadDir(dir)
	// if err != nil {
	// 	panic(err)
	// }

	coll.HarvestsByDAU = make(map[string][]DAUHarvest)
	coll.HarvestsByUnit = make(map[int][]UnitHarvest)

	// for _, f := range files {
	// log.Println("opening " + f.Name())
	csvFile, _ := os.Open("./huntData/CO2017.csv")
	reader := csv.NewReader(bufio.NewReader(csvFile))

	var season string
	onDataEntries := false

	//phase 1 collect by unit
	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		}

		if line[0] == "" || strings.ToLower(line[0]) == "total" {
			onDataEntries = false
		}

		//figure out what season we are on
		season = getSeason(line, season, "elk", 2017)
		if season == "invalid" {
			continue
		}

		//check for column titles
		if strings.ToLower(line[0]) == "unit" && strings.ToLower(line[1]) == "bulls" && strings.ToLower(line[2]) == "cows" {
			onDataEntries = true
			continue
		}

		if onDataEntries {
			unit, _ := strconv.Atoi(line[0])
			bulls, _ := strconv.Atoi(strings.Replace(line[1], ",", "", -1))
			cows, _ := strconv.Atoi(strings.Replace(line[2], ",", "", -1))
			calves, _ := strconv.Atoi(strings.Replace(line[3], ",", "", -1))
			harvest, _ := strconv.Atoi(strings.Replace(line[4], ",", "", -1))
			hunters, _ := strconv.Atoi(strings.Replace(line[5], ",", "", -1))
			success, _ := strconv.Atoi(line[6])
			recDays, _ := strconv.Atoi(line[7])
			//we are on a good line w/ int values
			coll.HarvestsByUnit[unit] = append(coll.HarvestsByUnit[unit], UnitHarvest{
				Season:  season,
				Bulls:   bulls,
				Cows:    cows,
				Calves:  calves,
				Harvest: harvest,
				Hunters: hunters,
				Success: success,
				RecDays: recDays,
			})
		}
		//else if check if line[0] begins with DAU signature
	}

	// log.Printf("COLL: %+v\n", coll)
	// }
	log.Fatal(http.ListenAndServe(":8080", router))
}
