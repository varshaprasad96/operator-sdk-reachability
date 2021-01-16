// package main

// import (
// 	"fmt"
// 	"io/ioutil"
// 	"os"
// 	"os/exec"
// 	"path"
// 	"path/filepath"
// 	"time"

// 	"github.com/operator-framework/api/pkg/operators"
// 	"github.com/operator-framework/operator-registry/pkg/registry"
// 	"github.com/tealeg/xlsx"
// 	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
// 	"k8s.io/apimachinery/pkg/runtime"
// 	"k8s.io/apimachinery/pkg/util/yaml"
// )

// var (
// 	builder = "operators.operatorframework.io/builder"
// 	layout  = "operators.operatorframework.io/project_layout"
// 	index   = []string{"registry.redhat.io/redhat/redhat-marketplace-index:v4.6", "quay.io/openshift-community-operators/catalog:latest",
// 		"registry.redhat.io/redhat/certified-operator-index:v4.6", "registry.redhat.io/redhat/redhat-operator-index:v4.6", "quay.io/operatorhubio/catalog:latest"}
// )

// func main() {

// 	runOpmCommand()

// 	files, err := getDirContents()

// 	if err != nil {
// 		fmt.Printf("%v", err)
// 	}

// 	err = getOutput(files)
// 	if err != nil {
// 		fmt.Printf("%v", err)
// 	}
// }

// func runOpmCommand() {
// 	var cmd *exec.Cmd

// 	// create a folder to store data
// 	cmd = exec.Command("mkdir", "tmp")

// 	err := cmd.Run()
// 	if err != nil {
// 		fmt.Printf("error creating tmp directory. Delete if already exists")
// 	}

// 	// run opm command, binary is already present in the root of the project. Package name is a placeholder.
// 	for _, indexName := range index {
// 		cmd = exec.Command("./opm", "index", "export", "-i", indexName, "-o", "api-operator", "-f", "tmp")
// 		cmd.Stdout = os.Stdout
// 		cmd.Stderr = os.Stderr
// 		err = cmd.Run()
// 		if err != nil {
// 			fmt.Printf("Error running opm command with index %s : %v", indexName, err)
// 		}
// 	}
// }

// func getSearchDir() (string, error) {
// 	pwd, err := os.Getwd()
// 	if err != nil {
// 		return "", err
// 	}

// 	wd := filepath.Join(pwd, "tmp")
// 	return wd, nil
// }

// func getDirContents() ([]os.FileInfo, error) {
// 	dir, e := getSearchDir()
// 	if e != nil {
// 		return nil, e
// 	}
// 	files, err := ioutil.ReadDir(dir)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return files, nil
// }

// func getOutput(files []os.FileInfo) error {
// 	output := xlsx.NewFile()
// 	sheet, err := output.AddSheet("report")

// 	// Initilize report by writing column names
// 	initializeReport(sheet)

// 	for _, file := range files {
// 		if file.Name() == "package.yaml" {
// 			continue
// 		}
// 		path, e := os.Getwd()
// 		if e != nil {
// 			return err
// 		}
// 		csvManifests, err := ReadCSVFromBundleDirectory(path + "/tmp/" + file.Name())
// 		if err != nil {
// 			return err
// 		}

// 		addValueToSheet(sheet, csvManifests)
// 	}

// 	defer func() {
// 		outputName := time.Now().Format("Mon-Jan2-15:04:05PST-2006")
// 		if err := output.Save("report/" + outputName + ".xlsx"); err != nil {
// 			println(err.Error())
// 		}
// 		cmd := exec.Command("rm", "-rf", "tmp")
// 		err = cmd.Run()
// 		if err != nil {
// 			fmt.Printf("error deleting tmp files")
// 		}

// 	}()

// 	return nil
// }

// func doSDKAnnotationsExist(csv *registry.ClusterServiceVersion) (string, string, bool) {
// 	annotations := csv.GetAnnotations()

// 	_, ok := annotations[builder]
// 	if ok {
// 		return annotations[builder], annotations[layout], true
// 	}
// 	return "", "", false
// }

// func addValueToSheet(sh *xlsx.Sheet, csvList *[]registry.ClusterServiceVersion) {
// 	for _, csv := range *csvList {
// 		row := sh.AddRow()
// 		row.AddCell().Value = csv.GetName()

// 		builder, layout, sdkStampsExist := doSDKAnnotationsExist(&csv)
// 		if sdkStampsExist {
// 			row.AddCell().Value = "Yes"
// 			row.AddCell().Value = csv.GetAnnotations()["createdAt"]
// 			row.AddCell().Value = builder
// 			row.AddCell().Value = layout
// 		} else {
// 			row.AddCell().Value = "No"
// 			row.AddCell().Value = csv.GetAnnotations()["createdAt"]
// 		}
// 	}
// }

// func initializeReport(sh *xlsx.Sheet) {
// 	row := sh.AddRow()
// 	row.AddCell().Value = "Operator name"
// 	row.AddCell().Value = "Do sdk labels exist"
// 	row.AddCell().Value = "Created At"
// 	row.AddCell().Value = "operator-builder"
// 	row.AddCell().Value = "operator-layout"
// }

// func ReadCSVFromBundleDirectory(bundleDir string) (*[]registry.ClusterServiceVersion, error) {
// 	dirContent, err := ioutil.ReadDir(bundleDir)
// 	if err != nil {
// 		return nil, fmt.Errorf("error reading bundle directory %s, %v", bundleDir, err)
// 	}

// 	files := []string{}
// 	for _, f := range dirContent {
// 		if !f.IsDir() {
// 			files = append(files, f.Name())
// 		}
// 	}

// 	csvList := make([]registry.ClusterServiceVersion, 0)

// 	csv := registry.ClusterServiceVersion{}
// 	for _, file := range files {
// 		yamlReader, err := os.Open(path.Join(bundleDir, file))
// 		if err != nil {
// 			continue
// 		}

// 		unstructuredCSV := unstructured.Unstructured{}

// 		decoder := yaml.NewYAMLOrJSONDecoder(yamlReader, 30)
// 		if err = decoder.Decode(&unstructuredCSV); err != nil {
// 			continue
// 		}

// 		if unstructuredCSV.GetKind() != operators.ClusterServiceVersionKind {
// 			continue
// 		}

// 		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredCSV.UnstructuredContent(),
// 			&csv); err != nil {
// 			return nil, err
// 		}

// 		csvList = append(csvList, csv)

// 	}
// 	return &csvList, nil

// }
