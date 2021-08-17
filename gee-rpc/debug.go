package gee_rpc

import (
	"fmt"
	"html/template"
	"net/http"
)

const debugText = `<html>
	<body>
	<title>RPC Services</title>
	{{range .}}
	<hr>
	Service {{.Name}}
	<hr>
		<table>
		<th align=center>Method</th><th align=center>Calls</th>
		{{range $name, $mtype := .Method}}
			<tr>
			<td align=left font=fixed>{{$name}}({{$mtype.ArgType}}, {{$mtype.ReplayType}}) error</td>
			<td align=center>{{$mtype.NumCalls}}</td>
			</tr>
		{{end}}
		</table>
	{{end}}
	</body>
	</html>`

var debug = template.Must(template.New("RPC debug").Parse(debugText))

type debugHTTP struct {
	*Server
}

func (d debugHTTP) ServeHTTP(writer http.ResponseWriter, r *http.Request) {
	var services []debugService
	d.serviceMap.Range(func(sName, sevi interface{}) bool {
		svc := sevi.(*service)
		services = append(services, debugService{
			Name:   sName.(string),
			Method: svc.method,
		})
		return true
	})

	err := debug.Execute(writer, services)
	if err != nil {
		_, _ = fmt.Fprintln(writer, "rpc: error executing template:", err.Error())
	}
}

type debugService struct {
	Name   string
	Method map[string]*methodType
}
