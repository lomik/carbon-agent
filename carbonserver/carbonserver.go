/*
 * Copyright 2013-2016 Fabian Groffen, Damian Gryski, Vladimir Smirnov
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package carbonserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/NYTimes/gziphandler"
	logger "github.com/Sirupsen/logrus"
	pb "github.com/dgryski/carbonzipper/carbonzipperpb"
	trigram "github.com/dgryski/go-trigram"
	"github.com/dgryski/httputil"
	"github.com/gogo/protobuf/proto"
	pickle "github.com/kisielk/og-rek"
	"github.com/lomik/go-carbon/cache"
	"github.com/lomik/go-carbon/helper"
	"github.com/lomik/go-carbon/points"
	whisper "github.com/lomik/go-whisper"
)

type metricStruct struct {
	RenderRequests          uint64
	RenderErrors            uint64
	NotFound                uint64
	FindRequests            uint64
	FindErrors              uint64
	FindZero                uint64
	InfoRequests            uint64
	InfoErrors              uint64
	CacheHits               uint64
	CacheMiss               uint64
	CacheFullHits           uint64
	CacheTimeouts           uint64
	CacheRequests           uint64
	CacheWaitTimeNS         uint64
	CacheWaitTimeOverheadNS uint64
	DiskWaitTimeNS          uint64
	DiskRequests            uint64
	PointsFetched           uint64
}

type CarbonserverListener struct {
	helper.Stoppable
	cache             *cache.Cache
	readTimeout       time.Duration
	whisperData       string
	buckets           int
	maxGlobs          int
	scanFrequency     time.Duration
	metricsAsCounters bool
	tcpListener       *net.TCPListener

	fileIdx atomic.Value

	metrics     metricStruct
	exitChan    chan struct{}
	timeBuckets []uint64
}

type fileIndex struct {
	idx   trigram.Index
	files []string
}

func NewCarbonserverListener(cache *cache.Cache) *CarbonserverListener {
	return &CarbonserverListener{
		// Config variables
		metricsAsCounters: false,
		cache:             cache,
	}
}

func (listener *CarbonserverListener) SetWhisperData(whisperData string) {
	listener.whisperData = strings.TrimRight(whisperData, "/")
}
func (listener *CarbonserverListener) SetMaxGlobs(maxGlobs int) {
	listener.maxGlobs = maxGlobs
}
func (listener *CarbonserverListener) SetBuckets(buckets int) {
	listener.buckets = buckets
}
func (listener *CarbonserverListener) SetScanFrequency(scanFrequency time.Duration) {
	listener.scanFrequency = scanFrequency
}
func (listener *CarbonserverListener) SetReadTimeout(readTimeout time.Duration) {
	listener.readTimeout = readTimeout
}
func (listener *CarbonserverListener) SetMetricsAsCounters(metricsAsCounters bool) {
	listener.metricsAsCounters = metricsAsCounters
}

func (listener *CarbonserverListener) CurrentFileIndex() *fileIndex {
	p := listener.fileIdx.Load()
	if p == nil {
		return nil
	}
	return p.(*fileIndex)
}
func (listener *CarbonserverListener) UpdateFileIndex(fidx *fileIndex) { listener.fileIdx.Store(fidx) }

func (listener *CarbonserverListener) fileListUpdater(dir string, tick <-chan time.Time, force <-chan struct{}, exit <-chan struct{}) {

	for {

		select {
		case <-exit:
			return
		case <-tick:
		case <-force:
		}

		var files []string

		t0 := time.Now()

		err := filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
			if err != nil {
				logger.Infof("[carbonserver] error processing %q: %v", p, err)
				return nil
			}

			if info.IsDir() || strings.HasSuffix(info.Name(), ".wsp") {
				files = append(files, strings.TrimPrefix(p, listener.whisperData))
			}

			return nil
		})

		logger.Debugln("[carbonserver] file scan took", time.Since(t0), ",", len(files), "items")
		t0 = time.Now()

		idx := trigram.NewIndex(files)

		logger.Debugln("[carbonserver] indexing took", time.Since(t0), len(idx), "trigrams")

		pruned := idx.Prune(0.95)

		logger.Debugln("[carbonserver] pruned", pruned, "common trigrams")

		if err == nil {
			listener.UpdateFileIndex(&fileIndex{idx, files})
		}
	}
}

func (listener *CarbonserverListener) expandGlobs(query string) ([]string, []bool) {
	var useGlob bool

	if star := strings.IndexByte(query, '*'); strings.IndexByte(query, '[') == -1 && strings.IndexByte(query, '?') == -1 && (star == -1 || star == len(query)-1) {
		useGlob = true
	}

	/* things to glob:
	 * - carbon.relays  -> carbon.relays
	 * - carbon.re      -> carbon.relays, carbon.rewhatever
	 * - carbon.[rz]    -> carbon.relays, carbon.zipper
	 * - carbon.{re,zi} -> carbon.relays, carbon.zipper
	 * - match is either dir or .wsp file
	 * unfortunately, filepath.Glob doesn't handle the curly brace
	 * expansion for us */

	query = strings.Replace(query, ".", "/", -1)

	var globs []string
	if !strings.HasSuffix(query, "*") {
		globs = append(globs, query+".wsp")
	}
	globs = append(globs, query)
	// TODO(dgryski): move this loop into its own function + add tests
	for {
		bracematch := false
		var newglobs []string
		for _, glob := range globs {
			lbrace := strings.Index(glob, "{")
			rbrace := -1
			if lbrace > -1 {
				rbrace = strings.Index(glob[lbrace:], "}")
				if rbrace > -1 {
					rbrace += lbrace
				}
			}

			if lbrace > -1 && rbrace > -1 {
				bracematch = true
				expansion := glob[lbrace+1 : rbrace]
				parts := strings.Split(expansion, ",")
				for _, sub := range parts {
					if len(newglobs) > listener.maxGlobs {
						break
					}
					newglobs = append(newglobs, glob[:lbrace]+sub+glob[rbrace+1:])
				}
			} else {
				if len(newglobs) > listener.maxGlobs {
					break
				}
				newglobs = append(newglobs, glob)
			}
		}
		globs = newglobs
		if !bracematch {
			break
		}
	}

	var files []string

	fidx := listener.CurrentFileIndex()

	if fidx != nil && !useGlob {
		// use the index
		docs := make(map[trigram.DocID]struct{})

		for _, g := range globs {

			gpath := "/" + g

			ts := extractTrigrams(g)

			// TODO(dgryski): If we have 'not enough trigrams' we
			// should bail and use the file-system glob instead

			ids := fidx.idx.QueryTrigrams(ts)

			for _, id := range ids {
				docid := trigram.DocID(id)
				if _, ok := docs[docid]; !ok {
					matched, err := filepath.Match(gpath, fidx.files[id])
					if err == nil && matched {
						docs[docid] = struct{}{}
					}
				}
			}
		}

		for id := range docs {
			files = append(files, listener.whisperData+fidx.files[id])
		}

		sort.Strings(files)
	}

	// Not an 'else' clause because the trigram-searching code might want
	// to fall back to the file-system glob

	if useGlob || fidx == nil {
		// no index or we were asked to hit the filesystem
		for _, g := range globs {
			nfiles, err := filepath.Glob(listener.whisperData + "/" + g)
			if err == nil {
				files = append(files, nfiles...)
			}
		}
	}

	leafs := make([]bool, len(files))
	for i, p := range files {
		s, err := os.Stat(p)
		if err != nil {
			continue
		}
		p = p[len(listener.whisperData+"/"):]
		if !s.IsDir() && strings.HasSuffix(p, ".wsp") {
			p = p[:len(p)-4]
			leafs[i] = true
		} else {
			leafs[i] = false
		}
		files[i] = strings.Replace(p, "/", ".", -1)
	}

	return files, leafs
}

func (listener *CarbonserverListener) findHandler(wr http.ResponseWriter, req *http.Request) {
	// URL: /metrics/find/?local=1&format=pickle&query=the.metric.path.with.glob

	t0 := time.Now()

	atomic.AddUint64(&listener.metrics.FindRequests, 1)

	req.ParseForm()
	format := req.FormValue("format")
	query := req.FormValue("query")

	if format != "json" && format != "pickle" && format != "protobuf" {
		atomic.AddUint64(&listener.metrics.FindErrors, 1)
		logger.Infof("[carbonserver] dropping invalid uri (format=%s): %s",
			format, req.URL.RequestURI())
		http.Error(wr, "Bad request (unsupported format)",
			http.StatusBadRequest)
		return
	}

	if query == "" {
		atomic.AddUint64(&listener.metrics.FindErrors, 1)
		logger.Infof("[carbonserver] dropping invalid request (query=): %s", req.URL.RequestURI())
		http.Error(wr, "Bad request (no query)", http.StatusBadRequest)
		return
	}

	files, leafs := listener.expandGlobs(query)

	if format == "json" || format == "protobuf" {
		name := req.FormValue("query")
		response := pb.GlobResponse{
			Name:    &name,
			Matches: make([]*pb.GlobMatch, 0),
		}

		for i, p := range files {
			response.Matches = append(response.Matches, &pb.GlobMatch{Path: proto.String(p), IsLeaf: proto.Bool(leafs[i])})
		}

		var b []byte
		var err error
		switch format {
		case "json":
			b, err = json.Marshal(response)
		case "protobuf":
			b, err = proto.Marshal(&response)
		}
		if err != nil {
			atomic.AddUint64(&listener.metrics.FindErrors, 1)
			logger.Infof("[carbonserver] failed to create %s data for glob %s: %s",
				format, *response.Name, err)
			return
		}
		wr.Write(b)
	} else if format == "pickle" {
		// [{'metric_path': 'metric', 'intervals': [(x,y)], 'isLeaf': True},]
		var metrics []map[string]interface{}
		var m map[string]interface{}

		for i, p := range files {
			m = make(map[string]interface{})
			m["metric_path"] = p
			// m["intervals"] = dunno how to do a tuple here
			m["isLeaf"] = leafs[i]
			metrics = append(metrics, m)
		}

		wr.Header().Set("Content-Type", "application/pickle")
		pEnc := pickle.NewEncoder(wr)
		pEnc.Encode(metrics)
	}

	if len(files) == 0 {
		// to get an idea how often we search for nothing
		atomic.AddUint64(&listener.metrics.FindZero, 1)
	}

	logger.Debugf("[carbonserver] find: %d hits for %s in %v", len(files), req.FormValue("query"), time.Since(t0))
	return
}

func fetchCachedData(data []points.Point, fetchFromTime, fetchUntilTime, step int32) ([]float64, int32, int32) {
	var cacheFromTime, cacheUntilTime int32
	cachedValues := make([]float64, 0)
	prevTs := int32(0)
	if len(data) == 0 {
		return cachedValues, cacheFromTime, cacheUntilTime
	}

	for _, item := range data {
		ts := int32(item.Timestamp) - int32(item.Timestamp)%step
		// Data not from requested range, we don't need it.
		if ts >= fetchUntilTime || ts <= fetchFromTime || prevTs == ts {
			continue
		}
		for i := prevTs; prevTs != 0 && i < ts; i += step {
			cachedValues = append(cachedValues, math.NaN())
		}
		prevTs = ts
		cachedValues = append(cachedValues, item.Value)
		if cacheFromTime == 0 || cacheFromTime > ts {
			cacheFromTime = ts
		}
		if cacheUntilTime == 0 || cacheUntilTime < ts {
			cacheUntilTime = ts
		}
	}
	return cachedValues, cacheFromTime, cacheUntilTime
}

func (listener *CarbonserverListener) fetchHandler(wr http.ResponseWriter, req *http.Request) {
	// URL: /render/?target=the.metric.name&format=pickle&from=1396008021&until=1396022421

	atomic.AddUint64(&listener.metrics.RenderRequests, 1)
	req.ParseForm()
	metric := req.FormValue("target")
	format := req.FormValue("format")
	from := req.FormValue("from")
	until := req.FormValue("until")

	t0 := time.Now()

	// Make sure we log which metric caused a panic()
	defer func() {
		if r := recover(); r != nil {
			var buf [4096]byte
			runtime.Stack(buf[:], false)
			logger.Infof("[carbonserver] panic handling request: %v\n%s\n%s", r, req.RequestURI, string(buf[:]))
		}
	}()

	if format != "json" && format != "pickle" && format != "protobuf" {
		atomic.AddUint64(&listener.metrics.RenderErrors, 1)
		logger.Infof("[carbonserver] dropping invalid uri (format=%s): %s",
			format, req.URL.RequestURI())
		http.Error(wr, "Bad request (unsupported format)",
			http.StatusBadRequest)
		return
	}

	files, leafs := listener.expandGlobs(metric)

	var badTime bool

	i, err := strconv.Atoi(from)
	if err != nil {
		logger.Infof("[carbonserver] fromTime (%s) invalid: %s (in %s)", from, err, req.URL.RequestURI())
		badTime = true
	}
	fromTime := int(i)
	i, err = strconv.Atoi(until)
	if err != nil {
		logger.Infof("[carbonserver] untilTime (%s) invalid: %s (in %s)", from, err, req.URL.RequestURI())
		badTime = true
	}
	untilTime := int(i)

	if badTime {
		atomic.AddUint64(&listener.metrics.RenderErrors, 1)
		http.Error(wr, "Bad request (invalid from/until time)", http.StatusBadRequest)
		return
	}

	var multi pb.MultiFetchResponse
	for i, metric := range files {
		if !leafs[i] {
			logger.Debugf("[carbonserver] skipping directory = %q", metric)
			// can't fetch a directory
			continue
		}
		var step int32
		var cacheFromTime int32
		var cacheUntilTime int32
		cacheGotEverything := false

		// We need to obtain the metadata from whisper file anyway.
		path := listener.whisperData + "/" + strings.Replace(metric, ".", "/", -1) + ".wsp"
		w, err := whisper.Open(path)
		if err != nil {
			// the FE/carbonzipper often requests metrics we don't have
			// We shouldn't really see this any more -- expandGlobs() should filter them out
			atomic.AddUint64(&listener.metrics.NotFound, 1)
			logger.Infof("[carbonserver] error opening %q: %v", path, err)
			continue
		}

		retentions := w.Retentions()
		now := int(time.Now().Unix())
		diff := now - fromTime
		bestStep := int32(retentions[0].SecondsPerPoint())
		for _, retention := range retentions {
			if retention.MaxRetention() >= diff {
				step = int32(retention.SecondsPerPoint())
				break
			}
		}

		fetchUntilTime := int32(untilTime)
		fetchFromTime := int32(fromTime)

		if step == 0 {
			atomic.AddUint64(&listener.metrics.RenderErrors, 1)
			logger.Infof("[carbonserver] Can't find proper archive for the request for metric %q", path)
			continue
		}

		cachedValues := make([]float64, 0)
		var cacheData []points.Point

		cacheStartTime := time.Now()
		if step != bestStep {
			logger.Debugf("[carbonserver] Cache is not supported for this query (required step != best step). path=%q fromTime=%v untilTime=%v step=%v bestStep=%v", path, fromTime, untilTime, step, bestStep)
		} else {
			// query cache
			cacheData = listener.cache.Get(metric)
			waitTime := uint64(time.Since(cacheStartTime).Nanoseconds())
			atomic.AddUint64(&listener.metrics.CacheWaitTimeOverheadNS, waitTime)
		}

		if cacheData != nil {
			atomic.AddUint64(&listener.metrics.CacheRequests, 1)

			cachedValues, cacheFromTime, cacheUntilTime = fetchCachedData(cacheData, fetchFromTime, fetchUntilTime, step)
			logger.Debugf("[carbonserver] fetched cached metric=%v from=%v until=%v", metric, cacheFromTime, cacheUntilTime)

			if cacheFromTime != 0 {
				if cacheFromTime <= fetchFromTime {
					for fetchUntilTime > cacheUntilTime {
						cachedValues = append(cachedValues, math.NaN())
						cacheUntilTime += step
					}
					fetchUntilTime = cacheUntilTime
					fetchFromTime = cacheFromTime
					cacheGotEverything = true
					atomic.AddUint64(&listener.metrics.CacheFullHits, 1)
				} else {
					atomic.AddUint64(&listener.metrics.CacheHits, 1)
				}
			} else {
				atomic.AddUint64(&listener.metrics.CacheMiss, 1)
			}
			waitTime := uint64(time.Since(cacheStartTime).Nanoseconds())
			atomic.AddUint64(&listener.metrics.CacheWaitTimeNS, waitTime)
		}

		// End of cache query
		var values []float64
		if !cacheGotEverything {
			atomic.AddUint64(&listener.metrics.DiskRequests, 1)
			diskStartTime := time.Now()
			logger.Debugf("[carbonserver] fetching disk metric=%v from=%v until=%v", metric, fetchFromTime, fetchUntilTime)

			points, err := w.Fetch(int(fetchFromTime), int(fetchUntilTime))
			w.Close()
			if err != nil {
				atomic.AddUint64(&listener.metrics.RenderErrors, 1)
				logger.Infof("[carbonserver] failed to fetch points from %s: %s", path, err)
				continue
			}

			if points == nil {
				atomic.AddUint64(&listener.metrics.NotFound, 1)
				logger.Debugf("[carbonserver] Metric time range not found: metric=%s from=%d to=%d ", metric, fromTime, untilTime)
				continue
			}
			values = points.Values()

			fetchFromTime = int32(points.FromTime())
			fetchUntilTime = int32(points.UntilTime())
			step = int32(points.Step())

			waitTime := uint64(time.Since(diskStartTime).Nanoseconds())
			atomic.AddUint64(&listener.metrics.DiskWaitTimeNS, waitTime)
		} else {
			values = make([]float64, 0)
			w.Close()
		}

		startTime := fetchFromTime
		stopTime := fetchUntilTime
		points := (stopTime - startTime) / step
		atomic.AddUint64(&listener.metrics.PointsFetched, uint64(points))
		response := pb.FetchResponse{
			Name:      proto.String(metric),
			StartTime: &startTime,
			StopTime:  &stopTime,
			StepTime:  &step,
			Values:    make([]float64, points),
			IsAbsent:  make([]bool, points),
		}

		ts := startTime
		cacheCursor := 0
		diskCursor := 0
		for i := range response.Values {
			p := math.NaN()
			if cacheFromTime != 0 && ts >= cacheFromTime && ts <= cacheUntilTime {
				p = cachedValues[cacheCursor]
				cacheCursor++
			} else if fetchFromTime != 0 && ts >= fetchFromTime && ts <= fetchUntilTime {
				p = values[diskCursor]
				diskCursor++
			}

			if math.IsNaN(p) {
				response.Values[i] = 0
				response.IsAbsent[i] = true
			} else {
				response.Values[i] = p
				response.IsAbsent[i] = false
			}
			ts += step
		}

		multi.Metrics = append(multi.Metrics, &response)
	}

	var b []byte
	switch format {
	case "json":
		wr.Header().Set("Content-Type", "application/json")
		b, err = json.Marshal(multi)

	case "protobuf":
		wr.Header().Set("Content-Type", "application/protobuf")
		b, err = proto.Marshal(&multi)

	case "pickle":
		// transform protobuf data into what pickle expects
		//[{'start': 1396271100, 'step': 60, 'name': 'metric',
		//'values': [9.0, 19.0, None], 'end': 1396273140}

		var response []map[string]interface{}

		for _, metric := range multi.GetMetrics() {

			var m map[string]interface{}

			m = make(map[string]interface{})
			m["start"] = metric.StartTime
			m["step"] = metric.StepTime
			m["end"] = metric.StopTime
			m["name"] = metric.Name

			mv := make([]interface{}, len(metric.Values))
			for i, p := range metric.Values {
				if metric.IsAbsent[i] {
					mv[i] = nil
				} else {
					mv[i] = p
				}
			}

			m["values"] = mv
			response = append(response, m)
		}

		wr.Header().Set("Content-Type", "application/pickle")
		var buf bytes.Buffer
		pEnc := pickle.NewEncoder(&buf)
		err = pEnc.Encode(response)
		b = buf.Bytes()
	}

	if err != nil {
		atomic.AddUint64(&listener.metrics.RenderErrors, 1)
		logger.Infof("[carbonserver] failed to create %s data for %s: %s", format, "<metric>", err)
		return
	}
	wr.Write(b)

	logger.Debugf("[carbonserver] fetch: served %q from %d to %d in %v", metric, fromTime, untilTime, time.Since(t0))
}

func (listener *CarbonserverListener) infoHandler(wr http.ResponseWriter, req *http.Request) {
	// URL: /info/?target=the.metric.name&format=json

	atomic.AddUint64(&listener.metrics.InfoRequests, 1)
	req.ParseForm()
	metric := req.FormValue("target")
	format := req.FormValue("format")

	if format == "" {
		format = "json"
	}

	if format != "json" && format != "protobuf" {
		atomic.AddUint64(&listener.metrics.InfoErrors, 1)
		logger.Infof("[carbonserver] dropping invalid uri (format=%s): %s",
			format, req.URL.RequestURI())
		http.Error(wr, "Bad request (unsupported format)",
			http.StatusBadRequest)
		return
	}

	path := listener.whisperData + "/" + strings.Replace(metric, ".", "/", -1) + ".wsp"
	w, err := whisper.Open(path)

	if err != nil {
		atomic.AddUint64(&listener.metrics.NotFound, 1)
		logger.Debugf("[carbonserver] failed to %s", err)
		http.Error(wr, "Metric not found", http.StatusNotFound)
		return
	}

	defer w.Close()

	aggr := w.AggregationMethod()
	maxr := int32(w.MaxRetention())
	xfiles := float32(w.XFilesFactor())
	rets := make([]*pb.Retention, 0, 4)
	for _, retention := range w.Retentions() {
		spp := int32(retention.SecondsPerPoint())
		nop := int32(retention.NumberOfPoints())
		rets = append(rets, &pb.Retention{
			SecondsPerPoint: &spp,
			NumberOfPoints:  &nop,
		})
	}
	response := pb.InfoResponse{
		Name:              &metric,
		AggregationMethod: &aggr,
		MaxRetention:      &maxr,
		XFilesFactor:      &xfiles,
		Retentions:        rets,
	}

	var b []byte
	switch format {
	case "json":
		b, err = json.Marshal(response)
	case "protobuf":
		b, err = proto.Marshal(&response)
	}
	if err != nil {
		atomic.AddUint64(&listener.metrics.RenderErrors, 1)
		logger.Infof("[carbonserver] failed to create %s data for %s: %s", format, path, err)
		return
	}
	wr.Write(b)

	logger.Debugf("[carbonserver] served info for %s", metric)
	return
}

func (listener *CarbonserverListener) Stat(send helper.StatCallback) {
	sender := helper.SendAndSubstractUint64
	if listener.metricsAsCounters {
		sender = helper.SendUint64
	}
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	pauseNS := uint64(m.PauseTotalNs)
	alloc := uint64(m.Alloc)
	totalAlloc := uint64(m.TotalAlloc)
	numGC := uint64(m.NumGC)

	sender("render_requests", &listener.metrics.RenderRequests, send)
	sender("render_errors", &listener.metrics.RenderErrors, send)
	sender("notfound", &listener.metrics.NotFound, send)
	sender("find_requests", &listener.metrics.FindRequests, send)
	sender("find_errors", &listener.metrics.FindErrors, send)
	sender("find_zero", &listener.metrics.FindZero, send)
	sender("cache_hits", &listener.metrics.CacheHits, send)
	sender("cache_miss", &listener.metrics.CacheMiss, send)
	sender("cache_full_hits", &listener.metrics.CacheFullHits, send)
	sender("cache_timeouts", &listener.metrics.CacheTimeouts, send)
	sender("cache_wait_time_ns", &listener.metrics.CacheWaitTimeNS, send)
	sender("cache_wait_time_overhead_ns", &listener.metrics.CacheWaitTimeOverheadNS, send)
	sender("cache_requests", &listener.metrics.CacheRequests, send)
	sender("disk_wait_time_ns", &listener.metrics.DiskWaitTimeNS, send)
	sender("disk_requests", &listener.metrics.DiskRequests, send)
	sender("points_fetched", &listener.metrics.PointsFetched, send)

	sender("alloc", &alloc, send)
	sender("total_alloc", &totalAlloc, send)
	sender("num_gc", &numGC, send)
	sender("pause_ns", &pauseNS, send)
	for i := 0; i <= listener.buckets; i++ {
		sender(fmt.Sprintf("requests_in_%dms_to_%dms", i*100, (i+1)*100), &listener.timeBuckets[i], send)
	}
}

func (listener *CarbonserverListener) Stop() error {
	listener.exitChan <- struct{}{}
	listener.tcpListener.Close()
	return nil
}

func (listener *CarbonserverListener) Listen(listen string) error {
	logger.Warnln("[carbonserver] carbonserver support is still experimental, use at your own risk")
	logger.Infoln("[carbonserver] starting carbonserver")

	logger.Infof("[carbonserver] reading whisper files from: %s", listener.whisperData)

	logger.Infof("[carbonserver] maximum brace expansion set to: %d", listener.maxGlobs)

	if listener.scanFrequency != 0 {
		logger.Infoln("[carbonserver] use file cache with scan frequency", listener.scanFrequency)
		force := make(chan struct{})
		listener.exitChan = make(chan struct{})
		go listener.fileListUpdater(listener.whisperData, time.Tick(listener.scanFrequency), force, listener.exitChan)
		force <- struct{}{}
	}

	// +1 to track every over the number of buckets we track
	listener.timeBuckets = make([]uint64, listener.buckets+1)

	carbonserverMux := http.NewServeMux()
	carbonserverMux.HandleFunc("/metrics/find/", httputil.TrackConnections(httputil.TimeHandler(listener.findHandler, listener.bucketRequestTimes)))
	carbonserverMux.HandleFunc("/render/", httputil.TrackConnections(httputil.TimeHandler(listener.fetchHandler, listener.bucketRequestTimes)))
	carbonserverMux.HandleFunc("/info/", httputil.TrackConnections(httputil.TimeHandler(listener.infoHandler, listener.bucketRequestTimes)))

	carbonserverMux.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "User-agent: *\nDisallow: /")
	})

	logger.Infof("[carbonserver] listening on %s", listen)
	tcpAddr, err := net.ResolveTCPAddr("tcp", listen)
	if err != nil {
		return err
	}
	listener.tcpListener, err = net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return err
	}

	go http.Serve(listener.tcpListener, gziphandler.GzipHandler(carbonserverMux))
	return nil
}

func (listener *CarbonserverListener) renderTimeBuckets() interface{} {
	return listener.timeBuckets
}

func (listener *CarbonserverListener) bucketRequestTimes(req *http.Request, t time.Duration) {

	ms := t.Nanoseconds() / int64(time.Millisecond)

	bucket := int(math.Log(float64(ms)) * math.Log10E)

	if bucket < 0 {
		bucket = 0
	}

	if bucket < listener.buckets {
		atomic.AddUint64(&listener.timeBuckets[bucket], 1)
	} else {
		// Too big? Increment overflow bucket and log
		atomic.AddUint64(&listener.timeBuckets[listener.buckets], 1)
		logger.Infof("[carbonserver] Slow Request: %s: %s", t.String(), req.URL.String())
	}
}

func extractTrigrams(query string) []trigram.T {

	if len(query) < 3 {
		return nil
	}

	var start int
	var i int

	var trigrams []trigram.T

	for i < len(query) {
		if query[i] == '[' || query[i] == '*' || query[i] == '?' {
			trigrams = trigram.Extract(query[start:i], trigrams)

			if query[i] == '[' {
				for i < len(query) && query[i] != ']' {
					i++
				}
			}

			start = i + 1
		}
		i++
	}

	if start < i {
		trigrams = trigram.Extract(query[start:i], trigrams)
	}

	return trigrams
}
