package decoder

import (
	"bytes"
	"net"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/cespare/xxhash"
	"github.com/coocood/freecache"
	"github.com/negbie/heplify-server/config"
	"github.com/negbie/logp"
	"github.com/negbie/sipparser"
)

// The first 4 bytes are the string "HEP3". The next 2 bytes are the length of the
// whole message (len("HEP3") + length of all the chucks we have. The next bytes
// are all the chuncks created by makeChuncks()
// Bytes: 0 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 27 28 29 30 31......
//        +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//        | "HEP3"|len|chuncks(0x0001|0x0002|0x0003|0x0004|0x0007|0x0008|0x0009|0x000a|0x000b|......)
//        +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

var (
	hepVer = []byte("HEP3")
	webrtc = []byte("WEBRTC")
	dedup  = freecache.NewCache(20 * 1024 * 1024)
)

// HEP chuncks
const (
	Version   = 1  // Chunk 0x0001 IP protocol family (0x02=IPv4, 0x0a=IPv6)
	Protocol  = 2  // Chunk 0x0002 IP protocol ID (0x06=TCP, 0x11=UDP)
	IP4SrcIP  = 3  // Chunk 0x0003 IPv4 source address
	IP4DstIP  = 4  // Chunk 0x0004 IPv4 destination address
	IP6SrcIP  = 5  // Chunk 0x0005 IPv6 source address
	IP6DstIP  = 6  // Chunk 0x0006 IPv6 destination address
	SrcPort   = 7  // Chunk 0x0007 Protocol source port
	DstPort   = 8  // Chunk 0x0008 Protocol destination port
	Tsec      = 9  // Chunk 0x0009 Unix timestamp, seconds
	Tmsec     = 10 // Chunk 0x000a Unix timestamp, microseconds
	ProtoType = 11 // Chunk 0x000b Protocol type (DNS, LOG, RTCP, SIP)
	NodeID    = 12 // Chunk 0x000c Capture client ID
	NodePW    = 14 // Chunk 0x000e Authentication key (plain text / TLS connection)
	Payload   = 15 // Chunk 0x000f Captured packet payload
	CID       = 17 // Chunk 0x0011 Correlation ID
	Vlan      = 18 // Chunk 0x0012 VLAN
)

// HEP represents HEP packet
type HEP struct {
	Version   uint32 `protobuf:"varint,1,req,name=Version" json:"Version"`
	Protocol  uint32 `protobuf:"varint,2,req,name=Protocol" json:"Protocol"`
	SrcIP     string `protobuf:"bytes,3,req,name=SrcIP" json:"SrcIP"`
	DstIP     string `protobuf:"bytes,4,req,name=DstIP" json:"DstIP"`
	SrcPort   uint32 `protobuf:"varint,5,req,name=SrcPort" json:"SrcPort"`
	DstPort   uint32 `protobuf:"varint,6,req,name=DstPort" json:"DstPort"`
	Tsec      uint32 `protobuf:"varint,7,req,name=Tsec" json:"Tsec"`
	Tmsec     uint32 `protobuf:"varint,8,req,name=Tmsec" json:"Tmsec"`
	ProtoType uint32 `protobuf:"varint,9,req,name=ProtoType" json:"ProtoType"`
	NodeID    uint32 `protobuf:"varint,10,req,name=NodeID" json:"NodeID"`
	NodePW    string `protobuf:"bytes,11,req,name=NodePW" json:"NodePW"`
	Payload   string `protobuf:"bytes,12,req,name=Payload" json:"Payload"`
	CID       string `protobuf:"bytes,13,req,name=CID" json:"CID"`
	Vlan      uint32 `protobuf:"varint,14,req,name=Vlan" json:"Vlan"`
	NetSrcIP  net.IP
	NetDstIP  net.IP
	Timestamp time.Time
	SIP       *sipparser.SipMsg
	Host      string
	Tag       string
	Node      string
}

// DecodeHEP returns a parsed HEP message
func DecodeHEP(packet []byte) (*HEP, error) {
	hep := &HEP{}
	err := hep.parse(packet)
	if err != nil {
		return nil, err
	}
	return hep, nil
}

func (h *HEP) parse(packet []byte) error {
	var err error
	if bytes.HasPrefix(packet, hepVer) {
		err = h.parseHEP(packet)
		if err != nil {
			logp.Warn("%v", err)
			return err
		}
	} else if bytes.HasSuffix(packet, webrtc) {
		if err = h.parseWebRTC(packet[:len(packet)-6]); err == nil {
			return nil
		}
		return err
	} else {
		err = h.Unmarshal(packet)
		if err != nil {
			logp.Warn("malformed packet with length %d which is neither hep nor protobuf encapsulated", len(packet))
			return err
		}
	}

	h.normPayload()
	if h.ProtoType == 0 {
		return nil
	}

	h.Timestamp = time.Unix(int64(h.Tsec), int64(h.Tmsec*1000))
	t := time.Now()
	d := t.Sub(h.Timestamp)
	if d < 0 || (h.Tsec == 0 && h.Tmsec == 0) {
		logp.Debug("heptime", "future packet timestamp: %d, now: %d, delta: %d from nodeID %d",
			h.Timestamp.UnixNano(), t.UnixNano(), d, h.NodeID)
		h.Timestamp = t
	}

	if h.ProtoType == 1 && len(h.Payload) > 32 {
		err = h.parseSIP()
		if err != nil {
			logp.Warn("%v\n%q\nnodeID: %d, protoType: %d, version: %d, protocol: %d, length: %d, flow: %s:%d->%s:%d\n\n",
				err, h.Payload, h.NodeID, h.ProtoType, h.Version, h.Protocol, len(h.Payload), h.SrcIP, h.SrcPort, h.DstIP, h.DstPort)
			return err
		}

		if len(config.Setting.DiscardMethod) > 0 {
			for k := range config.Setting.DiscardMethod {
				if config.Setting.DiscardMethod[k] == h.SIP.CseqMethod {
					h.ProtoType = 0
					return nil
				}
			}
		}
	} else if h.ProtoType == 112 && len(config.Setting.FilterHost) > 0 {
		for _, host := range config.Setting.FilterHost {
			if strings.Index(h.Payload, host) > -1 {
				h.Host = host
				return nil
			}
		}
	}
	h.Node = strconv.FormatUint(uint64(h.NodeID), 10)
	for {
		if strings.HasSuffix(h.CID, "_b2b-1") {
			h.CID = h.CID[:len(h.CID)-6]
			continue
		}
		break
	}
	logp.Debug("hep", "%+v\n\n", h)
	return nil
}

var fixUTF8 = func(r rune) rune {
	if r == utf8.RuneError || r == '\x00' {
		return -1
	}
	return r
}

func (h *HEP) normPayload() {
	if config.Setting.Dedup {
		hashVal := int64(xxhash.Sum64String(h.SrcIP)) + int64(h.SrcPort) + int64(xxhash.Sum64String(h.Payload))
		_, err := dedup.GetInt(hashVal)
		if err == nil {
			h.ProtoType = 0
			return
		}
		err = dedup.SetInt(hashVal, nil, 2)
		if err != nil {
			logp.Warn("%v", err)
		}
	}
	if !utf8.ValidString(h.Payload) {
		h.Payload = strings.Map(fixUTF8, h.Payload)
	} else if config.Setting.DBDriver == "postgres" && strings.Index(h.Payload, "\x00") > -1 {
		h.Payload = strings.Map(fixUTF8, h.Payload)
	}
}

func HEPTypeString(pktType uint32) (label string) {
	switch pktType {
	case 1:
		label = "sip"
	case 5:
		label = "rtcp"
	case 34:
		label = "rtpagent"
	case 35:
		label = "rtcpxr"
	case 38:
		label = "horaclifix"
	case 53:
		label = "dns"
	case 100:
		label = "log"
	default:
		label = strconv.Itoa(int(pktType))
	}
	return label
}
