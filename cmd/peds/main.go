package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/pkg/errors"
	"github.com/tobgu/peds/internal/templates"
	"go/format"
	"io"
	"os"
	"regexp"
	"strings"
	"text/tabwriter"
	"text/template"
)

func usage(fs *flag.FlagSet) func() {
	return func() {
		os.Stderr.WriteString("USAGE\n")
		os.Stderr.WriteString("peds\n\n")
		os.Stderr.WriteString("FLAGS\n")
		w := tabwriter.NewWriter(os.Stderr, 0, 2, 2, ' ', 0)
		fs.VisitAll(func(f *flag.Flag) {
			fmt.Fprintf(w, "\t-%s %s\t%s\n", f.Name, f.DefValue, f.Usage)
		})
		w.Flush()
		os.Stderr.WriteString("\n")
	}
}

func logAndExit(err error) {
	fmt.Fprint(os.Stderr, "Error: ", err, "\n")
	os.Exit(1)
}

func main() {
	// TODO: - Documentation
	//       - Experience report
	//       - Clean up/unify naming, template generation?
	//       - Review public/private functions and types

	flagSet := flag.NewFlagSet("server", flag.ExitOnError)
	var (
		vectors = flagSet.String("vectors", "", "Vec1<int>")
		maps    = flagSet.String("maps", "", "Map1<int,string>;Map2<float,int>")
		sets    = flagSet.String("sets", "", "Set1<int>")
		file    = flagSet.String("file", "", "path/to/file.go")
		imports = flagSet.String("imports", "", "import1;import2")
		pkg     = flagSet.String("pkg", "", "package_name")
	)

	flagSet.Usage = usage(flagSet)
	if err := flagSet.Parse(os.Args[1:]); err != nil {
		logAndExit(err)
	}

	buf := &bytes.Buffer{}

	if err := renderCommon(buf, *pkg, *imports); err != nil {
		logAndExit(err)
	}

	if err := renderVectors(buf, *vectors); err != nil {
		logAndExit(err)
	}

	if err := renderMaps(buf, *maps); err != nil {
		logAndExit(err)
	}

	if err := renderSet(buf, *sets); err != nil {
		logAndExit(err)
	}

	if err := writeFile(buf, *file); err != nil {
		logAndExit(err)
	}
}

///////////////
/// Helpers ///
///////////////

func removeWhiteSpaces(s string) string {
	return strings.Join(strings.Fields(s), "")
}

type templateSpec struct {
	name     string
	template string
}

func renderTemplates(specs []templateSpec, templateData interface{}, dst io.Writer) error {
	for _, s := range specs {
		t := template.New(s.name)
		t, err := t.Parse(s.template)
		if err != nil {
			return err
		}

		err = t.Execute(dst, templateData)
		if err != nil {
			return err
		}
	}

	return nil
}

//////////////
/// Common ///
//////////////

func renderCommon(buf *bytes.Buffer, pkgName, imports string) error {
	importTemplate := `{{if .Imports}}
	import (
	{{range $imp := .Imports}}
		"{{$imp}}"
	{{end}}
	)
	{{end}}`

	imports = removeWhiteSpaces(imports)
	pkgName = removeWhiteSpaces(pkgName)
	return renderTemplates([]templateSpec{
		{name: "pkg", template: "package {{index .PackageName 0}}\n"},
		{name: "imports", template: importTemplate},
		{name: "common", template: templates.CommonTemplate}},
		map[string][]string{"PackageName": {pkgName}, "Imports": strings.Split(imports, ";")}, buf)
}

//////////////
/// Vector ///
//////////////

func renderVectors(buf *bytes.Buffer, vectors string) error {
	vectors = removeWhiteSpaces(vectors)
	if vectors == "" {
		return nil
	}

	vectorSpecs, err := parseVectorSpecs(vectors)
	if err != nil {
		return err
	}

	for _, spec := range vectorSpecs {
		err := renderTemplates([]templateSpec{
			{name: "vector", template: templates.VectorTemplate},
			{name: "slice", template: templates.SliceTemplate}},
			spec, buf)

		if err != nil {
			return err
		}
	}

	return nil
}

type vectorSpec struct {
	VectorTypeName string
	TypeName       string
}

func parseVectorSpecs(vectorDescriptor string) ([]vectorSpec, error) {
	result := make([]vectorSpec, 0)
	descriptors := strings.Split(vectorDescriptor, ";")
	r := regexp.MustCompile(`([A-Za-z0-9]+)<([A-Za-z0-9.]+)>`)
	for _, d := range descriptors {
		m := r.FindStringSubmatch(strings.TrimSpace(d))
		if len(m) != 3 {
			return nil, fmt.Errorf("Invalid vector specification: %s", d)
		}

		result = append(result, vectorSpec{VectorTypeName: m[1], TypeName: m[2]})
	}

	return result, nil
}

///////////
/// Map ///
///////////

func renderMaps(buf *bytes.Buffer, maps string) error {
	maps = removeWhiteSpaces(maps)
	if maps == "" {
		return nil
	}

	mapSpecs, err := parseMapSpecs(maps)
	if err != nil {
		return err
	}

	for _, spec := range mapSpecs {
		err := renderTemplates([]templateSpec{
			{name: "private_map_template", template: templates.PrivateMapTemplate},
			{name: "public_map_template", template: templates.PublicMapTemplate}},
			spec, buf)

		if err != nil {
			return err
		}
	}

	return nil
}

type mapSpec struct {
	MapTypeName      string
	MapItemTypeName  string
	MapKeyTypeName   string
	MapValueTypeName string
	MapKeyHashFunc   string
}

func hashFunc(typ string) string {
	for _, hashTyp := range []string{"byte", "bool", "rune", "string",
		"int8", "uint8", "int16", "uint16", "int32", "uint32", "int64", "uint64", "int", "uint", "float32", "float64"} {
		if typ == hashTyp {
			return typ + "Hash"
		}
	}

	return "interfaceHash"
}

func parseMapSpecs(mapDescriptor string) ([]mapSpec, error) {
	result := make([]mapSpec, 0)
	descriptors := strings.Split(mapDescriptor, ";")
	r := regexp.MustCompile(`([A-Za-z0-9]+)<([A-Za-z0-9.]+),([A-Za-z0-9.]+)>`)
	for _, d := range descriptors {
		m := r.FindStringSubmatch(strings.TrimSpace(d))
		if len(m) != 4 {
			return nil, fmt.Errorf("Invalid map specification: %s", d)
		}

		keyTypeName := m[2]
		result = append(result,
			mapSpec{MapTypeName: m[1], MapItemTypeName: m[1] + "Item", MapKeyTypeName: keyTypeName, MapValueTypeName: m[3], MapKeyHashFunc: hashFunc(keyTypeName)})
	}

	return result, nil
}

///////////
/// Set ///
///////////

func renderSet(buf *bytes.Buffer, sets string) error {
	sets = strings.Join(strings.Fields(sets), "")
	if sets == "" {
		return nil
	}

	setSpecs, err := parseSetSpecs(sets)
	if err != nil {
		return err
	}

	for _, spec := range setSpecs {
		err := renderTemplates([]templateSpec{
			{name: "private_map_template", template: templates.PrivateMapTemplate},
			{name: "set_template", template: templates.SetTemplate}},
			spec, buf)

		if err != nil {
			return err
		}
	}

	return err
}

type setSpec struct {
	mapSpec
	SetTypeName string
}

func parseSetSpecs(setDescriptor string) ([]setSpec, error) {
	result := make([]setSpec, 0)
	descriptors := strings.Split(setDescriptor, ";")
	r := regexp.MustCompile(`([A-Za-z0-9]+)<([A-Za-z0-9.]+)>`)
	for _, d := range descriptors {
		m := r.FindStringSubmatch(strings.TrimSpace(d))
		if len(m) != 3 {
			return nil, fmt.Errorf("Invalid set specification: %s", d)
		}

		keyTypeName := m[2]
		mapName := "private" + m[1] + "Map"
		result = append(result,
			setSpec{
				mapSpec:     mapSpec{MapTypeName: mapName, MapItemTypeName: mapName + "Item", MapKeyTypeName: keyTypeName, MapValueTypeName: "struct{}", MapKeyHashFunc: hashFunc(keyTypeName)},
				SetTypeName: m[1]})
	}

	return result, nil
}

////////////
/// File ///
////////////

func writeFile(buf *bytes.Buffer, file string) error {
	if file == "" {
		return errors.New("Output file must be specified")
	}

	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()

	// The equivalent of "go fmt" before writing content
	src := buf.Bytes()
	src, err = format.Source(src)
	if err != nil {
		return err
	}

	_, err = f.Write(src)
	return err
}
