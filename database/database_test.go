package database

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/valyala/bytebufferpool"
	"glaye/heplify-server/config"
	"glaye/heplify-server/decoder"
)

var dbCh = make(chan *decoder.HEP, 20000)
var hepPacket = []byte{0x48, 0x45, 0x50, 0x33, 0x3, 0xa, 0x0, 0x0, 0x0, 0x1, 0x0, 0x7, 0x2, 0x0, 0x0, 0x0, 0x2, 0x0, 0x7, 0x11, 0x0, 0x0, 0x0, 0x3, 0x0, 0xa, 0xc0, 0xa8, 0xf7, 0xfa, 0x0, 0x0, 0x0, 0x4, 0x0, 0xa, 0xc0, 0xa8, 0xf5, 0xfa, 0x0, 0x0, 0x0, 0x7, 0x0, 0x8, 0x13, 0xc4, 0x0, 0x0, 0x0, 0x8, 0x0, 0x8, 0x13, 0xc4, 0x0, 0x0, 0x0, 0x9, 0x0, 0xa, 0x5a, 0xa2, 0x9b, 0x98, 0x0, 0x0, 0x0, 0xa, 0x0, 0xa, 0x0, 0x1, 0xd2, 0xf4, 0x0, 0x0, 0x0, 0xb, 0x0, 0x7, 0x1, 0x0, 0x0, 0x0, 0xc, 0x0, 0xa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xe, 0x0, 0x6, 0x0, 0x0, 0x0, 0xf, 0x2, 0xa7, 0x53, 0x49, 0x50, 0x2f, 0x32, 0x2e, 0x30, 0x20, 0x32, 0x30, 0x30, 0x20, 0x4f, 0x4b, 0xd, 0xa, 0x43, 0x61, 0x6c, 0x6c, 0x2d, 0x49, 0x44, 0x3a, 0x20, 0x42, 0x43, 0x30, 0x39, 0x39, 0x38, 0x38, 0x34, 0x40, 0x36, 0x64, 0x66, 0x63, 0x66, 0x66, 0x65, 0x38, 0xd, 0xa, 0x43, 0x53, 0x65, 0x71, 0x3a, 0x20, 0x32, 0x31, 0x35, 0x38, 0x33, 0x34, 0x34, 0x38, 0x39, 0x20, 0x4f, 0x50, 0x54, 0x49, 0x4f, 0x4e, 0x53, 0xd, 0xa, 0x46, 0x72, 0x6f, 0x6d, 0x3a, 0x20, 0x3c, 0x73, 0x69, 0x70, 0x3a, 0x31, 0x39, 0x32, 0x2e, 0x31, 0x36, 0x38, 0x2e, 0x31, 0x31, 0x31, 0x2e, 0x31, 0x31, 0x31, 0x3a, 0x35, 0x30, 0x36, 0x30, 0x3e, 0x3b, 0x74, 0x61, 0x67, 0x3d, 0x36, 0x64, 0x66, 0x63, 0x66, 0x66, 0x65, 0x38, 0x2b, 0x31, 0x2b, 0x62, 0x30, 0x61, 0x39, 0x30, 0x30, 0x30, 0x33, 0x2b, 0x63, 0x39, 0x65, 0x66, 0x63, 0x32, 0x30, 0x62, 0xd, 0xa, 0x54, 0x6f, 0x3a, 0x20, 0x3c, 0x73, 0x69, 0x70, 0x3a, 0x31, 0x39, 0x32, 0x2e, 0x31, 0x36, 0x38, 0x2e, 0x31, 0x31, 0x31, 0x2e, 0x31, 0x31, 0x31, 0x3a, 0x35, 0x30, 0x36, 0x30, 0x3b, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x3d, 0x75, 0x64, 0x70, 0x3e, 0x3b, 0x74, 0x61, 0x67, 0x3d, 0x31, 0x38, 0x30, 0x34, 0x61, 0x34, 0x37, 0x64, 0x2b, 0x31, 0x2b, 0x65, 0x31, 0x30, 0x35, 0x30, 0x34, 0x37, 0x30, 0x2b, 0x62, 0x31, 0x32, 0x38, 0x61, 0x35, 0x36, 0x39, 0xd, 0xa, 0x56, 0x69, 0x61, 0x3a, 0x20, 0x53, 0x49, 0x50, 0x2f, 0x32, 0x2e, 0x30, 0x2f, 0x55, 0x44, 0x50, 0x20, 0x31, 0x39, 0x32, 0x2e, 0x31, 0x36, 0x38, 0x2e, 0x31, 0x31, 0x31, 0x2e, 0x31, 0x31, 0x31, 0x3a, 0x35, 0x30, 0x36, 0x30, 0x3b, 0x62, 0x72, 0x61, 0x6e, 0x63, 0x68, 0x3d, 0x7a, 0x39, 0x68, 0x47, 0x34, 0x62, 0x4b, 0x2b, 0x32, 0x31, 0x66, 0x31, 0x31, 0x33, 0x65, 0x37, 0x65, 0x33, 0x64, 0x30, 0x34, 0x63, 0x38, 0x34, 0x36, 0x31, 0x34, 0x38, 0x61, 0x39, 0x61, 0x64, 0x37, 0x36, 0x30, 0x37, 0x61, 0x65, 0x66, 0x61, 0x31, 0x2b, 0x36, 0x64, 0x66, 0x63, 0x66, 0x66, 0x65, 0x38, 0x2b, 0x31, 0xd, 0xa, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x3a, 0x20, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0xd, 0xa, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x2d, 0x4c, 0x65, 0x6e, 0x67, 0x74, 0x68, 0x3a, 0x20, 0x37, 0x38, 0xd, 0xa, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x2d, 0x54, 0x79, 0x70, 0x65, 0x3a, 0x20, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x73, 0x64, 0x70, 0xd, 0xa, 0x53, 0x75, 0x70, 0x70, 0x6f, 0x72, 0x74, 0x65, 0x64, 0x3a, 0x20, 0x31, 0x30, 0x30, 0x72, 0x65, 0x6c, 0x2c, 0x20, 0x74, 0x69, 0x6d, 0x65, 0x72, 0xd, 0xa, 0x41, 0x63, 0x63, 0x65, 0x70, 0x74, 0x2d, 0x4c, 0x61, 0x6e, 0x67, 0x75, 0x61, 0x67, 0x65, 0x3a, 0x20, 0x65, 0x6e, 0xd, 0xa, 0x41, 0x63, 0x63, 0x65, 0x70, 0x74, 0x2d, 0x45, 0x6e, 0x63, 0x6f, 0x64, 0x69, 0x6e, 0x67, 0x3a, 0x20, 0x69, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0xd, 0xa, 0x41, 0x63, 0x63, 0x65, 0x70, 0x74, 0x3a, 0x20, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x73, 0x64, 0x70, 0x2c, 0x20, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x69, 0x73, 0x75, 0x70, 0x2c, 0x20, 0x6d, 0x75, 0x6c, 0x74, 0x69, 0x70, 0x61, 0x72, 0x74, 0x2f, 0x6d, 0x69, 0x78, 0x65, 0x64, 0xd, 0xa, 0x41, 0x6c, 0x6c, 0x6f, 0x77, 0x3a, 0x20, 0x49, 0x4e, 0x56, 0x49, 0x54, 0x45, 0x2c, 0x20, 0x41, 0x43, 0x4b, 0x2c, 0x20, 0x43, 0x41, 0x4e, 0x43, 0x45, 0x4c, 0x2c, 0x20, 0x42, 0x59, 0x45, 0x2c, 0x20, 0x4f, 0x50, 0x54, 0x49, 0x4f, 0x4e, 0x53, 0x2c, 0x20, 0x4e, 0x4f, 0x54, 0x49, 0x46, 0x59, 0x2c, 0x20, 0x50, 0x52, 0x41, 0x43, 0x4b, 0x2c, 0x20, 0x55, 0x50, 0x44, 0x41, 0x54, 0x45, 0x2c, 0x20, 0x49, 0x4e, 0x46, 0x4f, 0x2c, 0x20, 0x52, 0x45, 0x46, 0x45, 0x52, 0xd, 0xa, 0xd, 0xa, 0x76, 0x3d, 0x30, 0xd, 0xa, 0x6f, 0x3d, 0x2d, 0x20, 0x30, 0x20, 0x30, 0x20, 0x49, 0x4e, 0x20, 0x49, 0x50, 0x34, 0x20, 0x30, 0x2e, 0x30, 0x2e, 0x30, 0x2e, 0x30, 0xd, 0xa, 0x73, 0x3d, 0x2d, 0xd, 0xa, 0x63, 0x3d, 0x49, 0x4e, 0x20, 0x49, 0x50, 0x34, 0x20, 0x30, 0x2e, 0x30, 0x2e, 0x30, 0x2e, 0x30, 0xd, 0xa, 0x74, 0x3d, 0x30, 0x20, 0x30, 0xd, 0xa, 0x6d, 0x3d, 0x61, 0x75, 0x64, 0x69, 0x6f, 0x20, 0x30, 0x20, 0x52, 0x54, 0x50, 0x2f, 0x41, 0x56, 0x50, 0x20, 0x38}
var jsonIsup = []byte(`{"cic": 3324, "dpc": 5236, "opc": 7414, "msg_name": "IAM", "msg_type": 1, "hop_counter": 28, "forward_call": {"isup": 1, "isup_name": "ISDN user part used all the way", "isdn_access": 0, "sccp_method": 0, "interworking": 0, "isup_preference": 0, "isdn_access_name": "originating access non-ISDN", "sccp_method_name": "no indication", "end_to_end_method": 0, "interworking_name": "no interworking encountered (No. 7 signalling all the way)", "isup_preference_name": "ISDN user part preferred all the way", "end_to_end_information": 0, "end_to_end_method_name": "no end-to-end method available (only link-by-link method available)", "end_to_end_information_name": "no end-to-end information available", "national_international_call": 0, "national_international_call_name": "call to be treated as a national call"}, "called_number": {"inn": 0, "npi": 1, "num": "9495368", "ton": 3, "inn_name": "routing to internal network number allowed", "npi_name": "ISDN (Telephony) numbering plan (ITU-T Recommendation E.164)", "ton_name": "national (significant) number"}, "calling_party": {"num": 10, "name": "ordinary calling subscriber"}, "calling_number": {"ni": 0, "npi": 1, "num": "9182789", "ton": 3, "ni_name": "complete", "npi_name": "ISDN (Telephony) numbering plan (ITU-T Recommendation E.164)", "restrict": 1, "screened": 3, "ton_name": "national (significant) number", "restrict_name": "presentation restricted", "screened_name": "network provided"}, "transmission_medium": {"num": 3, "name": "3.1 kHz audio"}, "nature_of_connnection": {"satellite": 0, "echo_device": 0, "satellite_name": "no satellite circuit in the connection", "continuity_check": 0, "echo_device_name": "outgoing echo control device not included", "continuity_check_name": "continuity check not required"}}`)

func init() {
	config.Setting.DBWorker = runtime.NumCPU()
	config.Setting.DBBulk = 10000
	config.Setting.DBDriver = "mock"
	config.Setting.DBShema = "mock"
	go func() {
		db := New("mock")
		db.Chan = dbCh
		if err := db.Run(); err != nil {
			fmt.Println(err)
		}
	}()
}

func BenchmarkInsert(b *testing.B) {
	hep, err := decoder.DecodeHEP(hepPacket)
	if err != nil {
		b.Error(err)
	}

	//hep.SIP.CseqMethod = "INVITE"
	runtime.GC()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dbCh <- hep
	}
}

func TestMakeISUPDataHeader(t *testing.T) {
	bpp := bytebufferpool.Get()
	defer bytebufferpool.Put(bpp)
	wantJi := `{"cic":3324,"dpc":5236,"opc":7414,"msg_name":"IAM","called_number":"9495368","calling_number":"9182789","callid":"5236:7414:3324"}`
	_, gotJi := makeISUPDataHeader(jsonIsup, bpp)
	if gotJi != wantJi {
		t.Errorf("[TestMakeISUPDataHeader failed]\nwant:\t%s\ngot:\t%s", wantJi, gotJi)
	}
}

func BenchmarkMakeISUPDataHeader(b *testing.B) {
	bpp := bytebufferpool.Get()
	defer bytebufferpool.Put(bpp)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = makeISUPDataHeader(jsonIsup, bpp)
	}
}
