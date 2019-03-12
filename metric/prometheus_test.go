package metric

import (
	"fmt"
	"testing"

	"github.com/negbie/heplify-server/config"
	"github.com/negbie/heplify-server/decoder"
)

var pmCh = make(chan *decoder.HEP, 20000)
var hepPacket = []byte{0x48, 0x45, 0x50, 0x33, 0x3, 0xa, 0x0, 0x0, 0x0, 0x1, 0x0, 0x7, 0x2, 0x0, 0x0, 0x0, 0x2, 0x0, 0x7, 0x11, 0x0, 0x0, 0x0, 0x3, 0x0, 0xa, 0xc0, 0xa8, 0xf7, 0xfa, 0x0, 0x0, 0x0, 0x4, 0x0, 0xa, 0xc0, 0xa8, 0xf5, 0xfa, 0x0, 0x0, 0x0, 0x7, 0x0, 0x8, 0x13, 0xc4, 0x0, 0x0, 0x0, 0x8, 0x0, 0x8, 0x13, 0xc4, 0x0, 0x0, 0x0, 0x9, 0x0, 0xa, 0x5a, 0xa2, 0x9b, 0x98, 0x0, 0x0, 0x0, 0xa, 0x0, 0xa, 0x0, 0x1, 0xd2, 0xf4, 0x0, 0x0, 0x0, 0xb, 0x0, 0x7, 0x1, 0x0, 0x0, 0x0, 0xc, 0x0, 0xa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xe, 0x0, 0x6, 0x0, 0x0, 0x0, 0xf, 0x2, 0xa7, 0x53, 0x49, 0x50, 0x2f, 0x32, 0x2e, 0x30, 0x20, 0x32, 0x30, 0x30, 0x20, 0x4f, 0x4b, 0xd, 0xa, 0x43, 0x61, 0x6c, 0x6c, 0x2d, 0x49, 0x44, 0x3a, 0x20, 0x42, 0x43, 0x30, 0x39, 0x39, 0x38, 0x38, 0x34, 0x40, 0x36, 0x64, 0x66, 0x63, 0x66, 0x66, 0x65, 0x38, 0xd, 0xa, 0x43, 0x53, 0x65, 0x71, 0x3a, 0x20, 0x32, 0x31, 0x35, 0x38, 0x33, 0x34, 0x34, 0x38, 0x39, 0x20, 0x4f, 0x50, 0x54, 0x49, 0x4f, 0x4e, 0x53, 0xd, 0xa, 0x46, 0x72, 0x6f, 0x6d, 0x3a, 0x20, 0x3c, 0x73, 0x69, 0x70, 0x3a, 0x31, 0x39, 0x32, 0x2e, 0x31, 0x36, 0x38, 0x2e, 0x31, 0x31, 0x31, 0x2e, 0x31, 0x31, 0x31, 0x3a, 0x35, 0x30, 0x36, 0x30, 0x3e, 0x3b, 0x74, 0x61, 0x67, 0x3d, 0x36, 0x64, 0x66, 0x63, 0x66, 0x66, 0x65, 0x38, 0x2b, 0x31, 0x2b, 0x62, 0x30, 0x61, 0x39, 0x30, 0x30, 0x30, 0x33, 0x2b, 0x63, 0x39, 0x65, 0x66, 0x63, 0x32, 0x30, 0x62, 0xd, 0xa, 0x54, 0x6f, 0x3a, 0x20, 0x3c, 0x73, 0x69, 0x70, 0x3a, 0x31, 0x39, 0x32, 0x2e, 0x31, 0x36, 0x38, 0x2e, 0x31, 0x31, 0x31, 0x2e, 0x31, 0x31, 0x31, 0x3a, 0x35, 0x30, 0x36, 0x30, 0x3b, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x3d, 0x75, 0x64, 0x70, 0x3e, 0x3b, 0x74, 0x61, 0x67, 0x3d, 0x31, 0x38, 0x30, 0x34, 0x61, 0x34, 0x37, 0x64, 0x2b, 0x31, 0x2b, 0x65, 0x31, 0x30, 0x35, 0x30, 0x34, 0x37, 0x30, 0x2b, 0x62, 0x31, 0x32, 0x38, 0x61, 0x35, 0x36, 0x39, 0xd, 0xa, 0x56, 0x69, 0x61, 0x3a, 0x20, 0x53, 0x49, 0x50, 0x2f, 0x32, 0x2e, 0x30, 0x2f, 0x55, 0x44, 0x50, 0x20, 0x31, 0x39, 0x32, 0x2e, 0x31, 0x36, 0x38, 0x2e, 0x31, 0x31, 0x31, 0x2e, 0x31, 0x31, 0x31, 0x3a, 0x35, 0x30, 0x36, 0x30, 0x3b, 0x62, 0x72, 0x61, 0x6e, 0x63, 0x68, 0x3d, 0x7a, 0x39, 0x68, 0x47, 0x34, 0x62, 0x4b, 0x2b, 0x32, 0x31, 0x66, 0x31, 0x31, 0x33, 0x65, 0x37, 0x65, 0x33, 0x64, 0x30, 0x34, 0x63, 0x38, 0x34, 0x36, 0x31, 0x34, 0x38, 0x61, 0x39, 0x61, 0x64, 0x37, 0x36, 0x30, 0x37, 0x61, 0x65, 0x66, 0x61, 0x31, 0x2b, 0x36, 0x64, 0x66, 0x63, 0x66, 0x66, 0x65, 0x38, 0x2b, 0x31, 0xd, 0xa, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x3a, 0x20, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0xd, 0xa, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x2d, 0x4c, 0x65, 0x6e, 0x67, 0x74, 0x68, 0x3a, 0x20, 0x37, 0x38, 0xd, 0xa, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x2d, 0x54, 0x79, 0x70, 0x65, 0x3a, 0x20, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x73, 0x64, 0x70, 0xd, 0xa, 0x53, 0x75, 0x70, 0x70, 0x6f, 0x72, 0x74, 0x65, 0x64, 0x3a, 0x20, 0x31, 0x30, 0x30, 0x72, 0x65, 0x6c, 0x2c, 0x20, 0x74, 0x69, 0x6d, 0x65, 0x72, 0xd, 0xa, 0x41, 0x63, 0x63, 0x65, 0x70, 0x74, 0x2d, 0x4c, 0x61, 0x6e, 0x67, 0x75, 0x61, 0x67, 0x65, 0x3a, 0x20, 0x65, 0x6e, 0xd, 0xa, 0x41, 0x63, 0x63, 0x65, 0x70, 0x74, 0x2d, 0x45, 0x6e, 0x63, 0x6f, 0x64, 0x69, 0x6e, 0x67, 0x3a, 0x20, 0x69, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0xd, 0xa, 0x41, 0x63, 0x63, 0x65, 0x70, 0x74, 0x3a, 0x20, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x73, 0x64, 0x70, 0x2c, 0x20, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x69, 0x73, 0x75, 0x70, 0x2c, 0x20, 0x6d, 0x75, 0x6c, 0x74, 0x69, 0x70, 0x61, 0x72, 0x74, 0x2f, 0x6d, 0x69, 0x78, 0x65, 0x64, 0xd, 0xa, 0x41, 0x6c, 0x6c, 0x6f, 0x77, 0x3a, 0x20, 0x49, 0x4e, 0x56, 0x49, 0x54, 0x45, 0x2c, 0x20, 0x41, 0x43, 0x4b, 0x2c, 0x20, 0x43, 0x41, 0x4e, 0x43, 0x45, 0x4c, 0x2c, 0x20, 0x42, 0x59, 0x45, 0x2c, 0x20, 0x4f, 0x50, 0x54, 0x49, 0x4f, 0x4e, 0x53, 0x2c, 0x20, 0x4e, 0x4f, 0x54, 0x49, 0x46, 0x59, 0x2c, 0x20, 0x50, 0x52, 0x41, 0x43, 0x4b, 0x2c, 0x20, 0x55, 0x50, 0x44, 0x41, 0x54, 0x45, 0x2c, 0x20, 0x49, 0x4e, 0x46, 0x4f, 0x2c, 0x20, 0x52, 0x45, 0x46, 0x45, 0x52, 0xd, 0xa, 0xd, 0xa, 0x76, 0x3d, 0x30, 0xd, 0xa, 0x6f, 0x3d, 0x2d, 0x20, 0x30, 0x20, 0x30, 0x20, 0x49, 0x4e, 0x20, 0x49, 0x50, 0x34, 0x20, 0x30, 0x2e, 0x30, 0x2e, 0x30, 0x2e, 0x30, 0xd, 0xa, 0x73, 0x3d, 0x2d, 0xd, 0xa, 0x63, 0x3d, 0x49, 0x4e, 0x20, 0x49, 0x50, 0x34, 0x20, 0x30, 0x2e, 0x30, 0x2e, 0x30, 0x2e, 0x30, 0xd, 0xa, 0x74, 0x3d, 0x30, 0x20, 0x30, 0xd, 0xa, 0x6d, 0x3d, 0x61, 0x75, 0x64, 0x69, 0x6f, 0x20, 0x30, 0x20, 0x52, 0x54, 0x50, 0x2f, 0x41, 0x56, 0x50, 0x20, 0x38}
var RTCPStat = `{"sender_information":{"ntp_timestamp_sec":3719322562,"ntp_timestamp_usec":3758534470,"rtp_timestamp":360902880,"packets":4017,"octets":642720},"ssrc":2543003035,"type":202,"report_count":1,"report_blocks":[{"source_ssrc":1393754395,"fraction_lost":0,"packets_lost":0,"highest_seq_no":29662,"ia_jitter":159,"lsr":0,"dlsr":0}],"report_blocks_xr":{"end_system_delay":11},"sdes_ssrc":2540000035}`
var XRTPStat = `CS=123;PS=131918;ES=132131;OS=21106880;SP=0/0;SO=0;QS=-;PR=132118;ER=132131;OR=21138880;CR=0;SR=0;QR=-;PL=22,33;BL=0;LS=0;RB=0/0;SB=-/-;EN=PCMA;DE=PCMA;JI=23,111;DL=34,25,70;IP=172.29.208.227:7078,172.17.66.118:21518`
var VQSessionReport = `VQSessionReport: CallTerm
						CallID:825962570309-8ds5sl3mca99
						LocalID:<sip:5004@10.0.3.252>
						RemoteID:<sip:520@10.0.3.252;user=phone>
						OrigID:<sip:5004@10.0.3.252>
						LocalAddr:IP=10.0.3.321 PORT=57460 SSRC=0x014EA261
						LocalMAC:0004135310DB
						RemoteAddr:IP=10.0.3.123 PORT=10034 SSRC=0x1F634EA2
						DialogID:825962570309-8ds5sl3mca99;to-tag=gqj87t0stF-M8g.kPREKLthaGl030mze;from-tag=2ygtpy7bgk
						x-UserAgent:snom821/873_19_20130321
						LocalMetrics:
						Timestamps:START=2016-06-16T07:47:14Z STOP=2016-06-16T07:47:21Z
						SessionDesc:PT=8 PD=PCMA SR=8000 PPS=50 SSUP=off
						x-SIPmetrics:SVA=RG SRD=392 SFC=0
						x-SIPterm:SDC=OK SDT=7 SDR=OR
						JitterBuffer:JBA=3 JBR=2 JBN=20 JBM=20 JBX=240
						PacketLoss:NLR=5.5 JDR=0.0
						BurstGapLoss:BLD=0.0 BD=0 GLD=0.0 GD=5930 GMIN=16
						Delay:RTD=0 ESD=0 IAJ=0
						QualityEst:MOSLQ=3.8 MOSCQ=4.2`

var janusStat = `{"media":"audio","base":48000,"rtt":11,"lost":3,"lost-by-remote":4,"jitter-local":2,"jitter-remote":0,"in-link-quality":100,"in-media-link-quality":100,"out-link-quality":0,"out-media-link-quality":0,"packets-received":250,"packets-sent":250,"bytes-received":12840,"bytes-sent":5000,"bytes-received-lastsec":2611,"bytes-sent-lastsec":1020,"nacks-received":0,"nacks-sent":55}`

func init() {
	config.Setting.PromAddr = ":9999"
	config.Setting.PromTargetName = "proxy_inc_ip,proxy_out_ip"
	config.Setting.PromTargetIP = "192.168.245.250,192.168.247.250"
	go func() {
		metric := New("prometheus")
		metric.Chan = pmCh
		if err := metric.Run(); err != nil {
			fmt.Println(err)
		}
	}()
}

func BenchmarkExpose(b *testing.B) {
	hep, err := decoder.DecodeHEP(hepPacket)
	if err != nil {
		b.Error(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pmCh <- hep
	}
}

func BenchmarkDissectRTCPXRStats(b *testing.B) {
	hep, err := decoder.DecodeHEP(hepPacket)
	if err != nil {
		b.Error(err)
	}
	hep.ProtoType = 35
	hep.Payload = VQSessionReport
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pmCh <- hep
	}
}

func BenchmarkDissectRTCPStats(b *testing.B) {
	hep, err := decoder.DecodeHEP(hepPacket)
	if err != nil {
		b.Error(err)
	}
	hep.ProtoType = 5
	hep.Payload = RTCPStat
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pmCh <- hep
	}
}

func BenchmarkDissectXRTPStats(b *testing.B) {
	hep, err := decoder.DecodeHEP(hepPacket)
	if err != nil {
		b.Error(err)
	}
	hep.ProtoType = 1
	hep.SIP.RTPStatVal = XRTPStat
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pmCh <- hep
	}
}

func BenchmarkDissectJanusStats(b *testing.B) {
	hep, err := decoder.DecodeHEP(hepPacket)
	if err != nil {
		b.Error(err)
	}
	hep.ProtoType = 1032
	hep.Payload = janusStat
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pmCh <- hep
	}
}
