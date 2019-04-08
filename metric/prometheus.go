package metric

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/pkg/errors"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/coocood/freecache"
	lru "github.com/hashicorp/golang-lru"
	"glaye/heplify-server/config"
	"glaye/heplify-server/decoder"
	"glaye/heplify-server/logp"
)

type Prometheus struct {
	TargetEmpty bool
	TargetIP    []string
	TargetName  []string
	TargetMap   map[string]string
	TargetConf  *sync.RWMutex
	cache       *freecache.Cache
	lruID       *lru.Cache
}

const (
	RedisURL            = "redis://172.16.24.80:6400"
	redisMaxIdle        = 3   //最大空闲连接数
	redisIdleTimeoutSec = 240 //最大空闲连接时间
	RedisPassword       = "cqtredis1234"
)

var (
	pool        = NewRedisPool(RedisURL)
	ydRegexp, _ = regexp.Compile("1(3[4-9]|47|5[0-2|7-9]|70[3|5|6]|78|8[2-4|7|8]|98|65)[0-9]*")
	ltRegexp, _ = regexp.Compile("1(3[0-2]|45|5[5|6]|70[4|7|8|9]|7[1|5|6]|8[5|6]|6(6|7))[0-9]*")
	dxRegexp, _ = regexp.Compile("1(33|49|53|70[0|1|2]|7[3|7]|8[0|1|9]|9[1|9])[0-9]*")
)

func (p *Prometheus) setup() (err error) {
	p.TargetConf = new(sync.RWMutex)
	p.TargetIP = strings.Split(cutSpace(config.Setting.PromTargetIP), ",")
	p.TargetName = strings.Split(cutSpace(config.Setting.PromTargetName), ",")

	if len(p.TargetIP) == len(p.TargetName) && p.TargetIP != nil && p.TargetName != nil {
		if len(p.TargetIP[0]) == 0 || len(p.TargetName[0]) == 0 {
			logp.Info("expose metrics without or unbalanced targets")
			p.TargetIP[0] = ""
			p.TargetName[0] = ""
			p.TargetEmpty = true
			p.cache = freecache.NewCache(60 * 1024 * 1024)
		} else {
			for i := range p.TargetName {
				logp.Info("prometheus tag assignment %d: %s -> %s", i+1, p.TargetIP[i], p.TargetName[i])
			}
			p.TargetMap = make(map[string]string)
			for i := 0; i < len(p.TargetName); i++ {
				p.TargetMap[p.TargetIP[i]] = p.TargetName[i]
			}
		}
	} else {
		logp.Info("please give every PromTargetIP a unique IP and PromTargetName a unique name")
		return fmt.Errorf("faulty PromTargetIP or PromTargetName")
	}

	p.lruID, err = lru.New(1e5)
	if err != nil {
		return err
	}

	return err
}

func (p *Prometheus) expose(hCh chan *decoder.HEP) {

	for pkt := range hCh {
		nodeID := strconv.Itoa(int(pkt.NodeID))
		labelType := decoder.HEPTypeString(pkt.ProtoType)

		packetsByType.WithLabelValues(labelType).Inc()
		packetsBySize.WithLabelValues(labelType).Set(float64(len(pkt.Payload)))

		if pkt.SIP != nil && pkt.ProtoType == 1 {
			var st, dt string
			if !p.TargetEmpty {
				var ok bool
				st, ok = p.TargetMap[pkt.SrcIP]
				if ok {
					methodResponses.WithLabelValues(st, "src", nodeID, pkt.SIP.FirstMethod, pkt.SIP.CseqMethod).Inc()
				}
				dt, ok = p.TargetMap[pkt.DstIP]
				if ok {
					methodResponses.WithLabelValues(dt, "dst", nodeID, pkt.SIP.FirstMethod, pkt.SIP.CseqMethod).Inc()
				}
			} else {
				_, err := p.cache.Get([]byte(pkt.CID + pkt.SIP.FirstMethod + pkt.SIP.CseqMethod))
				if err == nil {
					continue
				}
				err = p.cache.Set([]byte(pkt.CID+pkt.SIP.FirstMethod+pkt.SIP.CseqMethod), nil, 600)
				if err != nil {
					logp.Warn("%v", err)
				}
				methodResponses.WithLabelValues("", "", nodeID, pkt.SIP.FirstMethod, pkt.SIP.CseqMethod).Inc()

			}

			// customCounter city operator and caller
			p.customCounter(pkt.SIP.FirstMethod, pkt.SIP.CseqMethod, pkt.SIP.FromUser, pkt.SIP.ToUser)

			p.requestDelay(st, dt, pkt.CID, pkt.SIP.FirstMethod, pkt.SIP.CseqMethod, pkt.Timestamp)

			if pkt.SIP.RTPStatVal != "" {
				p.dissectXRTPStats(st, pkt.SIP.RTPStatVal)
			}
			if pkt.SIP.ReasonVal != "" && strings.Contains(pkt.SIP.ReasonVal, "850") {
				reasonCause.WithLabelValues(extractXR("cause=", pkt.SIP.ReasonVal), pkt.Node).Inc()
			}
		} else if pkt.ProtoType == 5 {
			p.dissectRTCPStats(pkt.Node, []byte(pkt.Payload))
		} else if pkt.ProtoType == 34 {
			p.dissectRTPStats(pkt.Node, []byte(pkt.Payload))
		} else if pkt.ProtoType == 35 {
			p.dissectRTCPXRStats(pkt.Node, pkt.Payload)
		} else if pkt.ProtoType == 38 {
			p.dissectHoraclifixStats([]byte(pkt.Payload))
		} else if pkt.ProtoType == 112 {
			logSeverity.WithLabelValues(pkt.Node, pkt.CID, pkt.Host).Inc()
		} else if pkt.ProtoType == 1032 {
			p.dissectJanusStats([]byte(pkt.Payload))
		}
	}
}

func (p *Prometheus) requestDelay(st, dt, cid, sm, cm string, ts time.Time) {
	if !p.TargetEmpty && st == "" {
		return
	}

	//TODO: tweak performance avoid double lru add
	if (sm == "INVITE" && cm == "INVITE") || (sm == "REGISTER" && cm == "REGISTER") {
		_, ok := p.lruID.Get(cid)
		if !ok {
			p.lruID.Add(cid, ts)
			p.lruID.Add(st+cid, ts)
		}
	}

	if (cm == "INVITE" || cm == "REGISTER") && (sm == "180" || sm == "183" || sm == "200") {
		did := dt + cid
		t, ok := p.lruID.Get(did)
		if ok {
			if cm == "INVITE" {
				srd.WithLabelValues(st, dt).Set(float64(ts.Sub(t.(time.Time))))
			} else {
				rrd.WithLabelValues(st, dt).Set(float64(ts.Sub(t.(time.Time))))
			}
			p.lruID.Remove(cid)
			p.lruID.Remove(did)
		}
	}
}

func (p *Prometheus) customCounter(firstMethod, cseqMethod, caller, called string) {
	//判断主叫是否在统计列表中
	countItem, isCountCaller := p.isCountCaller(caller)
	if isCountCaller {
		// 获得被叫所在省份
		cityString, err := getProvinceFromCalled(called)
		if err != nil {
			return
		}
		// 获得被叫所属运营商
		operator, err := getOperatorFromCalled(called)
		if err != nil {
			logp.Err(err.Error())
			return
		}
		cityOperatorCallerMethodResponses.WithLabelValues(cityString, operator, countItem, firstMethod, cseqMethod).Inc()
	}
}

// redis连接池
func NewRedisPool(redisURL string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     redisMaxIdle,
		IdleTimeout: redisIdleTimeoutSec * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.DialURL(redisURL)
			if err != nil {
				return nil, fmt.Errorf("redis connection error: %s", err)
			}
			//验证redis密码
			if _, authErr := c.Do("AUTH", RedisPassword); authErr != nil {
				return nil, fmt.Errorf("redis auth password error: %s", authErr)
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			if err != nil {
				return fmt.Errorf("ping redis error: %s", err)
			}
			return nil
		},
	}
}

// redis里取key的value
func getStringValue(k string) (string, error) {
	c := pool.Get()
	defer c.Close()
	username, err := redis.String(c.Do("GET", k))
	if err != nil {
		//fmt.Println("Get Error: ", err.Error())
		//logp.Err(err.Error())
		return "", err
	}
	return username, nil
}

// 判断主叫是否在统计列表中
func (p *Prometheus) isCountCaller(key string) (string, bool) {

	callers, err := getStringValue("countCallerArray")
	if err != nil {
		return "", false
	}
	callerArray := strings.Split(callers, ",")
	for _, caller := range callerArray {
		//for的第一个参数是索引，这里用不上
		//fmt.Println(caller)
		match, _ := regexp.MatchString(".*"+caller, key)
		if match {
			return caller, true
		}
	}
	return "", false //到这里就是没匹配上

}

// 获得被叫所属省份
func getProvinceFromCalled(called string) (string, error) {
	if len(called) < 5 {
		return "", errors.New("called len not enough")
	}
	tmp := called
	if tmp[:1] == "0" {
		//fmt.Println("切割了")
		if len(tmp) >= 8 {
			tmp = tmp[1:8]
		} else {
			return "", errors.New(called + "called len not enough")
		}
	} else {
		if len(tmp) >= 7 {
			tmp = tmp[0:7]
		} else {
			return "", errors.New(called + "called len not enough")
		}
	}
	//fmt.Println(tmp)
	rtn, err := getStringValue(tmp)
	return rtn, err
}

// 获得被叫所属运营商
func getOperatorFromCalled(called string) (string, error) {
	match := ydRegexp.MatchString(called)
	if match {
		return "yidong", nil
	}
	match = ltRegexp.MatchString(called)
	if match {
		return "liantong", nil
	}
	match = dxRegexp.MatchString(called)
	if match {
		return "dianxin", nil
	}
	return "", errors.New(called + " no match operator!!!")
}
