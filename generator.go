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
			if len(s) == 0 {
				return s
			}

			r := rune(s[0])
			if !unicode.IsUpper(r) && !unicode.IsLower(r) {
				return s
			}
			return string(unicode.ToLower(r)) + s[1:]
		},
		"Title": func(s string) string {
			if len(s) == 0 {
				return s
			}

			r := rune(s[0])
			if !unicode.IsUpper(r) && !unicode.IsLower(r) {
				return s
			}
			return string(unicode.ToUpper(r)) + s[1:]
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

var mockTemplate = `package {{ .PackageName }}

type Mock{{.Name}} struct {
	T testing.TB

	{{ range $method := .Methods }}{{ $method.Name }}Mock func({{ range $arg := $method.Args }}{{$arg.Name}} {{$arg.Type}},{{ end }})({{ range $ret := $method.Returns }}{{$ret.Name}} {{$ret.Type}}, {{ end }})
	{{ $method.Name }}Calls []{{ $method.Name | Untitle }}Call

	{{ end }}
}

func NewMock{{.Name}}(t testing.TB) *Mock{{.Name}} {
	return &Mock{{.Name}}{T: t}
}

{{ $iname := .Name }}
{{ range $method := .Methods }}
type {{ $method.Name | Untitle }}Call struct {
	{{ range $i, $arg := $method.Args }}{{ $arg.Name | Title }} {{$arg.Type}}
	{{ end }}
	{{ range $i, $arg := $method.Returns }}Out{{$i}} {{$arg.Type}}
	{{ end }}
}

func (v *Mock{{$iname}}) {{ $method.Name }}({{ range $arg := $method.Args }}{{$arg.Name}} {{$arg.Type}},{{ end }})({{ range $ret := $method.Returns }}{{$ret.Name}} {{$ret.Type}}, {{ end }}) {
	if v.{{$method.Name}}Mock == nil {
		msg := "call to {{$method.Name}}, but Mock{{$method.Name}} is not set"
		if v.T == nil {
			panic(msg)
		}
		require.Fail(v.T, msg)
	}
	
	v.{{ $method.Name }}Calls = append(v.{{ $method.Name }}Calls,  {{ $method.Name | Untitle }}Call{
		{{ range $i, $arg:= $method.Args }}{{ $arg.Name | Title }}: {{$arg.Name}},
		{{ end }}
	})
	{{ if $method.Returns}}{{ range $i, $arg := $method.Returns}}out{{ $i }}{{ if IsNotLastArgument $method.Returns $i }},{{ end }}{{ end }} := {{ else }}return {{ end }}v.{{ $method.Name }}Mock({{ range $arg := $method.Args }}{{$arg.Name}}, {{ end }})
	{{ range $i, $arg := $method.Returns}}v.{{ $method.Name }}Calls[len(v.{{ $method.Name }}Calls)-1].Out{{ $i }} = out{{ $i }}{{ if IsNotLastArgument $method.Returns $i }}
	{{else}}{{end}}{{ end }}
	{{ if $method.Returns}}return {{ range $i, $arg := $method.Returns}}out{{ $i }}{{ if IsNotLastArgument $method.Returns $i }},{{ end }}{{ end }}{{ end }}
}

{{ end }}
`
