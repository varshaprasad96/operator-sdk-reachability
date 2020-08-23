package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/operator-framework/api/pkg/operators"
	"github.com/operator-framework/operator-registry/pkg/registry"
	"github.com/tealeg/xlsx"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
)

var (
	searchDir = "/Users/vnarsing/go/src/github.com/varshaprasad96/operator-sdk-rechability/tmp"
	builder   = "operators.operatorframework.io/builder"
	layout    = "operators.operatorframework.io/project_layout"
)

func main() {

	files, err := getDirContents()

	if err != nil {
		fmt.Printf("%v", err)
	}

	err = getOutput(files)
	if err != nil {
		fmt.Printf("%v", err)
	}
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
		csvManifests, err := ReadCSVFromBundleDirectory(path + "/tmp/" + file.Name())
		if err != nil {
			return err
		}

		addValueToSheet(sheet, csvManifests)
	}

	defer func() {
		outputName := time.Now().Format("Mon-Jan2-15:04:05PST-2006")
		if err := output.Save(outputName + ".xlsx"); err != nil {
			println(err.Error())
		}
	}()

	return nil
}

func doSDKAnnotationsExist(csv *registry.ClusterServiceVersion) (string, string, bool) {
	annotations := csv.GetAnnotations()

	_, ok := annotations[builder]
	if ok {
		return annotations[builder], annotations[layout], true
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

func addValueToSheet(sh *xlsx.Sheet, csvList *[]registry.ClusterServiceVersion) {
	for _, csv := range *csvList {
		row := sh.AddRow()
		row.AddCell().Value = csv.GetName()

		builder, layout, sdkStampsExist := doSDKAnnotationsExist(&csv)
		if sdkStampsExist {
			row.AddCell().Value = "Yes"
			row.AddCell().Value = builder
			row.AddCell().Value = layout
		} else {
			row.AddCell().Value = "No"
		}
	}
}

func initializeReport(sh *xlsx.Sheet) {
	row := sh.AddRow()
	row.AddCell().Value = "Operator name"
	row.AddCell().Value = "Do sdk lebels exist"
	row.AddCell().Value = "operator-builder"
	row.AddCell().Value = "operator-layout"
}

func ReadCSVFromBundleDirectory(bundleDir string) (*[]registry.ClusterServiceVersion, error) {
	dirContent, err := ioutil.ReadDir(bundleDir)
	if err != nil {
		return nil, fmt.Errorf("error reading bundle directory %s, %v", bundleDir, err)
	}

	files := []string{}
	for _, f := range dirContent {
		if !f.IsDir() {
			files = append(files, f.Name())
		}
	}

	csvList := make([]registry.ClusterServiceVersion, 0)

	csv := registry.ClusterServiceVersion{}
	for _, file := range files {
		yamlReader, err := os.Open(path.Join(bundleDir, file))
		if err != nil {
			continue
		}

		unstructuredCSV := unstructured.Unstructured{}

		decoder := yaml.NewYAMLOrJSONDecoder(yamlReader, 30)
		if err = decoder.Decode(&unstructuredCSV); err != nil {
			continue
		}

		if unstructuredCSV.GetKind() != operators.ClusterServiceVersionKind {
			continue
		}

		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredCSV.UnstructuredContent(),
			&csv); err != nil {
			return nil, err
		}

		csvList = append(csvList, csv)

	}
	return &csvList, nil

}
