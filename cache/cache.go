package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
)

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

func getOperator(called string) (string, error) {
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
	return "", errors.New("no match operator!!!")
}

// NewRedisPool 返回redis连接池
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

func set(k, v string) {
	c := pool.Get()
	defer c.Close()
	_, err := c.Do("SET", k, v)
	if err != nil {
		fmt.Println("set error", err.Error())
	}
}

func incr(k string) {
	c := pool.Get()
	defer c.Close()
	_, err := c.Do("INCR", k)
	if err != nil {
		fmt.Println("set error", err.Error())
	}
}

func getStringValue(k string) (string, error) {
	c := pool.Get()
	defer c.Close()
	value, err := redis.String(c.Do("GET", k))
	if err != nil {
		//fmt.Println("Get Error: ", err.Error())
		return "", err
	}
	return value, nil
}

func SetKeyExpire(k string, ex int) {
	c := pool.Get()
	defer c.Close()
	_, err := c.Do("EXPIRE", k, ex)
	if err != nil {
		fmt.Println("set error", err.Error())
	}
}

func CheckKey(k string) bool {
	c := pool.Get()
	defer c.Close()
	exist, err := redis.Bool(c.Do("EXISTS", k))
	if err != nil {
		fmt.Println(err)
		return false
	} else {
		return exist
	}
}

func DelKey(k string) error {
	c := pool.Get()
	defer c.Close()
	_, err := c.Do("DEL", k)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func SetJson(k string, data interface{}) error {
	c := pool.Get()
	defer c.Close()
	value, _ := json.Marshal(data)
	n, _ := c.Do("SETNX", k, value)
	if n != int64(1) {
		return errors.New("set failed")
	}
	return nil
}

func getJsonByte(k string) ([]byte, error) {
	c := pool.Get()
	jsonGet, err := redis.Bytes(c.Do("GET", k))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return jsonGet, nil
}

func getCity(called string) (string, error) {
	if len(called) < 5 {
		return "", errors.New("called len not enough")
	}
	tmp := called
	if tmp[:1] == "0" {
		//fmt.Println("切割了")
		tmp = tmp[1:8]
	} else {
		tmp = tmp[0:7]
	}
	//fmt.Println(tmp)
	rtn, err := getStringValue(tmp)
	return rtn, err
}
func main() {
	//incr("a")
	//incr("a")
	//incr("a")
	//incr("a")
	//incr("a")
	//incr("a")
	//incr("a")
	//incr("a")
	//incr("")
	set("countCallerArray", "95078001,95078002,95078003,95078006,95078007,95078008,95078711,95753,95066")
	//fmt.Println(getStringValue("a"))
	//s := "13810107660"
	//s = string([]rune(s)[1:])
	//getCity("13810107660")
	origcaller := "1381010766095078003"
	//callers := "95078001,95078002,95078003,95078711,95753"
	callers, _ := getStringValue("countCallerArray")
	callerArray := strings.Split(callers, ",")
	for _, caller := range callerArray {
		//for的第一个参数是索引，这里用不上
		fmt.Println(caller)
		match, _ := regexp.MatchString(".*"+caller, origcaller)
		fmt.Println(match)
	}
	fmt.Println("------------------------")
	called := "1559083"
	cityString, err := getCity(called)
	if err != nil {
		fmt.Printf("%s not found city!!!\n", called)
	} else {
		fmt.Printf("%s found city is %s\n", called, cityString)
	}
	//fmt.Printf("城市：%s", cityString)
	//fmt.Printf("错误：%s", err)
	//node = 10240120060
	node := 10240120060
	fmt.Printf("%T\n", node)
	nodeID := strconv.Itoa(int(node))
	//nodeID = strconv.FormatInt(int64(node), 10)

	fmt.Println("333333333333333333333333")
	fmt.Println(nodeID)

	r, _ := regexp.Compile("1(3[4-9]|47|5[0-2|7-9]|70[3|5|6]|78|8[2-4|7|8]|98|65)[0-9]*")
	print(r.MatchString("13810107660"))
	fmt.Println("===========================")
	fmt.Println(getOperator("13810107660"))
	fmt.Println(getOperator("18632366206"))
	fmt.Println(getOperator("13332353323"))

	a, b := getCity("1")
	fmt.Println(a, b)

}
