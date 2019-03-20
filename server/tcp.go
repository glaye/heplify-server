package input

import (
	"bufio"
	"encoding/binary"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"glaye/heplify-server/logp"
)

func (h *HEPInput) serveTCP(addr string) {
	ta, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		logp.Err("%v", err)
		return
	}

	ln, err := net.ListenTCP("tcp", ta)
	if err != nil {
		logp.Err("%v", err)
		return
	}

	var wg sync.WaitGroup

	for {
		select {
		case <-h.quitTCP:
			logp.Info("stopping TCP listener on %s", ln.Addr())
			ln.Close()
			wg.Wait()
			close(h.exitedTCP)
			return
		default:
			ln.SetDeadline(time.Now().Add(1e9))
			conn, err := ln.Accept()
			if err != nil {
				if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
					continue
				}
				logp.Err("failed to accept TCP connection: %v", err.Error())
			}
			logp.Info("new TCP connection %s -> %s", conn.RemoteAddr(), conn.LocalAddr())
			wg.Add(1)
			go func() {
				h.handleTCP(conn)
				wg.Done()
			}()
		}
	}
}

func (h *HEPInput) handleTCP(c net.Conn) {
	defer func() {
		logp.Info("closing TCP connection from %s", c.RemoteAddr())
		c.Close()
	}()

	r := bufio.NewReader(c)
	readBytes := func(buffer []byte) (int, error) {
		n := uint(0)
		for n < uint(len(buffer)) {
			nn, err := r.Read(buffer[n:])
			n += uint(nn)
			if err != nil {
				return 0, err
			}
		}
		return int(n), nil
	}
	for {
		select {
		case <-h.quitTCP:
			return
		default:
		}

		c.SetReadDeadline(time.Now().Add(1e9))
		buf := h.buffer.Get().([]byte)
		hb, err := r.Peek(6)
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			} else {
				return
			}
		} else {
			size := binary.BigEndian.Uint16(hb[4:6])
			if size > maxPktLen {
				logp.Warn("unexpected packet, did you send TLS into plain TCP input?")
				return
			}
			n, err := readBytes(buf[:size])
			if err != nil || n > maxPktLen {
				logp.Warn("%v, unusal packet size with %d bytes", err, n)
				atomic.AddUint64(&h.stats.ErrCount, 1)
				continue
			}
			h.inCh <- buf[:n]
			atomic.AddUint64(&h.stats.PktCount, 1)
		}
	}
}
