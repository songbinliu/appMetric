package server

import (
	"bytes"
	"fmt"
	"encoding/json"
	"github.com/golang/glog"
	"html/template"
	"io"
	"net/http"

	"appMetric/pkg/util"
)

var (
	htmlHeadTemplate string = `
	<html><head><title>{{.PageTitle}}</title>
	<link rel="icon" type="image/jpg" href="data:;base64,iVBORw0KGgo">
	</head><boday><center>
	<h1>{{.PageHead}}</h1>
	<hr width="50%">
	`

	htmlFootTemplate string = `
	<hr width="50%">hostName:  {{.HostName}}
	<br/>
	hostIP: {{.HostIP}} : {{.HostPort}}
	<br/>
	ClientIP: {{.ClientIP}}
	<br/>
	OriginalClient: {{.OriginalClient}}
	</center></body></html>
	`
)

func getHead(title string, head string) (string, error) {
	tmp, err := template.New("head").Parse(htmlHeadTemplate)
	if err != nil {
		glog.Errorf("Failed to parse image template %v:%v", htmlHeadTemplate, err)
		return "", fmt.Errorf("parse failed")
	}

	var result bytes.Buffer
	data := map[string]interface{}{"PageTitle": title, "PageHead": head}
	if err := tmp.Execute(&result, data); err != nil {
		glog.Errorf("Faile to execute template: %v", err)
		return "", fmt.Errorf("execute failed.")
	}

	return result.String(), nil
}

// handle pages "/", "/index.html", "index.htm"
func (s *MetricServer) handleWelcome(path string, w http.ResponseWriter, r *http.Request) {
	head, err := getHead("Welcome", "Introduction")
	if err != nil {
		glog.Errorf("Failed to handle welcome page.")
		io.WriteString(w, "Internal Error")
		return
	}

	body := fmt.Sprintf("This is a web server to server Application latency and request-per-second.<br/> path: %s",
		path)

	foot := s.genPageFoot(r)

	io.WriteString(w, head+body+foot)
	return
}

func (s *MetricServer) genPageFoot(r *http.Request) string {
	tmp, err := template.New("foot").Parse(htmlFootTemplate)
	if err != nil {
		glog.Errorf("Failed to parse image template %v:%v", htmlFootTemplate, err)
		return ""
	}

	var result bytes.Buffer

	data := make(map[string]interface{})
	data["HostName"] = s.host
	data["HostIP"] = s.ip
	data["HostPort"] = s.port
	data["ClientIP"] = util.GetClientIP(r)
	data["OriginalClient"] = util.GetOriginalClientInfo(r)

	if err := tmp.Execute(&result, data); err != nil {
		glog.Errorf("Faile to execute template: %v", err)
		return ""
	}

	return result.String()
}

func (s *MetricServer) faviconHandler(w http.ResponseWriter, r *http.Request) {
	fpath := "/tmp/favicon.jpg"
	if !util.FileExists(fpath) {
		glog.Warningf("favicon file[%v] does not exist.", fpath)
		return
	}

	http.ServeFile(w, r, fpath)
	return
}

func (s *MetricServer) sendFailure(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadGateway)
	w.Write([]byte(`{"status":"error"}`))
	return
}

func (s *MetricServer) handlePodMetric(w http.ResponseWriter, r *http.Request) {
	mset, err := s.promClient.GetPodMetrics()
	if err != nil {
		glog.Errorf("Failed to get pod Metrics: %v", err)
		s.sendFailure(w, r)
		return
	}

	result, err := json.Marshal(mset)
	if err != nil {
		glog.Errorf("Failed to marshal json: %v", err)
		s.sendFailure(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(result)
	return
}

func (s *MetricServer) handleServiceMetric(w http.ResponseWriter, r *http.Request) {
	mset, err := s.promClient.GetServiceMetrics()
	if err != nil {
		glog.Errorf("Failed to get service Metrics: %v", err)
		s.sendFailure(w, r)
		return
	}

	result, err := json.Marshal(mset)
	if err != nil {
		glog.Errorf("Failed to marshal service json: %v", err)
		s.sendFailure(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(result)
	return
}
