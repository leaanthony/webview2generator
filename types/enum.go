package types

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"text/template"
)

type EnumDeclaration struct {
	Name   string       `parser:"'typedef' 'enum' @Ident"`
	Values []*EnumValue `parser:"'{' (@@)+ '}' Ident ';'"`

	// private
	decl *Declaration
}

func (d *EnumDeclaration) Process(decl *Declaration) error {
	d.decl = decl
	for index, value := range d.Values {
		if value.Value == "" {
			value.Value = strconv.Itoa(index)
		}
	}
	decl.library.enums.Add(d.Name)
	return nil
}

func path(p string) string {
	p = filepath.FromSlash(p)
	res, err := filepath.Abs(p)
	if err != nil {
		log.Fatal(err)
	}
	return res
}

func (d *EnumDeclaration) Generate(packageName string, w io.Writer) error {
	data := struct {
		PackageName string
		Name        string
		Values      []*EnumValue
	}{
		PackageName: packageName,
		Name:        d.Name,
		Values:      d.Values,
	}
	templateData, err := os.ReadFile(path("types/templates/enum.tmpl"))
	if err != nil {
		return err
	}
	tmpl, err := template.New("Enum").Parse(string(templateData))
	if err != nil {
		log.Fatalln(err)
	}
	return tmpl.Execute(w, &data)
}

type EnumValue struct {
	Key   string `parser:"@Ident"`
	Value string `parser:"('=' @Hex)? ','?"`
}
