package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
)

type currency struct {
	Name  string
	Code  string
	Num   string
	Scale string
}

func main() {
	// Open the input file and read its contents
	data, err := readCsvFile(filepath.Join("scripts", "currency", "currency_data.csv"))
	if err != nil {
		panic(fmt.Errorf("error reading CSV file: %v", err))
	}

	// Convert the CSV records to a list of Currency objects
	currs := convertDataToCurrencies(data)

	// Generate Go code from the Currency objects using a template
	code, err := generateGoCode(filepath.Join("scripts", "currency", "currency_data.tmpl"), currs)
	if err != nil {
		panic(fmt.Errorf("error generating Go code: %v", err))
	}

	// Write the generated Go code to a file
	err = writeToFile("currency_data.go", code)
	if err != nil {
		panic(fmt.Errorf("error writing to file: %v", err))
	}
}

func readCsvFile(filename string) ([][]string, error) {
	// Open the CSV file
	in, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func() { _ = in.Close() }()

	// Read the CSV records
	reader := csv.NewReader(in)
	_, err = reader.Read() // header
	if err != nil {
		return nil, err
	}
	recs, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	return recs, nil
}

func convertDataToCurrencies(data [][]string) []currency {
	// Sort the CSV records by currency code
	less := func(i, j int) bool {
		a := data[i][1]
		b := data[j][1]
		switch a {
		case "XXX":
			return true
		case "XTS":
			return true
		}
		return a < b
	}
	sort.Slice(data, less)

	// Convert the CSV records to Currency objects
	currs := []currency{}
	for _, rec := range data {
		curr := currency{
			Name:  rec[0],
			Code:  rec[1],
			Num:   rec[2],
			Scale: rec[3],
		}
		currs = append(currs, curr)
	}
	return currs
}

func generateGoCode(filename string, currs []currency) ([]byte, error) {
	// Create a new template object from the template file
	fmap := template.FuncMap{
		"lower": strings.ToLower,
	}
	tmpl, err := template.New(filepath.Base(filename)).Funcs(fmap).ParseFiles(filename)
	if err != nil {
		return nil, err
	}

	// Execute the template
	var output bytes.Buffer
	err = tmpl.Execute(&output, currs)
	if err != nil {
		return nil, err
	}

	// Format the output as Go code
	formatted, err := format.Source(output.Bytes())
	if err != nil {
		return nil, err
	}
	return formatted, nil
}

func writeToFile(filename string, content []byte) error {
	// Write the content to a file
	out, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()
	writer := bufio.NewWriter(out)
	_, err = writer.Write(content)
	if err != nil {
		return err
	}
	err = writer.Flush()
	if err != nil {
		return err
	}
	return nil
}
