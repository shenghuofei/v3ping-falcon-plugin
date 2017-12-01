package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var src string
var lock = new(sync.RWMutex)

type cfgdata struct {
	Dest     string `json:"dest"`
	Metric   string `json:"metric"`
	Interval int    `json:"interval"`
}

type cfg struct {
	Errno int `json:"errno"`
	Data  []cfgdata
}

func get_src() {
	src, _ = os.Hostname()
}

func get_cfg() cfg {
	url := "http://cfg-server.addr/?q=" + src
	retry := 5
	args := map[string]interface{}{"q": src}
	//fmt.Println(args)
	cfgres := cfg{}
	for i := 0; i < retry; i++ {
		err, res := push("GET", url, "application/json", args, 3)
		//fmt.Printf("%s",res)
		if err != nil {
			time.Sleep(1 * time.Second)
		} else {
			err = json.Unmarshal(res, &cfgres)
			if err == nil && cfgres.Errno == 0 {
				//fmt.Println("fetch white api data to json fail")
				return cfgres
			}
		}
	}
	os.Exit(2)
	return cfgres
}

func do_ping(timestamp int64, metric, ip string, interval int, inte float64, res *[]map[string]interface{}, wg1 *sync.WaitGroup) {
	url := "http://127.0.0.1:1988/v1/push"
	relip := strings.TrimSpace(strings.Split(ip, "@")[0])
	//interval = 3
	var loss float64 = -1
	cmd := fmt.Sprintf("sudo ping -n -w 45 -c %d -i %.2f -q %s", interval, inte, relip)
	ecmd := exec.Command("/bin/sh", "-c", cmd)
	//fmt.Println("do",ip,relip,cmd,ecmd)
	stdout, _ := ecmd.StdoutPipe()
	if err := ecmd.Start(); err == nil {
		ping_res, _ := ioutil.ReadAll(stdout)
		//fmt.Printf("ping_res %v\n", string(ping_res))
		pat := `(\d+)% packet loss`
		reg := regexp.MustCompile(pat)
		loss_res := reg.FindStringSubmatch(string(ping_res))
		if len(loss_res) >= 2 {
			loss, _ = strconv.ParseFloat(loss_res[1], 64)
		}
	}

	mes := map[string]interface{}{"metric": "v3ping", "endpoint": "wfy-mac", "timestamp": timestamp, "step": 60, "value": loss, "counterType": "GAUGE", "tags": "type=detail,dest=" + ip + ",src=" + src + ",note=" + metric}
	jsonmes := []map[string]interface{}{mes}
	//fmt.Println(mes,url)
	push("POST", url, "application/json", jsonmes, 3)
	r := map[string]interface{}{"ip": ip, "loss": loss}
	//fmt.Println(r)
	lock.Lock()
	*res = append(*res, r)
	lock.Unlock()
	wg1.Done()
}

func get_one_ping(dest, metric string, interval int, wg *sync.WaitGroup) {
	var wg1 sync.WaitGroup
	url := "http://127.0.0.1:1988/v1/push"
	dest_ip_list := strings.Split(dest, ",")
	dest_ip_list = RemoveDuplicate(dest_ip_list)
	sum_loss := 0.0
	send_time := 40.0
	inte := Round(send_time/float64(interval), 2)
	res := []map[string]interface{}{}
	timestamp := time.Now().Unix()
	for _, ip := range dest_ip_list {
		//fmt.Println(ip)
		wg1.Add(1)
		go do_ping(timestamp, metric, ip, interval, inte, &res, &wg1)
	}
	wg1.Wait()
	//fmt.Println(res,sum_loss)
	res_len := len(res)
	var avg float64 = -1
	if res_len != 0 {
		for _, l := range res {
			sum_loss += l["loss"].(float64)
		}
		avg = Round(sum_loss/float64(res_len), 2)
	}
	//fmt.Println(sum_loss,res_len)
	//fmt.Println("avg",avg)
	mes := map[string]interface{}{"metric": "v3ping", "endpoint": "wfy-mac", "timestamp": timestamp, "step": 60, "value": avg, "counterType": "GAUGE", "tags": "type=avg,src=" + src + ",note=" + metric}
	jsonmes := []map[string]interface{}{mes}
	//fmt.Println(mes,url)
	push("POST", url, "application/json", jsonmes, 3)
	//err,postres := push("POST",url,"application/json",jsonmes,3)
	//fmt.Println(err,string(postres))
	//fmt.Println(metric)
	//fmt.Println("########")
	wg.Done()
}

func get_all_ping() {
	var wg sync.WaitGroup
	get_src()
	cfg := get_cfg()
	//fmt.Println(cfg)
	//fmt.Println("=================")
	//cfglen := len(cfg.Data)
	for _, c := range cfg.Data {
		//fmt.Println(c.Dest)
		//fmt.Println(c.Metric)
		//fmt.Println(c.Interval)
		wg.Add(1)
		go get_one_ping(c.Dest, c.Metric, c.Interval, &wg)
	}
	wg.Wait()
	//fmt.Println("Done")
}

func main() {
	/*tick := time.Tick(60 * time.Second)
	  for {
	      select {
	          case c := <-tick:
	              get_all_ping()
	              fmt.Println(c)
	      }
	  }*/
	get_all_ping()
}
