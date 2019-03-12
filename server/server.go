package input

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"glaye/heplify-server/config"
	"glaye/heplify-server/database"
	"glaye/heplify-server/decoder"
	"glaye/heplify-server/logp"
	"glaye/heplify-server/metric"
	"glaye/heplify-server/queue"
	"glaye/heplify-server/remotelog"
)

type HEPInput struct {
	useDB  bool
	useMQ  bool
	usePM  bool
	useES  bool
	useLK  bool
	inCh   chan []byte
	dbCh   chan *decoder.HEP
	mqCh   chan []byte
	pmCh   chan *decoder.HEP
	esCh   chan *decoder.HEP
	lkCh   chan *decoder.HEP
	buffer *sync.Pool
	quit   chan struct{}
	stats  HEPStats
	wg     *sync.WaitGroup
}

type HEPStats struct {
	DupCount uint64
	ErrCount uint64
	HEPCount uint64
	PktCount uint64
}

const maxPktLen = 8192

func NewHEPInput() *HEPInput {
	h := &HEPInput{
		inCh:   make(chan []byte, 40000),
		buffer: &sync.Pool{New: func() interface{} { return make([]byte, maxPktLen) }},
		quit:   make(chan struct{}),
		wg:     &sync.WaitGroup{},
	}
	if len(config.Setting.DBAddr) > 2 {
		h.useDB = true
		h.dbCh = make(chan *decoder.HEP, config.Setting.DBBuffer)
	}
	if len(config.Setting.MQAddr) > 2 && len(config.Setting.MQDriver) > 2 {
		h.useMQ = true
		h.mqCh = make(chan []byte, 40000)
	}
	if len(config.Setting.PromAddr) > 2 {
		h.usePM = true
		h.pmCh = make(chan *decoder.HEP, 40000)
	}
	if len(config.Setting.ESAddr) > 2 {
		h.useES = true
		h.esCh = make(chan *decoder.HEP, 40000)
	}
	if len(config.Setting.LokiURL) > 2 {
		h.useLK = true
		h.lkCh = make(chan *decoder.HEP, config.Setting.LokiBuffer)
	}

	return h
}

func (h *HEPInput) Run() {
	for n := 0; n < runtime.NumCPU()*4; n++ {
		go h.hepWorker()
	}

	if h.useDB {
		go func() {
			d := database.New(config.Setting.DBDriver)
			d.Chan = h.dbCh

			if err := d.Run(); err != nil {
				logp.Err("%v", err)
			}
		}()
	}

	if h.useMQ {
		go func() {
			q := queue.New(config.Setting.MQDriver)
			q.Topic = config.Setting.MQTopic
			q.Chan = h.mqCh

			if err := q.Run(); err != nil {
				logp.Err("%v", err)
			}
		}()
	}

	if h.usePM {
		go func() {
			m := metric.New("prometheus")
			m.Chan = h.pmCh

			if err := m.Run(); err != nil {
				logp.Err("%v", err)
			}
		}()
	}

	if h.useES {
		go func() {
			r := remotelog.New("elasticsearch")
			r.Chan = h.esCh

			if err := r.Run(); err != nil {
				logp.Err("%v", err)
			}
		}()
	}

	if h.useLK {
		go func() {
			l := remotelog.New("loki")
			l.Chan = h.lkCh

			if err := l.Run(); err != nil {
				logp.Err("%v", err)
			}
		}()
	}

	logp.Info("start %s with %#v\n", config.Version, config.Setting)
	go h.logStats()

	if config.Setting.HEPAddr != "" {
		h.wg.Add(1)
		go h.serveUDP()
	}
	if config.Setting.HEPTCPAddr != "" {
		go h.serveTCP()
	}
	if config.Setting.HEPTLSAddr != "" {
		go h.serveTLS()
	}
	if config.Setting.HTTPAddr != "" {
		go h.serveHTTP()
	}
}

func (h *HEPInput) End() {
	logp.Info("stopping heplify-server...")
	close(h.quit)
	h.wg.Wait()
	close(h.inCh)
	logp.Info("heplify-server has been stopped")
}

func (h *HEPInput) hepWorker() {
	var (
		lastWarn = time.Now()
		msg      = h.buffer.Get().([]byte)
		ok       bool
	)

OUT:
	for {
		h.buffer.Put(msg[:maxPktLen])

		select {
		case msg, ok = <-h.inCh:
			if !ok {
				break OUT
			}
		}

		hepPkt, err := decoder.DecodeHEP(msg)
		if err != nil {
			atomic.AddUint64(&h.stats.ErrCount, 1)
			continue
		} else if hepPkt.ProtoType == 0 {
			atomic.AddUint64(&h.stats.DupCount, 1)
			continue
		}

		atomic.AddUint64(&h.stats.HEPCount, 1)

		if h.useDB {
			select {
			case h.dbCh <- hepPkt:
			default:
				if time.Since(lastWarn) > 5e8 {
					logp.Warn("overflowing db channel, please adjust DBWorker or DBBuffer setting")
				}
				lastWarn = time.Now()
			}
		}

		if h.usePM {
			select {
			case h.pmCh <- hepPkt:
			default:
				if time.Since(lastWarn) > 5e8 {
					logp.Warn("overflowing metric channel")
				}
				lastWarn = time.Now()
			}
		}

		if h.useMQ {
			select {
			case h.mqCh <- append([]byte{}, msg...):
			default:
				if time.Since(lastWarn) > 5e8 {
					logp.Warn("overflowing queue channel")
				}
				lastWarn = time.Now()
			}
		}

		if h.useES {
			select {
			case h.esCh <- hepPkt:
			default:
				if time.Since(lastWarn) > 5e8 {
					logp.Warn("overflowing elasticsearch channel")
				}
				lastWarn = time.Now()
			}
		}

		if h.useLK {
			select {
			case h.lkCh <- hepPkt:
			default:
				if time.Since(lastWarn) > 5e8 {
					logp.Warn("overflowing loki channel")
				}
				lastWarn = time.Now()
			}
		}
	}
}

func (h *HEPInput) logStats() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		logp.Info("stats since last 5 minutes. PPS: %d, HEP: %d, Filtered: %d, Error: %d",
			atomic.LoadUint64(&h.stats.PktCount)/300,
			atomic.LoadUint64(&h.stats.HEPCount),
			atomic.LoadUint64(&h.stats.DupCount),
			atomic.LoadUint64(&h.stats.ErrCount),
		)
		atomic.StoreUint64(&h.stats.PktCount, 0)
		atomic.StoreUint64(&h.stats.HEPCount, 0)
		atomic.StoreUint64(&h.stats.DupCount, 0)
		atomic.StoreUint64(&h.stats.ErrCount, 0)
	}
}
