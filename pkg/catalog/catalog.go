package catalog

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/xralf/fluid/capnp/fluid"
	"github.com/xralf/fluid/pkg/common"

	"capnproto.org/go/capnp/v3"
)

var (
	logger *slog.Logger
)

func init() {
	logger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
	}))
	logger.Info("Catalog says welcome!")
}

type Catalog struct {
	root   System
	reader io.Reader
	writer io.Writer
}

// Used to print a JSON version of the catalog
type CatalogNode struct {
	Id          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type System struct {
	CatalogNode
	Databases []Database `json:"databases"`
}

type Database struct {
	CatalogNode
	Schemas []Schema `json:"schemas"`
}

type Schema struct {
	CatalogNode
	Tables []Table `json:"tables"`
}

type Table struct {
	CatalogNode
	Fields []Field `json:"fields"`
}

type Field struct {
	CatalogNode
	Type        string `json:"type"`
	Description string `json:"description"`
	Usage       string `json:"usage"`
}

func Example() {
	catalog := NewCatalog(os.Stdin, os.Stdout)
	catalog.ReadJson()
	catalog.WriteJson()
}

func NewCatalog(reader io.Reader, writer io.Writer) *Catalog {
	return &Catalog{
		reader: reader,
		writer: writer,
	}
}

func (c *Catalog) ReadJson() {
	bytes, err := io.ReadAll(c.reader)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(bytes, &c.root)
	if err != nil {
		panic(err)
	}
}

func (c *Catalog) WriteJson() {
	bytes, err := json.Marshal(c.root)
	if err != nil {
		panic(err)
	}
	_, err = c.writer.Write(bytes)
	if err != nil {
		panic(err)
	}
}

func (c *Catalog) ReadCapnp() {
	msg, err := capnp.NewDecoder(c.reader).Decode()
	if err != nil {
		panic(err)
	}

	// Extract the root struct from the message.
	system, err := fluid.ReadRootSystem(msg)
	if err != nil {
		panic(err)
	}
	c.root.Id = system.Id()
	if c.root.Name, err = system.Name(); err != nil {
		panic(err)
	}

	databases, err := system.Databases()
	if err != nil {
		panic(err)
	}

	for i := range databases.Len() {
		var d Database
		d.Id = databases.At(i).Id()
		if d.Name, err = databases.At(i).Name(); err != nil {
			panic(err)
		}
		schemas, err := databases.At(i).Schemas()
		if err != nil {
			panic(err)
		}

		for j := range schemas.Len() {
			var s Schema
			s.Id = schemas.At(j).Id()
			if s.Name, err = schemas.At(j).Name(); err != nil {
				panic(err)
			}
			tables, err := schemas.At(j).Tables()
			if err != nil {
				panic(err)
			}

			for k := range tables.Len() {
				var t Table
				t.Id = tables.At(k).Id()
				if t.Name, err = tables.At(k).Name(); err != nil {
					panic(err)
				}
				fields, err := tables.At(k).Fields()
				if err != nil {
					panic(err)
				}

				for l := range fields.Len() {
					var f Field
					if f.Name, err = fields.At(l).Name(); err != nil {
						panic(err)
					}
					f.Type = fields.At(l).Type().String()
					if f.Description, err = fields.At(l).Description(); err != nil {
						panic(err)
					}
					f.Usage = fields.At(l).Usage().String()

					t.Fields = append(t.Fields, f)
				}
				s.Tables = append(s.Tables, t)
			}
			d.Schemas = append(d.Schemas, s)
		}
		c.root.Databases = append(c.root.Databases, d)
	}
}

func (c *Catalog) WriteCapnp(csvTemplateFilePath string) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		panic(err)
	}

	sys := c.root

	system, err := fluid.NewRootSystem(seg)
	if err != nil {
		panic(err)
	}
	system.SetId(sys.Id)
	system.SetName(sys.Name)
	system.SetDescription(sys.Description)

	databases, err := system.NewDatabases(int32(len(sys.Databases)))
	if err != nil {
		panic(err)
	}
	for di, d := range sys.Databases {
		database := databases.At(di)
		database.SetId(d.Id)
		database.SetName(d.Name)
		database.SetDescription(d.Description)

		var schemas capnp.StructList[fluid.Schema]
		if schemas, err = database.NewSchemas(int32(len(d.Schemas))); err != nil {
			panic(err)
		}
		for si, s := range d.Schemas {

			schema := schemas.At(si)
			schema.SetId(s.Id)
			schema.SetName(s.Name)
			schema.SetDescription(s.Description)

			var tables capnp.StructList[fluid.Table]
			if tables, err = schema.NewTables(int32(len(s.Tables))); err != nil {
				panic(err)
			}
			for ti, t := range s.Tables {
				table := tables.At(ti)
				table.SetId(t.Id)
				table.SetName(t.Name)
				table.SetDescription(t.Description)

				var fields capnp.StructList[fluid.Field]
				if fields, err = table.NewFields(int32(len(t.Fields))); err != nil {
					panic(err)
				}

				var csvFields []string
				var csvTypes []string
				csvFields = append(csvFields, "#")
				csvTypes = append(csvTypes, "#")
				for fi, f := range t.Fields {
					field := fields.At(fi)

					if field.SetName(f.Name); err != nil {
						panic(err)
					}
					if field.SetType(typeToCapnpType(f.Type)); err != nil {
						panic(err)
					}
					if field.SetDescription(f.Description); err != nil {
						panic(err)
					}
					if field.SetUsage(usageToCapnpUsage(f.Usage)); err != nil {
						panic(err)
					}

					if err = fields.Set(fi, field); err != nil {
						panic(err)
					}

					csvFields = append(csvFields, f.Name)
					csvTypes = append(csvTypes, f.Type)
				}

				if err = table.SetFields(fields); err != nil {
					panic(err)
				}

				csvTemplateFileName := sys.Name + "_" + d.Name + "_" + s.Name + "_" + t.Name + ".csv"
				WriteCsvTemplateFile(csvTemplateFilePath+"/"+csvTemplateFileName, csvFields, csvTypes)
			}
		}
	}

	// Write the message to stdout.
	if err = capnp.NewEncoder(c.writer).Encode(msg); err != nil {
		panic(err)
	}
}

func WriteCsvTemplateFile(filePath string, fieldNames []string, fieldType []string) {
	var f *os.File
	var err error
	if f, err = os.Create(filePath); err != nil {
		panic(err)
	}
	defer f.Close()
	writer := csv.NewWriter(f)
	writer.Comma = common.CsvSeparator
	writer.Write(fieldNames)
	writer.Write(fieldType)
	writer.Flush()
}

func typeToCapnpType(t string) fluid.FieldType {
	switch t {
	case "boolean":
		return fluid.FieldType_boolean
	case "float64":
		return fluid.FieldType_float64
	case "integer64":
		return fluid.FieldType_integer64
	case "text":
		return fluid.FieldType_text
	case "timestamp":
		return fluid.FieldType_text
	}
	panic(fmt.Errorf("unknown field type: %v", t))
}

func usageToCapnpUsage(u string) fluid.FieldUsage {
	switch u {
	case common.FieldUsageData:
		return fluid.FieldUsage_data
	case common.FieldUsageTime:
		return fluid.FieldUsage_time
	case common.FieldUsageGroup:
		return fluid.FieldUsage_group
	case common.FieldUsageSequence:
		return fluid.FieldUsage_sequence
	}
	panic(fmt.Errorf("unknown usage: %v", u))
}

// func (c catalog) findTable(msg *capnp.Message, fullTableName string) (table fluid.Table) {
func FindTable(path string, fullTableName string) (msg *capnp.Message, table fluid.Table, err error) {
	parts := strings.Split(fullTableName, ".")

	systemName := parts[0]
	databaseName := parts[1]
	schemaName := parts[2]
	tableName := parts[3]

	var file *os.File
	if file, err = os.Open(path); err != nil {
		panic(err)
	}
	in := bufio.NewReader(file)
	if msg, err = capnp.NewDecoder(in).Decode(); err != nil {
		panic(err)
	}

	// Extract the root struct from the message.
	var y System
	var system fluid.System
	if system, err = fluid.ReadRootSystem(msg); err != nil {
		panic(err)
	}
	if y.Name, err = system.Name(); err != nil {
		panic(err)
	}
	if y.Name != systemName {
		err = errors.New("cannot find system name")
		return
	}

	var databases capnp.StructList[fluid.Database]
	if databases, err = system.Databases(); err != nil {
		panic(err)
	}

	for i := range databases.Len() {
		var d Database
		if d.Name, err = databases.At(i).Name(); err != nil {
			panic(err)
		}
		if d.Name != databaseName {
			continue
		}

		var schemas capnp.StructList[fluid.Schema]
		if schemas, err = databases.At(i).Schemas(); err != nil {
			panic(err)
		}

		for j := range schemas.Len() {
			var s Schema
			if s.Name, err = schemas.At(j).Name(); err != nil {
				panic(err)
			}
			if s.Name != schemaName {
				continue
			}

			var tables capnp.StructList[fluid.Table]
			if tables, err = schemas.At(j).Tables(); err != nil {
				panic(err)
			}

			for k := range tables.Len() {
				var t Table
				if t.Name, err = tables.At(k).Name(); err != nil {
					panic(err)
				}
				if t.Name == tableName {
					return msg, tables.At(k), nil
				}
			}
		}
	}

	err = errors.New("cannot find table")
	return
}

func FindField(path string, fullTableName string, fieldName string) (msg *capnp.Message, field fluid.Field, err error) {
	var table fluid.Table
	if msg, table, err = FindTable(path, fullTableName); err != nil {
		return
	}

	var fields fluid.Field_List
	if fields, err = table.Fields(); err != nil {
		return
	}

	for i := range fields.Len() {
		field = fields.At(i)
		var name string
		if name, err = field.Name(); err != nil {
			panic(err)
		}
		if name == fieldName {
			return
		}
	}
	return
}
