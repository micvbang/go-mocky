package mocky

import (
	"bytes"
	"fmt"
	"io"
	"text/template"
	"unicode"

	"golang.org/x/tools/imports"
)

type Interface struct {
	PackageName string
	Name        string
	Methods     []Method
}

type Method struct {
	Name    string
	Args    []Argument
	Returns []Argument
}

type Argument struct {
	Name string
	Type string
}

func Generate(wtr io.Writer, c Interface) error {
	funcs := template.FuncMap{
		"IsNotLastArgument": func(vs []Argument, i int) bool {
			return i < len(vs)-1
		},
		"Untitle": func(s string) string {
			return titleCasing(s, unicode.ToLower)
		},
		"Title": func(s string) string {
			return titleCasing(s, unicode.ToUpper)
		},
	}
	tmpl, err := template.New("mocky").Funcs(funcs).Parse(mockTemplate)
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(nil)
	err = tmpl.Execute(buf, c)
	if err != nil {
		return err
	}

	bs, err := imports.Process("", buf.Bytes(), nil)
	if err != nil {
		return fmt.Errorf("failed to format generated code: %w", err)
	}

	tot := 0
	for tot < len(bs) {
		n, err := wtr.Write(bs)
		if err != nil {
			return err
		}
		tot += n
	}
	return nil
}

// titleCasing uses runeCaser to set the casing of the first letter of s.
func titleCasing(s string, runeCaser func(rune) rune) string {
	if len(s) == 0 {
		return s
	}

	r := rune(s[0])
	if !unicode.IsUpper(r) && !unicode.IsLower(r) {
		return s
	}
	return string(runeCaser(r)) + s[1:]
}

var mockTemplate = `package {{ .PackageName }}

{{ $iname := .Name }}
type Mock{{.Name}} struct {
	{{ range $method := .Methods }}{{ $method.Name }}Mock func({{ range $arg := $method.Args }}{{$arg.Name}} {{$arg.Type}},{{ end }})({{ range $ret := $method.Returns }}{{$ret.Name}} {{$ret.Type}}, {{ end }})
	{{ $method.Name }}Calls []{{ $iname | Untitle }}{{ $method.Name }}Call

	{{ end }}
}

{{ range $method := .Methods }}
type {{ $iname | Untitle }}{{ $method.Name }}Call struct {
	{{ range $arg := $method.Args }}{{ $arg.Name | Title }} {{$arg.Type}}
	{{ end }}
	{{ range $i, $arg := $method.Returns }}Out{{$i}} {{$arg.Type}}
	{{ end }}
}

func (_v *Mock{{$iname}}) {{ $method.Name }}({{ range $arg := $method.Args }}{{$arg.Name}} {{$arg.Type}},{{ end }})({{ range $ret := $method.Returns }}{{$ret.Name}} {{$ret.Type}}, {{ end }}) {
	if _v.{{$method.Name}}Mock == nil {
		msg := fmt.Sprintf("call to %T.{{$method.Name}}, but Mock{{$method.Name}} is not set", _v)
		panic(msg)
	}
	
	_v.{{ $method.Name }}Calls = append(_v.{{ $method.Name }}Calls,  {{ $iname | Untitle }}{{ $method.Name }}Call{
		{{ range $i, $arg:= $method.Args }}{{ $arg.Name | Title }}: {{$arg.Name}},
		{{ end }}
	})
	{{ if $method.Returns}}{{ range $i, $arg := $method.Returns}}out{{ $i }}{{ if IsNotLastArgument $method.Returns $i }},{{ end }}{{ end }} := {{ else }}{{ end }}_v.{{ $method.Name }}Mock({{ range $arg := $method.Args }}{{$arg.Name}}, {{ end }})
	{{ range $i, $arg := $method.Returns}}_v.{{ $method.Name }}Calls[len(_v.{{ $method.Name }}Calls)-1].Out{{ $i }} = out{{ $i }}{{ if IsNotLastArgument $method.Returns $i }}
	{{else}}{{end}}{{ end }}
	{{ if $method.Returns}}return {{ range $i, $arg := $method.Returns}}out{{ $i }}{{ if IsNotLastArgument $method.Returns $i }},{{ end }}{{ end }}{{ end }}
}

{{ end }}
`
