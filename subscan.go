package main

import (
"sync"
"os"
"bufio"
"strings"
"flag"
"time"
"fmt"
"sync/atomic"
"net"
"github.com/valyala/fasthttp"
"github.com/joeguo/tldextract"
"strconv"
)

var wg sync.WaitGroup
var ops uint64 = 0 //atomic counter
var outpath = ""
var show bool = true

type Job struct {
	Work string
}

func doRequest(url string) bool {
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(url)
	resp := fasthttp.AcquireResponse()
	client := &fasthttp.Client{}
	c := client.DoTimeout(req, resp, 5*time.Second)
	if c != nil {
		return false
	}
	return true
}

func AppendStringToFile(path, text string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(text + "\r\n") // use "\n" in Unix, "\r\n" is Windows only
	if err != nil {
		return err
	}
	return nil
}

func sum(url string)bool {
	req := doRequest(url)
	if req == false {
		return false
	}
	fmt.Println(true, url)
	AppendStringToFile(outpath, url)
	atomic.AddUint64(&ops, 1)
	return true
}

func w(r string) {
	addr, err := net.LookupAddr(r)
	if err != nil {
		return
	}
	url := "http://" + strings.TrimRight(addr[0], ".")
	req := doRequest(url)
	if req == false {
		return
	}
	cache := "tld.cache"
	extract, _ := tldextract.New(cache, false)
	result := extract.Extract(url)
	if show == true{
		fmt.Println("Scanning:","edc."+result.Root+"."+result.Tld)
	}
	s := sum("edc."+result.Root+"."+result.Tld)
	if s == true{return}
	for i := 0; i < 100; i++ {
		t := strconv.Itoa(i)
		url := "edc"+t+"."+result.Root+"."+result.Tld
		if show == true{
			fmt.Println("Scanning:",url)
		}
		s := sum(url)
		if s == true{return}
	}
}

func produce(jobs chan<- *Job, inpath string) {
	defer wg.Done()
	g, _ := os.Open(inpath)
	scanner := bufio.NewScanner(g)
	for scanner.Scan() {
		r := scanner.Text()
		a := strings.TrimSpace(r)
		jobs <- &Job{a}
	}
	defer g.Close()
	close(jobs)
}

func consume(id int, jobs <-chan *Job) {
	for job := range jobs {
		w(job.Work)
	}
	defer wg.Done()
}

func main() {
	max_workers := flag.Int("t", 5000, "Threads as int")
	inpath := flag.String("i", "ips.csv", "in.txt")
	outpath_ := flag.String("o", "ip_out.txt", "out.txt")
	show_ := flag.Bool("s", true, "show log")
	flag.Parse()
	outpath = *outpath_
	show = *show_
	jobs := make(chan *Job, 150000) // Buffered channel
	start := time.Now()             //start timer
	// Start consumers:
	for i := 0; i < *max_workers+1; i++ { // creating consumers
		wg.Add(1)
		go consume(i, jobs)
	}
	wg.Add(1)
	// Start producing
	go produce(jobs, *inpath)
	wg.Wait() // Wait all consumers to finish processing jobs
	elapsed := time.Since(start)
	opsFinal := atomic.LoadUint64(&ops)
	fmt.Println("\n[INFO]found", opsFinal, "valid accounts.")
	fmt.Println("[INFO]time elapsed:", elapsed)
}
