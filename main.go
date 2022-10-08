package main

import (
	"fmt"
	"github.com/vicanso/go-charts/v2"
	"github.com/xuri/excelize/v2"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
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
	tmpPath := "./tmp"
	err := os.MkdirAll(tmpPath, 0700)
	if err != nil {
		return err
	}

	file := filepath.Join(tmpPath, "radar-chart.png")
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
		sumList := 0
		for _, cell := range val.List {
			cellValue, err := f.GetCellValue(f.GetSheetName(0), cell)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println("valueeeeeee: ", cellValue)
			fmt.Println("evalvalue: ", evalPoints(cellValue))
			sumList += evalPoints(cellValue)
			fmt.Println("sumList: ", sumList)

		}
		Values[i] = float64(sumList * 1000)
		fmt.Println("xxxxxxx-> ", Values[i])
	}

	fmt.Println(Legend)
	fmt.Println(RadarIndicator)
	fmt.Println(RadarFixed)
	fmt.Println(Values)
	values := [][]float64{
		Values,
		RadarFixed,
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

	/*Get value from cell by given worksheet name and cell reference.
	cell, err := f.GetCellValue(f.GetSheetName(0), "B2")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(cell)
	// Get all the rows in the Sheet1.
	rows, err := f.GetRows("Respuestas de formulario 1")
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, row := range rows {
		for _, colCell := range row {
			fmt.Print(colCell, "\t")
		}
		fmt.Println()
	}*/

}
