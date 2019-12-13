package p2p

import (
	"encoding/hex"
	"fmt"
	mapset "github.com/deckarep/golang-set"
	"sync"
	"time"
)

type P2PManager struct {
	config *P2PManagerConfig

	peerManager *PeerManager
	lookupPeers mapset.Set // []*Peer

	selfPeerName       string
	selfPeerId         []byte // len = 16
	selfRemotePublicIP []byte // is public ?

	//selfRemoteAddr net.Addr

	changeStatusLock sync.Mutex

	waitToReplyIsPublicPeer map[*Peer]struct {
		curt time.Time
		code uint32
	}

	recordOldPublicPeerTCPAddrs    mapset.Set // old public peer addrs set[string(byte(ip_port))]
	recordStaticPublicPeerTCPAddrs mapset.Set // static setting

	// handler
	customerDataHandler MsgDataHandler
}

func NewP2PManager(cnf *P2PManagerConfig, pmcnf *PeerManagerConfig) (*P2PManager, error) {

	p2p := &P2PManager{
		config:             cnf,
		selfRemotePublicIP: nil,
		lookupPeers:        mapset.NewSet(), //make([]*Peer, 0),
		waitToReplyIsPublicPeer: make(map[*Peer]struct {
			curt time.Time
			code uint32
		}, 0),
		recordOldPublicPeerTCPAddrs:    mapset.NewSet(),
		recordStaticPublicPeerTCPAddrs: mapset.NewSet(),
		customerDataHandler:            nil,
	}

	p2p.selfPeerId = cnf.ID
	p2p.selfPeerName = cnf.Name

	// -------- TEST START --------
	//p2p.selfPeerId = make([]byte, 16)
	//rand.Read(p2p.selfPeerId) // test
	//nnn := []byte(hex.EncodeToString(p2p.selfPeerId))
	//p2p.selfPeerName = "hcx_test_peer_" + string(nnn[:8])
	//fmt.Println("im: ", p2p.selfPeerName, string(nnn))
	// -------- TEST END --------

	// pmcnf := &PeerManagerConfig{}
	p2p.peerManager = NewPeerManager(pmcnf, p2p)

	return p2p, nil
}

func (p2p *P2PManager) SetMsgHandler(handler MsgDataHandler) {
	p2p.customerDataHandler = handler
}

func (p2p *P2PManager) Start() {

	go p2p.startListenTCP()
	go p2p.startListenUDP()

	go p2p.loop()

	go p2p.peerManager.Start()

	go p2p.tryConnectToStaticBootNodes()

	fmt.Println("[Peer] Start p2p manager id:", hex.EncodeToString(p2p.selfPeerId), "name:", p2p.selfPeerName,
		"listen on port TCP:", p2p.config.TCPListenPort, "UDP:", p2p.config.UDPListenPort)

}
