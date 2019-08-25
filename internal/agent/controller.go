package agent

import (
	"encoding/json"
	"fmt"

	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"
)

// ConntrackEntry conntrack entry
type ConntrackEntry struct {
	ReqSrcType string `json:"req_src_type"`
	ReqSrcNs   string `json:"req_src_ns"`
	ReqSrcName string `json:"req_src_name"`
	ReqSrcApp  string `json:"req_src_app"`
	ReqSrcIp   string `json:"req_src_ip"`

	ReqDstType string `json:"req_dst_type"`
	ReqDstNs   string `json:"req_dst_ns"`
	ReqDstName string `json:"req_dst_name"`
	ReqDstApp  string `json:"req_dst_app"`
	ReqDstIp   string `json:"req_dst_ip"`

	ReqSport string `json:"req_sport"`
	ReqDport string `json:"req_dport"`

	ResSrcType string `json:"res_src_type"`
	ResSrcNs   string `json:"res_src_ns"`
	ResSrcName string `json:"res_src_name"`
	ResSrcApp  string `json:"res_src_app"`
	ResSrcIp   string `json:"res_src_ip"`

	ResDstType string `json:"res_dst_type"`
	ResDstNs   string `json:"res_dst_ns"`
	ResDstName string `json:"res_dst_name"`
	ResDstApp  string `json:"res_dst_app"`
	ResDstIp   string `json:"res_dst_ip"`

	State     string `json:"state"`
	Transport string `json:"transport"`
	RepState  string `json:"rep_state"`
	CtFlag    string `json:"ct_flag"`
}

type ConntrackCount struct {
	Node       string         `json:"node"`
	Count      string         `json:"count"`
	Connection ConntrackEntry `json:"connection"`
}

var currentConntrack []ConntrackCount

func StartController(nodeName string, conntrackDir string, periodSeconds int, stop chan struct{}) {
	ticker := time.NewTicker(time.Duration(periodSeconds) * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				go func() {
					if err := UpdateConntrackEntries(nodeName, conntrackDir); err != nil {
						glog.Error(err.Error())
					}
				}()
			case <-stop:
				ticker.Stop()
				return
			}
		}
	}()
}

func UpdateConntrackEntries(nodeName string, conntrackScript string) error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from conntrack update", r)
		}
	}()

	fmt.Println("Updating conntrack entries")
	_, err := os.Stat(conntrackScript)
	if os.IsNotExist(err) {
		panic(fmt.Errorf("conntrack script, %s, does not exist", conntrackScript))
	}
	cmd := exec.Command("/bin/sh", "-c", conntrackScript)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("error running script: %s %+v", err.Error(), output)
	}
	var entries []ConntrackCount
	if err = json.Unmarshal(output, &entries); err != nil {
		return fmt.Errorf("error parsing conntrack output: %s", err.Error())
	}
	if len(entries) == 0 {
		glog.Warningf("zero entries from conntrack. output:\n%s", output)
	}
	for i := 0; i < len(entries); i++ {
		entries[i].Node = nodeName
		// request source
		reqSrc := GetHost(entries[i].Connection.ReqSrcIp)
		entries[i].Connection.ReqSrcNs = reqSrc.Namespace
		entries[i].Connection.ReqSrcApp = reqSrc.App
		entries[i].Connection.ReqSrcName = reqSrc.Name

		// request destination
		reqDst := GetHost(entries[i].Connection.ReqDstIp)
		entries[i].Connection.ReqDstNs = reqDst.Namespace
		entries[i].Connection.ReqDstApp = reqDst.App
		entries[i].Connection.ReqDstName = reqDst.Name

		// response source
		resSrc := GetHost(entries[i].Connection.ResSrcIp)
		entries[i].Connection.ResSrcNs = resSrc.Namespace
		entries[i].Connection.ResSrcApp = resSrc.App
		entries[i].Connection.ResSrcName = resSrc.Name

		// response destination
		resDst := GetHost(entries[i].Connection.ResDstIp)
		entries[i].Connection.ResDstNs = resDst.Namespace
		entries[i].Connection.ResDstApp = resDst.App
		entries[i].Connection.ResDstName = resDst.Name

		entries[i].Connection.State = strings.TrimSpace(entries[i].Connection.State)
		entries[i].Connection.RepState = strings.TrimSpace(entries[i].Connection.RepState)
		entries[i].Connection.CtFlag = strings.TrimSpace(entries[i].Connection.CtFlag)
	}
	// don't store the last entry (script adds terminating empty entry)
	if len(entries) > 0 {
		currentConntrack = entries[:len(entries)-1]
		glog.Info("Conntrack entries successfully updated")
	} else {
		glog.Warning("Warning: conntrack script return 0 entries")
	}
	return nil
}

// GetConnections return connections in prometheus format
func GetConnections(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	for _, entry := range currentConntrack {
		count, err := strconv.Atoi(entry.Count)
		if err != nil {
			fmt.Printf("Error parsing count from connection entry: %s", err.Error())
			return
		}
		_, err = w.Write([]byte(fmt.Sprintf("khan_connection{"+
			"node=\"%s\","+
			"src_type=\"%s\","+
			"src_ip=\"%s\","+
			"src_ns=\"%s\","+
			"src_app=\"%s\","+
			"src_name=\"%s\","+
			"dst_type=\"%s\","+
			"dst_ip=\"%s\","+
			"dst_ns=\"%s\","+
			"dst_app=\"%s\","+
			"dst_name=\"%s\","+
			"dst_port=\"%s\","+
			"transport=\"%s\","+
			"state=\"%s\","+
			"rep_state=\"%s\"} "+
			"%d\n",
			entry.Node,
			entry.Connection.ReqSrcType,
			entry.Connection.ReqSrcIp,
			entry.Connection.ReqSrcNs,
			entry.Connection.ReqSrcApp,
			entry.Connection.ReqSrcName,
			entry.Connection.ReqDstType,
			entry.Connection.ReqDstIp,
			entry.Connection.ReqDstNs,
			entry.Connection.ReqDstApp,
			entry.Connection.ReqDstName,
			entry.Connection.ReqDport,
			entry.Connection.Transport,
			entry.Connection.State,
			entry.Connection.RepState,
			count)))
		if err != nil {
			fmt.Printf("Error writing connections response: %s", err.Error())
			return
		}
	}
}
