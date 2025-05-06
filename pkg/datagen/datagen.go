package datagen

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/xralf/fluid/pkg/common"
)

const (
	CatalogTypeBoolean   = "boolean"
	CatalogTypeFloat64   = "float64"
	CatalogTypeInteger64 = "integer64"
	CatalogTypeText      = "text"
	CatalogTypeTimestamp = "timestamp"

	CsvTypeBoolean   = "bool"
	CsvTypeFloat     = "float"
	CsvTypeInteger   = "int"
	CsvTypeText      = "text"
	CsvTypeTimestamp = "timestamp"

	RangeBoolean   = 1
	RangeFloat64   = 64
	RangeInteger64 = 64
	RangeText      = 1 //3
	RangeTimestamp = 1
	RangeGroup     = 2

	charset = "abcdefghijklmnopqrstuvwxyz"
)

var (
	seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	log        zerolog.Logger
)

type Catalog struct {
	Id        int        `json:"id"`
	Name      string     `json:"name"`
	Databases []Database `json:"databases"`
}

type Database struct {
	Id      int      `json:"id"`
	Name    string   `json:"name"`
	Schemas []Schema `json:"schemas"`
}

type Schema struct {
	Id     int     `json:"id"`
	Name   string  `json:"name"`
	Tables []Table `json:"tables"`
}

type Table struct {
	Id     int     `json:"id"`
	Name   string  `json:"name"`
	Fields []Field `json:"fields"`
}

type Field struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Type  string `json:"type"`
	Usage string `json:"usage"`
}

const (
	zipfS    float64 = 2.0
	zipfV    float64 = 5.0
	zipfImax uint64  = 100
)

var (
	randGenerator *rand.Rand
)

func init() {
	//zerolog.SetGlobalLevel(zerolog.Disabled)
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log = zerolog.New(os.Stderr).With().Caller().Timestamp().Logger()
	log.Info().Msg("Data Generator says welcome!")

	randGenerator = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func AsText(values []int) (result []string) {
	for _, v := range values {
		result = append(result, strconv.Itoa(v))
	}
	return
}

func CreateData(fieldNames []string, fieldUsages []string, csvTypeNames []string, maxValues []int, numRows int, data *[][]string) {
	ts := time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local) // New Year's midnight

	zipfGenerators := make([]*rand.Zipf, len(maxValues))

	for n, name := range csvTypeNames {
		if name == CsvTypeInteger {
			max := maxValues[n]
			zipfImax := uint64(max)
			zipfGenerators[n] = rand.NewZipf(randGenerator, zipfS, zipfV, zipfImax)
		}
	}

	for i := 0; i < numRows; i++ {
		var row []string
		for n, name := range csvTypeNames {
			max := maxValues[n]
			var value string
			switch name {
			case CsvTypeBoolean:
				value = strconv.FormatBool(rand.Intn(max) == 1)
			case CsvTypeFloat:
				value = fmt.Sprintf("%.2f", RandomFloat(0, max))
			case CsvTypeInteger:
				if fieldNames[n] == "one" {
					value = strconv.Itoa(1) // for experiments of counts
				} else if fieldUsages[n] == common.FieldUsageSequence {
					value = strconv.Itoa(i)
				} else if fieldUsages[n] == common.FieldUsageGroup {
					value = strconv.Itoa(100 + rand.Intn(RangeGroup))
				} else {
					//value = strconv.Itoa(rand.Intn(max))
					number := zipfGenerators[n].Uint64()
					value = strconv.Itoa(int(number))
				}
			case CsvTypeText:
				if fieldUsages[n] == common.FieldUsageGroup {
					value = "xxx" + strconv.Itoa(rand.Intn(RangeGroup))
				} else if fieldUsages[n] == common.FieldUsageTime {
					value = fmt.Sprint(ts.Add(100 * time.Millisecond * time.Duration(i)).Format(time.RFC3339Nano))
				} else {
					value = RandomString(max)
				}
			case CsvTypeTimestamp:
				value = fmt.Sprint(ts.Add(10 * time.Millisecond * time.Duration(i)).Format(time.RFC3339Nano))
			default:
				value = ""
			}
			row = append(row, value)
		}
		*data = append(*data, row)
	}
}

func WriteData(preamble string, rows [][]string) {
	fmt.Print(preamble)
	csvFile := os.Stdout
	writer := csv.NewWriter(csvFile)
	writer.Comma = common.CsvSeparator

	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			panic(err)
		}
	}
	writer.Flush()
}

func ExtractFieldInfo(catalogFileName string, dotTableName string) (fieldNames []string, fieldTypes []string, fieldUsages []string, csvTypes []string, maxValues []int) {
	table, err := findTable(catalogFileName, dotTableName)
	if err != nil {
		panic(err)
	}
	for _, field := range table.Fields {
		fieldNames = append(fieldNames, field.Name)
		fieldTypes = append(fieldTypes, field.Type)
		fieldUsages = append(fieldUsages, field.Usage)

		var csvType string
		switch field.Type {
		case CatalogTypeInteger64:
			csvType = CsvTypeInteger
		case CatalogTypeFloat64:
			csvType = CsvTypeFloat
		case CatalogTypeBoolean:
			csvType = CsvTypeBoolean
		case CatalogTypeTimestamp:
			csvType = CsvTypeTimestamp
		case CatalogTypeText:
			if field.Usage == "timestamp" {
				csvType = CsvTypeTimestamp
			} else {
				csvType = CsvTypeText
			}
		default:
			panic(fmt.Errorf("cannot find type: %s", field.Type))
		}
		csvTypes = append(csvTypes, csvType)

		maxValue := 100 // default
		if csvType == CsvTypeInteger {
			if strings.Contains(field.Name, CatalogTypeInteger64) {
				maxValue = RangeInteger64
			}
		} else if csvType == CatalogTypeText {
			maxValue = RangeText
		}
		maxValues = append(maxValues, maxValue)
	}
	return
}

func findTable(fileName string, dotTableName string) (table Table, err error) {
	jsonFile, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()

	s := strings.Split(dotTableName, ".")
	//systemName := s[0]
	databaseName := s[1]
	schemaName := s[2]
	tableName := s[3]

	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		panic(err)
	}
	var catalog Catalog
	json.Unmarshal(byteValue, &catalog)

	err = errors.New("cannot find table")

	var i int
	databases := catalog.Databases
	n := len(databases)
	for i = 0; i < n; i++ {
		if databases[i].Name == databaseName {
			break
		}
	}
	if i >= n {
		return
	}

	schemas := databases[i].Schemas
	n = len(schemas)
	for i = 0; i < n; i++ {
		if schemas[i].Name == schemaName {
			break
		}
	}
	if i >= n {
		return
	}

	tables := schemas[i].Tables
	n = len(tables)
	for i = 0; i < n; i++ {
		if tables[i].Name == tableName {
			break
		}
	}
	if i >= n {
		return
	}

	return tables[i], nil
}

func RandomFloat(min int, max int) float32 {
	return float32(min) + rand.Float32()*float32(max-min)
}

func RandomString(length int) string {
	return StringWithCharset(length, charset)
}

func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
