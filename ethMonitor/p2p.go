package ethMonitor

import (
	"fmt"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

func (m *Monitor) addPeer(url string) (bool, error) {
	// Make sure the server is running, fail otherwise
	server := m.node.Server()
	if server == nil {
		return false, fmt.Errorf("node stopped")
	}
	node, err := enode.Parse(enode.ValidSchemes, url)
	if err != nil {
		return false, fmt.Errorf("invalid enode: %v", err)
	}
	server.AddPeer(node)
	server.AddTrustedPeer(node)
	return true, nil
}

func (m *Monitor) removePeer(url string) (bool, error) {
	// Make sure the server is running, fail otherwise
	server := m.node.Server()
	if server == nil {
		return false, fmt.Errorf("node stopped")
	}
	node, err := enode.Parse(enode.ValidSchemes, url)
	if err != nil {
		return false, fmt.Errorf("invalid enode: %v", err)
	}
	server.RemoveTrustedPeer(node)
	server.RemovePeer(node)
	return true, nil
}
