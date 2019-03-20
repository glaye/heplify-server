package input

import (
	"net"
	"sync/atomic"
	"time"

	"glaye/heplify-server/logp"
)

func (h *HEPInput) serveUDP(addr string) {
	ua, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		logp.Err("%v", err)
	}

	uc, err := net.ListenUDP("udp", ua)
	if err != nil {
		logp.Err("%v", err)
	}
	defer func() {
		logp.Info("stopping UDP listener on %s", uc.LocalAddr())
		uc.Close()
	}()

	for {
		select {
		case <-h.quitUDP:
			h.quitUDP <- true
			return
		default:
		}
		uc.SetReadDeadline(time.Now().Add(1e9))
		buf := h.buffer.Get().([]byte)
		n, err := uc.Read(buf)
		if err != nil {
			continue
		} else if n > maxPktLen {
			logp.Warn("received too big packet with %d bytes", n)
			atomic.AddUint64(&h.stats.ErrCount, 1)
			continue
		}
		h.inCh <- buf[:n]
		atomic.AddUint64(&h.stats.PktCount, 1)
	}
}
