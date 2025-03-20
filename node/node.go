package node

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/utsab818/my-orchestrator/stats"
	"github.com/utsab818/my-orchestrator/utils"
)

type Node struct {
	Name            string
	Ip              string
	Api             string
	Cores           int
	Memory          int
	MemoryAllocated int
	Disk            int
	DiskAllocated   int
	Role            string
	TaskCount       int
	Stats           stats.Stats
}

func NewNode(name string, api string, role string) *Node {
	return &Node{
		Name: name,
		Api:  api,
		Role: role,
	}
}

func (n *Node) GetStats() (*stats.Stats, error) {
	var resp *http.Response
	var err error

	url := fmt.Sprintf("%s/stats", n.Api)
	resp, err = utils.HTTPWithRetry(http.Get, url)
	if err != nil {
		msg := fmt.Sprintf("Unable to connect to %v. Permanent failure. \n", n.Api)
		log.Println(msg)
		return nil, errors.New(msg)
	}

	if resp.StatusCode != 200 {
		msg := fmt.Sprintf("Error retrieving stats from %v: %v", n.Api, err)
		log.Println(msg)
		return nil, errors.New(msg)
	}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var stats stats.Stats
	err = json.Unmarshal(body, &stats)
	if err != nil {
		msg := fmt.Sprintf("error decoding essage while getting stats for node %s", n.Name)
		log.Println(msg)
		return nil, errors.New(msg)
	}

	// Validate critical fields
	if stats.MemStats == nil {
		msg := fmt.Sprintf("Node %s stats missing MemStats: %s", n.Name, string(body))
		log.Println(msg)
		return nil, errors.New(msg)
	}
	if stats.CpuStats == nil { // Required for calculateCpuUsage
		msg := fmt.Sprintf("Node %s stats missing CpuStats: %s", n.Name, string(body))
		log.Println(msg)
		return nil, errors.New(msg)
	}

	n.Memory = int(stats.MemTotalKb())
	n.Disk = int(stats.DiskTotal())

	n.Stats = stats
	return &n.Stats, nil
}
