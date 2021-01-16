package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/tealeg/xlsx"
)

const source_redhat = "redhat"
const source_community = "community"
const source_marketplace = "marketplace"
const source_certified = "certified"
const source_operatorhub = "operatorhub"
const source_prod = "prod"

var (
	builder = "operators.operatorframework.io/builder"
	layout  = "operators.operatorframework.io/project_layout"
)

type ReportColumns struct {
	Operator           string
	Version            string
	Certified          string
	CreatedAt          string
	Company            string
	Repo               string
	OCPVersion         string
	SDKVersion         string
	OperatorType       string
	SDKVersionGithub   string
	OperatorTypeGithub string
	SourceRedhat       string
	SourceCommunity    string
	SourceMarketplace  string
	SourceCertified    string
	SourceOperatorHub  string
	SourceProd         string
	Channel            string
	DefaultChannel     string
}

type ImageSummary struct {
	ID          string            `json:"Id"`
	ParentId    string            `json:",omitempty"` // nolint
	RepoTags    []string          `json:",omitempty"`
	Created     string            `json:",omitempty"`
	Size        int64             `json:",omitempty"`
	SharedSize  int               `json:",omitempty"`
	VirtualSize int64             `json:",omitempty"`
	Labels      map[string]string `json:",omitempty"`
	Containers  int               `json:",omitempty"`
	ReadOnly    bool              `json:",omitempty"`
	Dangling    bool              `json:",omitempty"`

	// Podman extensions
	Names        []string `json:",omitempty"`
	Digest       string   `json:",omitempty"`
	Digests      []string `json:",omitempty"`
	ConfigDigest string   `json:",omitempty"`
	//	History      []string `json:",omitempty"`
}

var ReportMap map[string]ReportColumns

type Inputs struct {
	Path    string
	Source  string
	Version string
}

var InputList []Inputs

func main() {
	ReportMap = make(map[string]ReportColumns)
	args := os.Args[1:]

	if len(args) == 0 {
		fmt.Printf("path is a required argument\n")
		os.Exit(1)
	}

	InputList = make([]Inputs, 0)

	for i := 0; i < len(args); i++ {
		v := strings.Split(args[i], ":")
		inputs := Inputs{
			Path:    v[0],
			Source:  v[1],
			Version: v[2],
		}
		InputList = append(InputList, inputs)
	}

	for i := 0; i < len(InputList); i++ {
		db, err := sql.Open("sqlite3", InputList[i].Path)
		if err != nil {
			panic(err)
		}

		dump(db, InputList[i].Source, InputList[i].Version)
	}

	// printReport()

	err := getOutput()
	if err != nil {
		fmt.Printf("something wrong while writing the output")
		os.Exit(1)
	}
}

func dump(db *sql.DB, sourceDescription, ocpVersion string) {
	// execute db query
	row, err := db.Query("SELECT name, csv, bundlepath FROM operatorbundle where csv is not null  order by name")
	if err != nil {
		panic(err)
	}

	defer row.Close()

	var csvStruct v1alpha1.ClusterServiceVersion

	for row.Next() {
		var name string
		var csv string
		var bundlepath string
		var operatorType string
		var sdkVersion string

		row.Scan(&name, &csv, &bundlepath)
		err := json.Unmarshal([]byte(csv), &csvStruct)
		if err != nil {
			fmt.Printf("error unmarshalling csv %s\n", err.Error())
		}

		certified := csvStruct.ObjectMeta.Annotations["certified"]
		repo := csvStruct.ObjectMeta.Annotations["repository"]

		// get channel
		channel := "unknown"
		channel, err = getChannel(db, name)

		createdAt := csvStruct.ObjectMeta.Annotations["createdAt"]
		companyName := csvStruct.Spec.Provider.Name

		annotations := csvStruct.GetAnnotations()

		_, ok := annotations[builder]
		if ok {
			sdkVersion, operatorType = annotations[builder], annotations[layout]
		}

		// operatorType, sdkVersion := parseBundleImage(bundlepath)

		f, ok := ReportMap[name]
		if ok {
			//update the entry's source columns
			//fmt.Printf("Jeff - update an entry %s\n", name)
		} else {
			ReportMap[name] = ReportColumns{
				Operator:     name,
				Version:      csvStruct.Spec.Version.String(),
				Certified:    certified,
				CreatedAt:    createdAt,
				Company:      companyName,
				Repo:         repo,
				OCPVersion:   ocpVersion,
				SDKVersion:   sdkVersion,
				OperatorType: operatorType,
				Channel:      channel,
			}
			f = ReportMap[name]
		}

		switch sourceDescription {
		case source_redhat:
			f.SourceRedhat = "yes"
		case source_community:
			f.SourceCommunity = "yes"
		case source_marketplace:
			f.SourceMarketplace = "yes"
		case source_prod:
			f.SourceProd = "yes"
		case source_certified:
			f.SourceCertified = "yes"
		case source_operatorhub:
			f.SourceOperatorHub = "yes"
		}
		ReportMap[name] = f

	}

}

func getChannel(db *sql.DB, name string) (channel string, err error) {
	sqlString := fmt.Sprintf("SELECT c.name FROM channel c, operatorbundle o where c.head_operatorbundle_name = '%s'", name)

	row, err := db.Query(sqlString)
	if err != nil {
		panic(err)
	}

	defer row.Close()

	var channelName string
	for row.Next() {
		row.Scan(&channelName)
	}

	return channelName, nil
}

func printReport() {
	keys := make([]string, 0, len(ReportMap))
	for k := range ReportMap {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	// print the 1st row which acts as the spreadsheet header
	fmt.Println("operator|version|certified|created|company|repos|ocpversion|sdkversion|operatortype|sdkversion-github|operatortype-github|source-redhat|source-community|source-marketplace|source-certified|source-operatorhub|source-prod|channel")
	for _, k := range keys {
		f := ReportMap[k]
		fmt.Printf("%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s\n",
			f.Operator,
			f.Version,
			f.Certified,
			f.CreatedAt,
			f.Company,
			f.Repo,
			f.OCPVersion,
			f.SDKVersion,
			f.OperatorType,
			f.SDKVersionGithub,
			f.OperatorTypeGithub,
			f.SourceRedhat,
			f.SourceCommunity,
			f.SourceMarketplace,
			f.SourceCertified,
			f.SourceOperatorHub,
			f.SourceProd,
			f.Channel)
	}
}

func getOutput() error {
	output := xlsx.NewFile()
	sheet, err := output.AddSheet("overall-report")
	if err != nil {
		return fmt.Errorf("error creating xls sheet", err)
	}
	initializeReport(sheet)

	for _, value := range ReportMap {
		row := sheet.AddRow()

		// Add operator Name
		row.AddCell().Value = value.Operator

		// Add csv timestamp
		row.AddCell().Value = value.CreatedAt

		// Add name of the company
		row.AddCell().Value = value.Company

		// Add operator type
		row.AddCell().Value = value.OperatorType

		// Add sdk version
		row.AddCell().Value = value.SDKVersion

		// Add source-redhat
		row.AddCell().Value = value.SourceRedhat

		// Add source-community
		row.AddCell().Value = value.SourceCommunity

		// Add source-marketplace
		row.AddCell().Value = value.SourceCommunity

		// Add source-certified
		row.AddCell().Value = value.SourceCertified

		// Add source-operator Hub
		row.AddCell().Value = value.SourceOperatorHub

		// Add source-prod
		row.AddCell().Value = value.SourceProd
	}

	defer func() {
		outputName := time.Now().Format("Mon-Jan2-15:04:05PST-2006")
		if err := output.Save("report/" + outputName + ".xlsx"); err != nil {
			fmt.Printf("error whilesaving report")
		}
	}()

	return nil
}

func initializeReport(sh *xlsx.Sheet) {
	row := sh.AddRow()
	row.AddCell().Value = "Operator name"
	row.AddCell().Value = "CreatedAt - timestamp"
	row.AddCell().Value = "Company"
	row.AddCell().Value = "Operator type"
	row.AddCell().Value = "Sdk Version"
	row.AddCell().Value = "source-redhat"
	row.AddCell().Value = "source-community"
	row.AddCell().Value = "source-marketplace"
	row.AddCell().Value = "source-certified"
	row.AddCell().Value = "source-operatorhub"
	row.AddCell().Value = "source-prod"
}
