package main

import (
	"fmt"
	"github.com/vicanso/go-charts/v2"
	"github.com/xuri/excelize/v2"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
)

type InputData struct {
	Data []struct {
		Name  string   `yaml:"Name"`
		Total int      `yaml:"Total"`
		List  []string `yaml:"List"`
	} `yaml:"Data"`
}

func (c *InputData) getConf() *InputData {

	yamlFile, err := ioutil.ReadFile("config/example.yaml")
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
	return c

}

func evalPoints(input string) int {
	if input != "" {
		if input == "Sí" || input == "Proactivo" {
			return 5
		} else {
			val, _ := strconv.Atoi(input)
			return val
		}
	} else {
		return 0
	}
}

func writeFile(buf []byte) error {
	tmpPath := "./"
	err := os.MkdirAll(tmpPath, 0700)
	if err != nil {
		return err
	}

	file := filepath.Join(tmpPath, "grafico.png")
	err = ioutil.WriteFile(file, buf, 0600)
	if err != nil {
		return err
	}
	return nil
}

func main() {

	var conf InputData
	conf.getConf()

	Legend := []string{"Situación Actual", "Margen de mejora"}
	RadarIndicator := make([]string, len(conf.Data))
	RadarFixed := make([]float64, len(conf.Data))
	RadarImprovement := make([]float64, len(conf.Data))
	Values := make([]float64, len(conf.Data))

	f, err := excelize.OpenFile("origin-data/input.xlsx")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		// Close the spreadsheet.
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	for i, val := range conf.Data {
		RadarIndicator[i] = val.Name
		RadarFixed[i] = float64(val.Total * 1000)
		RadarImprovement[i] = float64(val.Total*1000 - (rand.Intn(10-0)+0)*20000/100)
		sumList := 0
		for _, cell := range val.List {
			cellValue, err := f.GetCellValue(f.GetSheetName(0), cell)
			if err != nil {
				fmt.Println(err)
				return
			}
			sumList += evalPoints(cellValue)
		}
		Values[i] = float64(sumList * 1000)
	}

	fmt.Println(Legend)
	fmt.Println(RadarIndicator)
	fmt.Println(RadarFixed)
	fmt.Println(RadarImprovement)
	fmt.Println(Values)
	values := [][]float64{
		Values,
		RadarImprovement,
	}
	p, err := charts.RadarRender(
		values,
		charts.LegendLabelsOptionFunc(Legend),
		charts.RadarIndicatorOptionFunc(RadarIndicator, RadarFixed),
	)
	if err != nil {
		panic(err)
	}

	buf, err := p.Bytes()
	if err != nil {
		panic(err)
	}
	err = writeFile(buf)
	if err != nil {
		panic(err)
	}

}
