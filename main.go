package main

import (
	"fmt"
	"github.com/vicanso/go-charts/v2"
	"github.com/xuri/excelize/v2"
	"gopkg.in/yaml.v3"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

const MAX_UPLOAD_SIZE = 1024 * 1024 * 3 // 3MB

type InputData struct {
	Data []struct {
		Name  string   `yaml:"Name"`
		Total int      `yaml:"Total"`
		List  []string `yaml:"List"`
	} `yaml:"Data"`
}

type Progress struct {
	TotalSize int64
	BytesRead int64
}

// Write is used to satisfy the io.Writer interface.
// Instead of writing somewhere, it simply aggregates
// the total bytes on each read
func (pr *Progress) Write(p []byte) (n int, err error) {
	n, err = len(p), nil
	pr.BytesRead += int64(n)
	pr.Print()
	return
}

// Print displays the current progress of the file upload
func (pr *Progress) Print() {
	if pr.BytesRead == pr.TotalSize {
		fmt.Println("DONE!")
		return
	}

	fmt.Printf("File upload in progress: %d\n", pr.BytesRead)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	http.ServeFile(w, r, "web/index.html")
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 32 MB is the default used by FormFile
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// get a reference to the fileHeaders
	files := r.MultipartForm.File["file"]

	for _, fileHeader := range files {
		if fileHeader.Size > MAX_UPLOAD_SIZE {
			http.Error(w, fmt.Sprintf("The uploaded image is too big: %s. Please use an image less than 3MB in size", fileHeader.Filename), http.StatusBadRequest)
			return
		}

		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defer file.Close()

		buff := make([]byte, 512)
		_, err = file.Read(buff)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = file.Seek(0, io.SeekStart)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = os.MkdirAll("./uploads", os.ModePerm)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		f, err := os.Create(fmt.Sprintf("./uploads/input.xlsx"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		defer f.Close()

		pr := &Progress{
			TotalSize: fileHeader.Size,
		}

		_, err = io.Copy(f, io.TeeReader(file, pr))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var conf InputData
		conf.getConf()

		Legend := []string{"Situación Actual", "Margen de mejora"}
		RadarIndicator := make([]string, len(conf.Data))
		RadarFixed := make([]float64, len(conf.Data))
		RadarImprovement := make([]float64, len(conf.Data))
		Values := make([]float64, len(conf.Data))

		f2, err2 := excelize.OpenFile("./uploads/input.xlsx")
		if err2 != nil {
			fmt.Println(err2)
			return
		}
		defer func() {
			// Close the spreadsheet.
			if err := f2.Close(); err != nil {
				fmt.Println(err)
			}
		}()

		for i, val := range conf.Data {
			RadarIndicator[i] = val.Name
			RadarFixed[i] = float64(val.Total * 1000)
			RadarImprovement[i] = float64(val.Total*1000 - (rand.Intn(10-0)+0)*20000/100)
			sumList := 0
			for _, cell := range val.List {
				cellValue, err := f2.GetCellValue(f2.GetSheetName(0), cell)
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

		w.Write(buf)

	}
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

	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/upload", uploadHandler)

	if err := http.ListenAndServe(":80", mux); err != nil {
		log.Fatal(err)
	}

}
