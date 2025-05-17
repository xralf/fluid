package codegen

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"capnproto.org/go/capnp/v3"
	"github.com/xralf/fluid/capnp/fluid"
	"github.com/xralf/fluid/pkg/utility"
)

const (
	CapnpCodeFilePath    = "capnp/data/data.capnp"
	GoCodeFilePath       = "pkg/_out/functions/functions.go"
	GoCodeVariablePrefix = "p"
)

type CapnpCode struct {
	Body string
}

type GoCode struct {
	ExprStack       []GoExpression // Golang snippets to generate a single expression
	Definitions     []string       // Golang code literals definitions
	VariableCounter int

	IngressFilter   goCodeItem
	AggregateFilter goCodeItem
	ProjectFilter   goCodeItem
	SessionOpen     goCodeItem
	SessionClose    goCodeItem
}

type GoExpression struct {
	Code string
	Kind Kind
}

type Kind int

const (
	Boolean Kind = iota
	Duration
	Float
	Integer
	String
	Timestamp
	Variable
)

type goCodeItem struct {
	Imports     []string
	Types       []string
	Functions   []string
	Condition   string   // The actual expression, like "x.a >= y.b"
	Definitions []string // Any declarations needed for the Condition
}

func GoCodeCreateFile(code GoCode) {
	var imports []string
	imports = append(imports, goDefaultImports())
	imports = append(imports, code.IngressFilter.Imports...)
	imports = append(imports, code.AggregateFilter.Imports...)
	imports = append(imports, code.ProjectFilter.Imports...)
	imports = append(imports, code.SessionOpen.Imports...)
	imports = append(imports, code.SessionClose.Imports...)

	var types []string
	types = append(types, goFilterType())
	types = append(types, code.IngressFilter.Types...)
	types = append(types, code.AggregateFilter.Types...)
	types = append(types, code.ProjectFilter.Types...)
	types = append(types, code.SessionOpen.Types...)
	types = append(types, code.SessionClose.Types...)
	types = removeDuplicates[string](types)

	var functions []string
	functions = append(functions, goInitFunction())
	functions = append(functions, code.IngressFilter.Functions...)
	functions = append(functions, code.AggregateFilter.Functions...)
	functions = append(functions, code.ProjectFilter.Functions...)
	functions = append(functions, code.SessionOpen.Functions...)
	functions = append(functions, code.SessionClose.Functions...)
	functions = removeDuplicates[string](functions)

	imports = addTimeImportIfMissing(imports, types)
	imports = addTimeImportIfMissing(imports, functions)
	imports = removeDuplicates[string](imports)

	s := goPackage()
	s += strings.Join(imports[:], "\n")
	s += strings.Join(types[:], "\n")
	s += strings.Join(functions[:], "\n")

	bytes := []byte(s)
	var err error
	if err = os.WriteFile(GoCodeFilePath, bytes, 0644); err != nil {
		panic(err)
	}
}

func addTimeImportIfMissing(imports []string, code []string) []string {
	found := false
	for _, v := range code {
		if strings.Contains(v, "time.Time") {
			found = true
			break
		}
	}
	if found {
		imports = append(imports, "import \"time\"")
	}
	return imports
}

func removeDuplicates[T comparable](values []T) (result []T) {
	allKeys := make(map[T]bool)
	for _, value := range values {
		if _, key := allKeys[value]; !key {
			allKeys[value] = true
			result = append(result, value)
		}
	}
	return result
}

func goPackage() string {
	return `
package functions
`
}

func goDefaultImports() string {
	return `
import "log/slog"
import "os"
import "github.com/xralf/fluid/capnp/data"
`
}

func goFilterType() string {
	return `
type Filterer interface {
  EvalIngressFilter(row data.IngressRow) (pass bool)
  EvalAggregateFilter(row data.AggregateRow) (pass bool)
  EvalSessionOpenFilter(row data.IngressRow) (pass bool)
  EvalSessionCloseFilter(row data.IngressRow) (pass bool)
  EvalProjectFilter(row data.EgressRow) (pass bool)
}

type Filter struct{}
`
}

func goInitFunction() string {
	return `
var (
	logger *slog.Logger
)

func init() {
	logger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
	}))
}
`
}

type FilterType int

const (
	IngressFilterType FilterType = iota
	AggregateFilterType
	ProjectFilterType
)

var (
	logger *slog.Logger
)

func Init() {
	logger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
	}))
	logger.Info("Catalog says welcome!")
}

func GoInternalPayload(nodeName string, node *fluid.Node, operatorType fluid.OperatorType, rootNode *fluid.Node) (code string) {
	code += "type Internal" + nodeName + "Payload struct {\n"

	var err error
	var fields capnp.StructList[fluid.Field]
	if fields, err = node.Fields(); err != nil {
		panic(err)
	}

	var name string
	for i := range fields.Len() {
		if name, err = fields.At(i).Name(); err != nil {
			panic(err)
		}
		field := fields.At(i)
		catalogUsage := field.Usage()
		goType := FindCatalogFieldType(rootNode, name, operatorType)
		code += GoFieldDeclaration(name, goType, catalogUsage)
	}

	code += "}\n"
	return
}

func GoUnusedInternalPayload(name string) (code string) {
	code += "// Internal" + name + "Payload is never called because the query has no " + name + " filter.\n"
	code += "type Internal" + name + "Payload struct {}\n"
	return
}

func GoTranslate(nodeName string, node *fluid.Node, root *fluid.Node) (code string) {
	var fields capnp.StructList[fluid.Field]
	var err error
	if fields, err = node.Fields(); err != nil {
		panic(err)
	}

	code += "func Translate" + nodeName + "Payload(in data." + nodeName + "Payload) (out Internal" + nodeName + "Payload) {\n"

	var name string
	for i := range fields.Len() {
		if name, err = fields.At(i).Name(); err != nil {
			panic(err)
		}
		field := fields.At(i)
		catalogUsage := field.Usage()
		goType := FindCatalogFieldType(root, name, node.Type())
		code += GoFieldMapping(name, goType, catalogUsage) + "\n"
	}

	code += "return\n"
	code += "}\n"

	return
}

func GoFilter(filterName string, payloadName string) (code string) {
	code += "func (f *Filter) Eval" + filterName + "Filter(row data." + payloadName + "Row) (pass bool) {\n"
	code += "var err error\n"
	code += "var payload data." + payloadName + "Payload\n"
	code += "if payload, err = row.Payload(); err != nil {\n"
	code += "panic(err)\n"
	code += "}\n"
	code += "internalPayload := Translate" + payloadName + "Payload(payload)\n"
	code += "pass = eval" + filterName + "Filter(internalPayload)\n"
	code += "return\n"
	code += "}\n"
	return
}

func GoEval(filterName string, payloadName string, list []string, body string) (code string) {
	var header string
	foundErr := false
	for _, v := range list {
		if strings.Contains(v, "err != nil") {
			foundErr = true
		}
		header += v
	}
	if foundErr {
		header = "var err error\n" + header
	}

	code = "\nfunc eval" + filterName + "Filter(" + GoCodeVariablePrefix + " Internal" + payloadName + "Payload) (pass bool) {\n"
	code += header
	code += "pass = " + body + "\n"
	code += "return\n"
	code += "}\n"

	return
}

func GoPassthroughEvalFunction(filterName string, payloadName string) (code string) {
	code += "// eval" + filterName + "Filter never blocks a row in the " + filterName + " filter.\n"
	code += "func eval" + filterName + "Filter(payload Internal" + payloadName + "Payload) (pass bool) {\n"
	code += "return true\n"
	code += "}\n"
	return
}

func GoFieldDeclaration(fieldName string, fieldType fluid.FieldType, fieldUsage fluid.FieldUsage) (code string) {
	var goTypeName string
	switch fieldType {
	case fluid.FieldType_boolean:
		goTypeName = "bool"
	case fluid.FieldType_float64:
		goTypeName = "float64"
	case fluid.FieldType_integer64:
		goTypeName = "int64"
	case fluid.FieldType_text:
		if fieldUsage == fluid.FieldUsage_time {
			goTypeName = "time.Time"
		} else {
			goTypeName = "string"
		}
	default:
		panic(fmt.Errorf("cannot find field type %v", fieldType))
	}

	return fieldName + " " + goTypeName + "\n"
}

func GoFieldMapping(fieldName string, fieldType fluid.FieldType, fieldUsage fluid.FieldUsage) (code string) {
	methodName := utility.UpcaseFirstLetter(fieldName) + "()"
	switch fieldType {
	case fluid.FieldType_boolean, fluid.FieldType_float64, fluid.FieldType_integer64:
		code = "out." + fieldName + " = in." + methodName
	case fluid.FieldType_text:
		if fieldUsage == fluid.FieldUsage_time {
			code = "if value, err := in." + methodName + "; err != nil {\n"
			code += "panic(err)\n"
			code += "} else if out." + fieldName + ", err = time.Parse(time.RFC3339Nano, value); err != nil {\n"
			code += "panic(err)\n"
			code += "}"
		} else {
			code = "if value, err := in." + methodName + "; err != nil {\n"
			code += "panic(err)\n"
			code += "} else {\n"
			code += "out." + fieldName + " = value\n"
			code += "}"
		}
	default:
		panic(fmt.Errorf("cannot find field type %v", fieldType))
	}
	return
}

func CapnpCreateDataFile(code CapnpCode) {
	code.Body = CapnpDataCodePreamble() + code.Body
	bytes := []byte(code.Body)
	var err error
	if err = os.WriteFile(CapnpCodeFilePath, bytes, 0644); err != nil {
		panic(err)
	}
}

func CapnpStructGroup(rootNode *fluid.Node, fields capnp.StructList[fluid.Field], fieldNames []string) (code string) {
	code += "\nstruct Group {\n"
	var name string
	var err error
	for i, fieldName := range fieldNames {
		for j := range fields.Len() {
			if name, err = fields.At(j).Name(); err != nil {
				panic(err)
			}
			if fieldName == name {
				theType := FindCatalogFieldType(rootNode, name, fluid.OperatorType_ingress)
				code += CapnpFieldDeclaration(fieldName, i, theType, 1)
			}
		}
	}

	code += "}\n"
	return
}

func CapnpStructIngressRow(rootNode *fluid.Node, fields capnp.StructList[fluid.Field]) (code string) {
	code += "\nstruct IngressRow {\n"
	code += "\tgroup @0 :Group;\n"
	code += "\tpayload @1 :IngressPayload;\n"
	code += "}\n"
	code += "\nstruct IngressPayload {\n"
	var name string
	var err error
	for i := range fields.Len() {
		if name, err = fields.At(i).Name(); err != nil {
			panic(err)
		}
		catalogType := fields.At(i).Type().String()
		logger.Info(fmt.Sprintf("1 %v (%v)", name, catalogType))
		capnpType := FindCatalogFieldType(rootNode, name, fluid.OperatorType_ingress)
		logger.Info(fmt.Sprintf("2 %v (%v)", name, capnpType))
		code += CapnpFieldDeclaration(name, i, capnpType, 1)
	}
	code += "}\n"

	return
}

func CapnpStructAggregateRow(fields capnp.StructList[fluid.Field]) (code string) {
	code += "\nstruct AggregateRow {\n"
	code += "\tgroup @0 :Group;\n"
	code += "\tpayload @1 :AggregatePayload;\n"
	code += "}\n"
	code += "\nstruct AggregatePayload {\n"
	for i := range fields.Len() {
		field := fields.At(i)
		var name string
		var err error
		if name, err = field.Name(); err != nil {
			panic(err)
		}
		typ := field.Type()
		code += CapnpFieldDeclaration(name, i, typ, 1)
	}
	code += "}\n"
	return
}

func CapnpStructEgressRow(fields capnp.StructList[fluid.Field]) (code string) {
	code += "\nstruct EgressRow {\n"
	code += "\tgroup @0 :Group;\n"
	code += "\tpayload @1 :EgressPayload;\n"
	code += "}\n"
	code += "\nstruct EgressPayload {\n"
	for i := range fields.Len() {
		field := fields.At(i)
		var name string
		var err error
		if name, err = field.Name(); err != nil {
			panic(err)
		}
		typ := field.Type()
		code += CapnpFieldDeclaration(name, i, typ, 1)
	}
	code += "}\n"

	return
}

func CapnpDataCodePreamble() (code string) {
	code = fmt.Sprintf("using Go = import \"/go.capnp\";\n%s;\n$Go.package(\"data\");\n$Go.import(\"github.com/xralf/fluid/capnp/data\");\n", utility.CreateCapnpId())
	return
}

func CapnpFieldDeclaration(fieldName string, index int, fieldType fluid.FieldType, indent int) string {
	var typ string
	switch fieldType {
	case fluid.FieldType_boolean:
		typ = "Bool"
	case fluid.FieldType_float64:
		typ = "Float64"
	case fluid.FieldType_integer64:
		typ = "Int64"
	case fluid.FieldType_text:
		typ = "Text"
	default:
		panic(errors.New("cannot find field type"))
	}

	indentation := ""
	for range indent {
		indentation += "\t"
	}

	return indentation + fieldName + " @" + strconv.Itoa(index) + " :" + typ + ";\n"
}

func FindCatalogFieldType(rootNode *fluid.Node, name string, operatorType fluid.OperatorType) (typ fluid.FieldType) {
	logger.Info(fmt.Sprintf("FindCatalogFieldType: name: %v\n", name))

	// Find the ingressNode node that has the information about all fields
	var ingressNode *fluid.Node
	var ok bool
	if ingressNode, ok = utility.FindFirstNodeByType(rootNode, operatorType); !ok {
		panic(errors.New("cannot find field type"))
	}

	// Go through the list of fields
	var fields capnp.StructList[fluid.Field]
	var err error
	if fields, err = ingressNode.Fields(); err != nil {
		panic(err)
	}
	for i := range fields.Len() {
		field := fields.At(i)
		var fieldName string
		if fieldName, err = field.Name(); err != nil {
			panic(err)
		}
		if fieldName == name {
			return field.Type()
		}
	}
	panic(errors.New("cannot find desired field"))
}
