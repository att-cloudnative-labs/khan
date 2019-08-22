package conntrack

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"egbitbucket.dtvops.net/com/agent/internal/agent/appmapping"
)

type ConntrackEntry struct {
	ReqSrcNs  string `json:"req_src_ns"`
	ReqSrcApp string `json:"req_src_app"`
	ReqSrcPod string `json:"req_src_pod"`
	ReqSrcIp  string `json:"req_src_ip"`

	ReqDstNs  string `json:"req_dst_ns"`
	ReqDstApp string `json:"req_dst_app"`
	ReqDstPod string `json:"req_dst_pod"`
	ReqDstIp  string `json:"req_dst_ip"`

	ReqSport string `json:"req_sport"`
	ReqDport string `json:"req_dport"`

	ResSrcNs  string `json:"res_src_ns"`
	ResSrcApp string `json:"res_src_app"`
	ResSrcPod string `json:"res_src_pod"`
	ResSrcIp  string `json:"res_src_ip"`

	ResDstNs  string `json:"res_dst_ns"`
	ResDstApp string `json:"res_dst_app"`
	ResDstPod string `json:"res_dst_pod"`
	ResDstIp  string `json:"res_dst_ip"`

	State    string `json:"state"`
	Transport string `json:"transport"`
	RepState string `json:"rep_state"`
	CtFlag   string `json:"ct_flag"`
}

type ConntrackCount struct {
	Node string `json:"node"`
	Count string `json:"count"`
	Connection ConntrackEntry `json:"connection"`
}

var currentConntrack []ConntrackCount

func StartUpdateTimer(nodeName string, conntrackDir string, periodSeconds int, stop chan struct{}) {
	ticker := time.NewTicker(time.Duration(periodSeconds) * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				go UpdateConntrackEntries(nodeName, conntrackDir)
			case <-stop:
				ticker.Stop()
				return
			}
		}
	}()
}

func UpdateConntrackEntries(nodeName string, conntrackScript string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from conntrack update", r)
		}
	}()

	fmt.Println("Updating conntrack entries")
	_, err := os.Stat(conntrackScript)
	if os.IsNotExist(err) {
		err = fmt.Errorf("conntrack script, %s, does not exist", conntrackScript)
		panic(err)
	}
	cmd := exec.Command("/bin/sh", "-c", conntrackScript)
	output, err := cmd.Output()
	if err != nil {
		log.Fatal(fmt.Errorf("error running script: %s %+v", err.Error(), output))
	}
	var entries []ConntrackCount
	json.Unmarshal(output, &entries)
	if len(entries) == 0 {
		fmt.Printf("0 entries from conntrack unmarshalled. Output:\n%s", output)
	}
	for i := 0; i < len(entries); i++ {
		entries[i].Node = nodeName
		// request source
		reqSrc := appmapping.Get(entries[i].Connection.ReqSrcIp)
		entries[i].Connection.ReqSrcNs = reqSrc.Namespace
		entries[i].Connection.ReqSrcApp = reqSrc.AppName
		entries[i].Connection.ReqSrcPod = reqSrc.PodName

		// request destination
		reqDst := appmapping.Get(entries[i].Connection.ReqDstIp)
		entries[i].Connection.ReqDstNs = reqDst.Namespace
		entries[i].Connection.ReqDstApp = reqDst.AppName
		entries[i].Connection.ReqDstPod = reqDst.PodName

		// response source
		resSrc := appmapping.Get(entries[i].Connection.ResSrcIp)
		entries[i].Connection.ResSrcNs = resSrc.Namespace
		entries[i].Connection.ResSrcApp = resSrc.AppName
		entries[i].Connection.ResSrcPod = resSrc.PodName

		// response destination
		resDst := appmapping.Get(entries[i].Connection.ResDstIp)
		entries[i].Connection.ResDstNs = resDst.Namespace
		entries[i].Connection.ResDstApp = resDst.AppName
		entries[i].Connection.ResDstPod = resDst.PodName

		entries[i].Connection.State = strings.TrimSpace(entries[i].Connection.State)
		entries[i].Connection.RepState = strings.TrimSpace(entries[i].Connection.RepState)
		entries[i].Connection.CtFlag = strings.TrimSpace(entries[i].Connection.CtFlag)
	}
	// don't store the last entry (script adds terminating empty entry)
	if len(entries) > 0 {
		currentConntrack = entries[:len(entries)-1]
		fmt.Println("Conntrack entries succesfully updated")
	} else {
		fmt.Println("Warning: conntrack script return 0 entries")
	}
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
		_, err = w.Write([]byte(fmt.Sprintf("khan_connection{node=\"%s\",src_ip=\"%s\",src_ns=\"%s\",src_app=\"%s\",src_pod=\"%s\",dst_ip=\"%s\",dst_ns=\"%s\",dst_app=\"%s\",dst_pod=\"%s\",dst_port=\"%s\",transport=\"%s\",state=\"%s\",rep_state=\"%s\"} %d\n", entry.Node, entry.Connection.ReqSrcIp, entry.Connection.ReqSrcNs, entry.Connection.ReqSrcApp, entry.Connection.ReqSrcPod, entry.Connection.ReqDstIp, entry.Connection.ReqDstNs, entry.Connection.ReqDstApp, entry.Connection.ReqDstPod, entry.Connection.ReqDport, entry.Connection.Transport, entry.Connection.State, entry.Connection.RepState, count)))
		if err != nil {
			fmt.Printf("Error writing connections response: %s", err.Error())
			return
		}
	}
}
