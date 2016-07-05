package rgeogo

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
)

type (
	rgeo struct {
		M map[int64]geo
		I int64Slice
	}

	geo struct {
		Country    string
		Region     string
		City       string
		PostalCode string
	}
)

type int64Slice []int64

func (s int64Slice) Len() int           { return len(s) }
func (s int64Slice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s int64Slice) Less(i, j int) bool { return s[i] < s[j] }
func (s int64Slice) Sort()              { sort.Sort(s) }
func (s int64Slice) Search(x int64) int {
	index := sort.Search(len(s), func(i int) bool { return s[i] >= x })
	if index < len(s) {
		return index
	} else {
		// this design is purely for the only use case in the module
		return 0
	}
}

func Setup(dataFolder string) error {
	files, err := ioutil.ReadDir(dataFolder)
	if err != nil {
		return err
	}

	for _, file := range files {
		fp, err := os.Open(dataFolder + "/" + file.Name())
		if err != nil {
			return fmt.Errorf("Error opening file %s: %s", dataFolder+"/"+file.Name(), err.Error())
		}

		// index each point and add them to the Object
		s := bufio.NewScanner(fp)
		var line string
		for s.Scan() {
			line = s.Text()
			if line != "" {
				content := strings.Split(line, ",")

				postal := content[0]
				lat, _ := strconv.ParseFloat(content[1], 64)
				lon, _ := strconv.ParseFloat(content[2], 64)
				city := content[3]
				region := content[4]

				// assume files are named as COUNTRY.csv
				geoIndex := encodeUInt64(lat, lon)
				if geoIndex != -1 {
					rgeoObject.M[geoIndex] = geo{
						Country:    strings.Split(file.Name(), ".")[0],
						Region:     region,
						City:       city,
						PostalCode: postal,
					}
				}

			}
		}

		fp.Close()
	}

	for k, _ := range rgeoObject.M {
		rgeoObject.I = append(rgeoObject.I, k)
	}
	rgeoObject.I.Sort()

	return nil
}
