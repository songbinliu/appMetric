package prometheus

import (
	"encoding/json"
	"github.com/prometheus/common/model"
)

type PromeResponse struct {
	Status    string    `json:"status"`
	Data      *PromData `json:"data,omitempty"`
	ErrorType string    `json:"errorType,omitempty"`
	Error     string    `json:"error,omitempty"`
}

type PromData struct {
	ResultType string          `json:"resultType"`
	Result     json.RawMessage `json:"result"`
}

/*
type TurboLatencyMetric struct {
	DestinationIP      string `json:"destination_ip,omitempty"`
	DestinationService string `json:"destination_service,omitempty"`
	DestinationUID     string `json:"destination_uid,omitempty"`
	ResponseCode       string `json:"response_code,omitempty"`
	SourceIP           string `json:"source_ip,omitempty"`
	SourceService      string `json:"source_service,omitemtpy"`
}
*/

type TurboLatencyMetric struct {
	DestinationUID string `json:"destination_uid,omitempty"`
	ResponseCode   string `json:"response_code,omitempty"`
}

type TurboLatency struct {
	Metric *TurboLatencyMetric `json:"metric"`
	Value  model.SamplePair    `json:"value"`
}

type RequestCountMetric struct {
	DestinationUID string `json:"destination_uid,omitempty"`
	ResponseCode   string `json:"response_code,omitempty"`
}

type RequestPerSecond struct {
	Metric *RequestCountMetric `json:"metric"`
	Value  model.SamplePair    `json:"value"`
}
