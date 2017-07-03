package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/prometheus/common/log"
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
		fmt.Fprintf(os.Stderr, "USAGE\n")
		fmt.Fprintf(os.Stderr, "peds\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "FLAGS\n")
		w := tabwriter.NewWriter(os.Stderr, 0, 2, 2, ' ', 0)
		fs.VisitAll(func(f *flag.Flag) {
			fmt.Fprintf(w, "\t-%s %s\t%s\n", f.Name, f.DefValue, f.Usage)
		})
		w.Flush()
		fmt.Fprintf(os.Stderr, "\n")
	}
}

type vectorSpec struct {
	VectorTypeName string
	TypeName       string
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

func main() {
	flagset := flag.NewFlagSet("server", flag.ExitOnError)
	var (
		//		maps = flagset.String("maps", "", "Map1<int,string>;Map2<float,int>")
		//		sets = flagset.String("sets", "", "Set1<int>")
		//		imports = flagset.String("imports", "", "import1;import2")

		vectors = flagset.String("vectors", "", "Vec1<int>")
		file    = flagset.String("file", "", "path/to/file.go")
		pkg     = flagset.String("pkg", "", "package_name")
	)

	flagset.Usage = usage(flagset)
	if err := flagset.Parse(os.Args[1:]); err != nil {
		log.Fatal(err)
	}

	buf := bytes.Buffer{}
	err := renderTemplates([]templateSpec{
		{name: "pkg", template: "package {{.PackageName}}\n"}, {name: "common", template: templates.CommonTemplate}},
		map[string]string{"PackageName": *pkg}, &buf)

	if err != nil {
		log.Fatal(err)
	}

	vectorSpecs, err := parseVectorSpecs(*vectors)
	if err != nil {
		log.Fatal(err)
	}

	for _, spec := range vectorSpecs {
		err := renderTemplates([]templateSpec{
			{name: "vector", template: templates.VectorTemplate},
			{name: "slice", template: templates.SliceTemplate}},
			spec, &buf)

		if err != nil {
			log.Fatal(err)
		}
	}

	f, err := os.Create(*file)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	src := buf.Bytes()
	src, err = format.Source(src)
	if err != nil {
		log.Fatal(err)
	}

	f.Write(src)
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
