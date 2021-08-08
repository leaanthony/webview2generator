package types

import (
	"bytes"
	"github.com/leaanthony/slicer"
	"io"
	"log"
	"os"
	"strings"
	"text/template"
)

type InterfaceDeclaration struct {
	Header    *InterfaceHeader   `parser:"'[' @@ ']'"`
	Name      string             `parser:"'interface' @Ident"`
	BaseClass string             `parser:" ':' @Ident '{' "`
	Methods   []*InterfaceMethod `parser:"@@+ '}'"`

	// private
	decl         *Declaration
	InvokeMethod *InterfaceMethod
	includes     slicer.StringSlicer
}

func (d *InterfaceDeclaration) Process(decl *Declaration) error {
	d.decl = decl

	// Find Invoke method
	for _, method := range d.Methods {
		err := method.Process(d)
		if err != nil {
			return err
		}
		if string(method.Name) == "Invoke" {
			d.InvokeMethod = method
			break
		}
	}
	if len(d.Methods) == 1 && d.Methods[0] == d.InvokeMethod {
		return nil
	}
	d.includes.AddUnique(`"unsafe"`)
	d.includes.AddUnique(`"golang.org/x/sys/windows"`)
	return nil
}

func (d *InterfaceDeclaration) Generate(packageName string, w io.Writer) error {
	err := d.generateVtbl(packageName, w)
	if err != nil {
		return err
	}

	err = d.generateInvoke(w)
	if err != nil {
		return err
	}

	err = d.generateInterfaceMethods(w)
	if err != nil {
		return err
	}

	return nil
}

func (d *InterfaceDeclaration) generateVtbl(packageName string, w io.Writer) error {
	data := struct {
		PackageName     string
		Name            string
		Methods         []*InterfaceMethod
		HasInvokeMethod bool
		Includes        []string
	}{
		PackageName:     packageName,
		Name:            d.Name,
		Methods:         d.Methods,
		HasInvokeMethod: d.HasInvokeMethod(),
		Includes:        d.includes.AsSlice(),
	}
	mustTemplate("Interface Vtbl", "interfacevtbl.tmpl", &data, w)
	return nil
}

func (d *InterfaceDeclaration) generateInvoke(w io.Writer) error {
	if d.InvokeMethod == nil {
		return nil
	}
	data := struct {
		Name         string
		InvokeMethod *InterfaceMethod
		Declaration  *InterfaceDeclaration
	}{
		Declaration:  d,
		Name:         d.Name,
		InvokeMethod: d.InvokeMethod,
	}
	mustTemplate("Interface Invoke", "interfaceInvoke.tmpl", &data, w)
	return nil
}

func (d *InterfaceDeclaration) HasInvokeMethod() bool {
	return d.InvokeMethod != nil
}

func mustTemplate(templateName string, filename string, data interface{}, w io.Writer) {
	templateData, err := os.ReadFile(path("types/templates/" + filename))
	if err != nil {
		log.Fatal(err)
	}
	tmpl, err := template.New(templateName).Parse(string(templateData))
	if err != nil {
		log.Fatal(err)
	}
	err = tmpl.Execute(w, data)
	if err != nil {
		log.Fatal(err)
	}
}

func (d *InterfaceDeclaration) generateInterfaceMethods(w io.Writer) error {
	if len(d.Methods) == 1 && d.Methods[0] == d.InvokeMethod {
		return nil
	}
	for _, method := range d.Methods {
		data := struct {
			Name   string
			Method *InterfaceMethod
		}{
			Name:   d.Name,
			Method: method,
		}
		mustTemplate("Interface Methods", "interfaceMethod.tmpl", &data, w)
	}
	return nil
}

type InterfaceMethod struct {
	Prop       *Prop               `parser:"('[' @('propget'|'propput') ']')?"`
	ReturnType string              `parser:"@Ident"`
	Name       InterfaceMethodName `parser:"@Ident '('"`
	Params     []*Param            `parser:" @@* ')' ';'"`

	// private
	GoMethodName string

	GoInputs        string
	InputParamNames string

	GoReturnTypes string

	ProcessedName    string
	inputParams      []*Param
	outputParams     []*Param
	OutputParamNames string
	GoOutputs        string
	decl             *InterfaceDeclaration
}

func (m *InterfaceMethod) Process(decl *InterfaceDeclaration) error {
	m.decl = decl
	// Generate Go Method name
	goMethodName := strings.TrimPrefix(decl.Name, "ICoreWebView2")
	goMethodName = strings.TrimSuffix(goMethodName, "Handler")
	goMethodName = strings.TrimSuffix(goMethodName, "Event")
	m.GoMethodName = goMethodName

	m.ProcessedName = string(m.Name)
	if m.Prop != nil {
		m.ProcessedName = string(*m.Prop) + m.ProcessedName
	}
	m.processParams()

	return nil
}

func (m *InterfaceMethod) processParams() {
	for _, param := range m.Params {
		param.Process(m)
		if param.IsOutputParam() {
			m.outputParams = append(m.outputParams, param)
		} else {
			m.inputParams = append(m.inputParams, param)
		}
	}

	m.processInputParams()
	m.processOutputParams()
}

func (m *InterfaceMethod) processInputParams() {
	var inputs slicer.StringSlicer
	var inputParamNames slicer.StringSlicer
	for _, param := range m.inputParams {
		inputs.Add(param.Name + " " + param.GoType)
		inputParamNames.Add(param.Name)
		param.processSetup()
	}
	m.GoInputs = inputs.Join(", ")
	m.InputParamNames = inputParamNames.Join(", ")
}

func (m *InterfaceMethod) processOutputParams() {
	var outputs slicer.StringSlicer
	var outputParamNames slicer.StringSlicer
	var outputParamTypes slicer.StringSlicer
	for _, param := range m.outputParams {
		outputs.Add(param.Name + " " + param.GoType)
		outputParamNames.Add(param.Name)
		outputParamTypes.Add(param.GoType)
		param.processSetup()
	}
	// Add the mandatory error
	outputs.Add("err error")
	outputParamNames.Add("err")
	outputParamTypes.Add("error")

	m.GoOutputs = outputs.Join(", ")
	m.OutputParamNames = outputParamNames.Join(", ")
	m.GoReturnTypes = outputParamTypes.Join(", ")
}

func (m *InterfaceMethod) SetupCode() string {
	var buffer bytes.Buffer
	for _, param := range m.Params {
		param.SetupCode(&buffer)
	}
	return buffer.String()
}

func (m *InterfaceMethod) CleanupCode() string {
	var buffer bytes.Buffer
	for _, param := range m.Params {
		param.CleanupCode(&buffer)
	}
	return buffer.String()
}

func (m *InterfaceMethod) VtableCallInputs() string {
	var buffer bytes.Buffer
	for _, input := range m.Params {
		buffer.WriteString("\t\t" + input.VtableCallInput + ",\n")
	}
	return buffer.String()
}

func (m *InterfaceMethod) ErrorValues() string {
	var errorValues slicer.StringSlicer
	for _, outputParam := range m.outputParams {
		errorValues.Add(defaultErrorValue(outputParam.GoType))
	}
	errorValues.Add("err")
	return errorValues.Join(", ")
}

func (m *InterfaceMethod) SuccessValues() string {
	var successValues slicer.StringSlicer
	for _, outputParam := range m.outputParams {
		successValues.Add(outputParam.GetVariableName())
	}
	successValues.Add("nil")
	return successValues.Join(", ")
}

type InterfaceHeader struct {
	UUID *UUID `parser:"'uuid' '(' @UUID ')' ',' 'object' ',' 'pointer_default' '(' 'unique' ')'"`
}

type InterfaceMethodName string

func (m *InterfaceMethodName) Capture(values []string) error {
	if len(values) == 0 {
		return nil
	}
	result := values[0]
	if strings.HasPrefix(values[0], "add_") {
		result = "Add" + result[4:]
	}
	if strings.HasPrefix(values[0], "remove_") {
		result = "Remove" + result[7:]
	}
	*m = InterfaceMethodName(result)
	return nil
}

type Prop string

func (p *Prop) Capture(values []string) error {
	if len(values) == 0 {
		return nil
	}
	result := strings.Title(values[0][4:])
	*p = Prop(result)
	return nil
}
