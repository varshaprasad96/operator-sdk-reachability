package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/operator-framework/operator-registry/pkg/registry"
	"github.com/tealeg/xlsx"
)

var (
	searchDir = "/Users/vnarsing/go/src/github.com/varshaprasad96/operator-sdk-rechability/tmp"
)

func main() {

	files, err := getDirContents()

	if err != nil {
		fmt.Printf("%v", err)
	}

	getOutput(files)

	// files, err := ioutil.ReadDir(searchDir)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// output := xlsx.NewFile()
	// sheet, err := output.AddSheet("report")

	// for _, file := range files {
	// 	path, e := os.Getwd()
	// 	if e != nil {
	// 		fmt.Printf("Error %v", err)
	// 	}
	// 	csv, err := registry.ReadCSVFromBundleDirectory(path + "/tmp/" + file.Name())
	// 	if err != nil {
	// 		fmt.Printf("Error %v", err)
	// 	}

	// 	row := sheet.AddRow()
	// 	cell := row.AddCell()
	// 	cell.Value = csv.GetName()
	// 	row.AddCell().Value = csv.Kind
	// }

	// outputName := time.Now().Format("Mon-Jan2-15:04:05PST-2006")
	// if err := output.Save(outputName + ".xlsx"); err != nil {
	// 	println(err.Error())
	// }
}

func getOutput(files []os.FileInfo) error {
	output := xlsx.NewFile()
	sheet, err := output.AddSheet("report")

	// Initilize report by writing column names
	initializeReport(sheet)

	for _, file := range files {
		path, e := os.Getwd()
		if e != nil {
			return err
		}
		csv, err := registry.ReadCSVFromBundleDirectory(path + "/tmp/" + file.Name())
		if err != nil {
			return err
		}

		addValueToSheet(sheet, csv)
	}
	outputName := time.Now().Format("Mon-Jan2-15:04:05PST-2006")
	if err := output.Save(outputName + ".xlsx"); err != nil {
		println(err.Error())
	}

	return nil
}

func doSDKAnnotationsExist(csv *registry.ClusterServiceVersion) (string, string, bool) {
	annotations := csv.GetAnnotations()

	_, ok := annotations["operators.operatorframework.io/builder"]
	if ok {
		return annotations["operators.operatorframework.io/builder"], annotations["operators.operatorframework.io/project_layout"], true
	}
	return "", "", false
}

func getDirContents() ([]os.FileInfo, error) {
	files, err := ioutil.ReadDir(searchDir)
	if err != nil {
		return nil, err
	}
	return files, nil
}

func addValueToSheet(sh *xlsx.Sheet, csv *registry.ClusterServiceVersion) {
	row := sh.AddRow()
	row.AddCell().Value = csv.GetName()

	builder, layout, sdkStampsExist := doSDKAnnotationsExist(csv)
	if sdkStampsExist {
		row.AddCell().Value = "Yes"
		row.AddCell().Value = builder
		row.AddCell().Value = layout
	} else {
		row.AddCell().Value = "No"
	}
}

func initializeReport(sh *xlsx.Sheet) {
	row := sh.AddRow()
	row.AddCell().Value = "Operator name"
	row.AddCell().Value = "Do sdk lebels exist"
	row.AddCell().Value = "csv-builder"
	row.AddCell().Value = "csv-layout"
}
