package agent

import (
	"fmt"
	"github.com/att-cloudnative-labs/khan/pkg/hosts"
	"github.com/florianl/go-conntrack"
	"runtime/debug"
	"sync"

	"net/http"
	"time"

	"github.com/golang/glog"
)

// https://github.com/torvalds/linux/blob/5924bbecd0267d87c24110cbe2041b5075173a25/include/net/tcp_states.h#L16

const (
	TcpProtoNum = 6
	UdpProtoNum = 17
)

// ConntrackCount count of conntrack entries with same tags
type ConntrackCount struct {
	Node            string     `json:"node"`
	Count           int        `json:"count"`
	OriginSrc       hosts.Host `json:"originSrc"`
	SourceIP        string     `json:"sourceIP"`
	OriginDst       hosts.Host `json:"originDst"`
	DestinationIP   string     `json:"destinationIP"`
	DestinationPort uint16     `json:"destinationPort"`
	ReplySrc        hosts.Host `json:"replySrc"`
	ReplyDst        hosts.Host `json:"replyDst"`

	State      string `json:"state"`
	ReplyState string `json:"repState"`
	Transport  string `json:"transport"`
}

var latestConntrackEntryMap = make(map[ConntrackCount]*ConntrackCount)

// ConntrackUpdater controller for updating conntrack counts (cache)
type ConntrackUpdater struct {
	nodeName      string
	periodSeconds int
	stopCh        chan bool
	wg            *sync.WaitGroup
}

// NewConntrackUpdater new ConntrackUpdater
func NewConntrackUpdater(nodeName string, periodSeconds int, wg *sync.WaitGroup) ConntrackUpdater {
	return ConntrackUpdater{
		nodeName:      nodeName,
		periodSeconds: periodSeconds,
		stopCh:        make(chan bool),
		wg:            wg,
	}
}

// StartConntrackUpdater start ConntrackUpdater
func (c *ConntrackUpdater) StartConntrackUpdater() {
	defer c.wg.Done()
	glog.Info("starting ConntrackUpdater")
	ticker := time.NewTicker(time.Duration(c.periodSeconds) * time.Second)
	for {
		select {
		case <-ticker.C:
			go func() {
				if err := UpdateConntrackEntries(c.nodeName); err != nil {
					glog.Error(err.Error())
				}
			}()
		case <-c.stopCh:
			glog.Info("shutting down ConntrackUpdater")
			ticker.Stop()
			return
		}
	}
}

// StopConntrackUpdater stop ConntrackUpdater
func (c *ConntrackUpdater) StopConntrackUpdater() {
	close(c.stopCh)
}

// UpdateConntrackEntries build the conntrack counts cache from conntrack table
func UpdateConntrackEntries(nodeName string) error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Recovered from conntrack update: %+v", r)
			debug.PrintStack()
		}
	}()

	fmt.Println("Updating conntrack entries")
	config := conntrack.Config{
		NetNS: 0,
	}
	nfct, err := conntrack.Open(&config)
	if err != nil {
		panic(fmt.Sprintf("nfct error: %+v", err.Error()))
	}
	cons, err := nfct.Dump(conntrack.Conntrack, conntrack.IPv4)
	if err != nil {
		panic(fmt.Sprintf("con error: %+v", err.Error()))
	}

	var currentConntrackEntryMap = make(map[ConntrackCount]*ConntrackCount)
	for _, con := range cons {
		if *con.Origin.Proto.Number != 6 && *con.Origin.Proto.Number != 17 {
			// skipping non-tcp/udp protocols as some such as IP-encapsulation (4) doesn't fit format below and would cause NPE
			continue
		}
		var state, transport string

		if *con.Origin.Proto.Number == TcpProtoNum {
			transport = "tcp"
			if con.ProtoInfo != nil && con.ProtoInfo.TCP != nil {
				state = stateToString(con)
			}
		}
		if *con.Origin.Proto.Number == UdpProtoNum {
			transport = "udp"
			if con.Status != nil && hasAssuredBit(*con.Status) {
				state = "ASSURED"
			} else {
				state = "UNREPLIED"
			}
		}
		entry := ConntrackCount{
			Node:            nodeName,
			OriginSrc:       GetHost(con.Origin.Src.String()),
			SourceIP:        con.Origin.Src.String(),
			OriginDst:       GetHost(con.Origin.Dst.String()),
			DestinationIP:   con.Origin.Dst.String(),
			DestinationPort: *con.Origin.Proto.DstPort,
			ReplySrc:        GetHost(con.Reply.Src.String()),
			ReplyDst:        GetHost(con.Reply.Dst.String()),
			State:           state,
			Transport:       transport,
			Count:           1,
		}
		//con.Status
		if currentConntrackEntryMap[entry] != nil {
			currentConntrackEntryMap[entry].Count = currentConntrackEntryMap[entry].Count + 1
		} else {
			currentConntrackEntryMap[entry] = &entry
		}
	}
	latestConntrackEntryMap = currentConntrackEntryMap
	return nil
}

// GetConnections return conntrack counts in prometheus format
func GetConnections(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	for _, entry := range latestConntrackEntryMap {
		_, err := w.Write([]byte(fmt.Sprintf("khan_connection{"+
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
			"dst_port=\"%d\","+
			"transport=\"%s\","+
			"state=\"%s\"} "+
			"%d\n",
			entry.Node,
			entry.OriginSrc.Type,
			entry.SourceIP,
			entry.OriginSrc.Namespace,
			entry.OriginSrc.App,
			entry.OriginSrc.Name,
			entry.OriginDst.Type,
			entry.DestinationIP,
			entry.OriginDst.Namespace,
			entry.OriginDst.App,
			entry.OriginDst.Name,
			entry.DestinationPort,
			entry.Transport,
			entry.State,
			entry.Count)))
		if err != nil {
			fmt.Printf("Error writing connections response: %s", err.Error())
			return
		}
	}
}

var TCPStateStrings = [...]string{"TCP_ESTABLISHED", "TCP_SYN_SENT", "TCP_SYN_RECV", "TCP_FIN_WAIT1", "TCP_FIN_WAIT2", "TCP_TIME_WAIT", "TCP_CLOSE", "TCP_CLOSE_WAIT", "TCP_LAST_ACK", "TCP_LISTEN", "TCP_CLOSING"}

func stateToString(con conntrack.Con) string {
	var state string
	if con.ProtoInfo != nil && con.ProtoInfo.TCP != nil {
		state = TCPStateStrings[*con.ProtoInfo.TCP.State]
	}
	return state
}

const IpsAssuredBit = 2

func hasAssuredBit(state uint32) bool {
	var mask uint32 = 1 << IpsAssuredBit
	return state&mask != 0
}
