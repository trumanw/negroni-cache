package negronicache

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"sync"
	"time"
)

const (
	CacheHeader     = "X-Cache"
	ProxyDateHeader = "Proxy-Date"
	GRPCStatusOK    = 0
)

var Writes sync.WaitGroup

var storeable = map[int]bool{
	// it seems like the grpc gateway can also accept response status code
	// to be 0. And it will automatically transfer the 0 to 200.
	GRPCStatusOK:                    true,
	http.StatusOK:                   true,
	http.StatusFound:                true,
	http.StatusNonAuthoritativeInfo: true,
	http.StatusMultipleChoices:      true,
	http.StatusMovedPermanently:     true,
	http.StatusGone:                 true,
	http.StatusNotFound:             true,
}

var cacheableByDefault = map[int]bool{
	GRPCStatusOK:                    true,
	http.StatusOK:                   true,
	http.StatusFound:                true,
	http.StatusNotModified:          true,
	http.StatusNonAuthoritativeInfo: true,
	http.StatusMultipleChoices:      true,
	http.StatusMovedPermanently:     true,
	http.StatusGone:                 true,
	http.StatusPartialContent:       true,
}

// Middleware is the cache middlware for negroni
type Middleware struct {
	Shared    bool
	validator *Validator
	cache     Cache
}

// NewMiddleware retrieves an instance of Cache handler
func NewMiddleware(cache Cache) *Middleware {
	return &Middleware{
		cache:  cache,
		Shared: false,
	}
}

func (ch *Middleware) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	cReq, err := NewCacheRequest(r)
	if err != nil {
		http.Error(rw, "invalid request: "+err.Error(),
			http.StatusBadRequest)
		return
	}

	if !cReq.isCacheable() {
		debugf("request not cacheable")
		rw.Header().Set(CacheHeader, "SKIP")
		ch.UpstreamWithCache(rw, cReq, next)
		return
	}

	res, err := ch.LookupInCached(cReq)
	if err != nil && err != ErrNotFoundInCache {
		http.Error(rw, "lookup error: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	cacheType := "private"
	if ch.Shared {
		cacheType = "shared"
	}

	if err == ErrNotFoundInCache {
		if cReq.CacheControl.Has("only-if-cached") {
			http.Error(rw, "key not in cache",
				http.StatusGatewayTimeout)
			return
		}
		debugf("%s %s not in %s cache", r.Method, r.URL.String(), cacheType)
		ch.UpstreamWithCache(rw, cReq, next)
		return
	}

	debugf("%s %s found in %s cache", r.Method, r.URL.String(), cacheType)

	res.Header().Set(CacheHeader, "HIT")
	ch.ServeResource(res, rw, cReq)

	if err := res.Close(); err != nil {
		errorf("Error closing resource: %s", err.Error())
	}
}

// ServeResource is used for wrapping the ResponseWriter and returning the request.
func (ch *Middleware) ServeResource(res *Resource, rw http.ResponseWriter, req *CacheRequest) {
	for key, headers := range res.Header() {
		for _, header := range headers {
			rw.Header().Add(key, header)
		}
	}

	age, err := res.Age()
	if err != nil {
		http.Error(rw, "Error calculating age: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	if age > (time.Hour*24) && res.HeuristicFreshness() > (time.Hour*24) {
		rw.Header().Add("Warning", `113 - "Heuristic Expiration"`)
	}

	freshness, err := ch.Freshness(res, req)
	if err != nil || freshness <= 0 {
		rw.Header().Add("Warning", `110 - "Response is Stale"`)
	}

	debugf("resource is %s old, updating age from %s",
		age.String(), rw.Header().Get("Age"))

	rw.Header().Set("Age", fmt.Sprintf("%.f", math.Floor(age.Seconds())))
	rw.Header().Set("Via", res.Via())

	// hacky handler for non-ok statuses
	if res.Status() != http.StatusOK {
		rw.WriteHeader(res.Status())
		io.Copy(rw, res)
	} else {
		http.ServeContent(rw, req.Request, "", res.LastModified(), res)
	}
}

// UpstreamWithCache returns the request to a specific handler and stores the result
func (ch *Middleware) UpstreamWithCache(rw http.ResponseWriter, r *CacheRequest, next http.HandlerFunc) {
	rs := NewResponseStreamer(rw)
	rdr, err := rs.Stream.NextReader()
	if err != nil {
		debugf("error creating next stream reader: %v", err)
		rw.Header().Set(CacheHeader, "SKIP")
		next(rw, r.Request)
		return
	}

	t := Clock()
	rw.Header().Set(CacheHeader, "SKIP")

	next(rs, r.Request)
	rs.Stream.Close()

	// Just the headers
	res := NewResourceBytes(rs.StatusCode, nil, rs.Header())
	if !ch.isCacheable(res, r) {
		rdr.Close()
		debugf("resource is uncacheable")
		rs.Header().Set(CacheHeader, "SKIP")
		return
	}

	b, err := ioutil.ReadAll(rdr)
	rdr.Close()
	if err != nil {
		debugf("error reading stream: %v", err)
		rs.Header().Set(CacheHeader, "SKIP")
		return
	}
	debugf("full upstream response took %s", Clock().Sub(t).String())
	res.ReadSeekCloser = &ByteReadSeekCloser{bytes.NewReader(b)}

	// if age, err := CorrectedAge(res.Header(), t, Clock()); err == nil {
	//     res.Header().Set("Age", strconv.Itoa(int(math.Ceil(age.Seconds()))))
	// } else {
	//     debugf("error calculating corrected age: %s", err.Error())
	// }

	rs.Header().Set(ProxyDateHeader, Clock().Format(http.TimeFormat))
	// Cache the http response
	ch.CacheResource(res, r)
}

// CacheResource can store the response in the cache.
func (ch *Middleware) CacheResource(res *Resource, r *CacheRequest) {
	Writes.Add(1)

	go func() {
		defer Writes.Done()
		t := Clock()
		keys := []string{r.Key.String()}
		headers := res.Header()

		if ch.Shared {
			res.RemovePrivateHeaders()
		}

		// store a secondary vary version
		if vary := headers.Get("Vary"); vary != "" {
			keys = append(keys, r.Key.Vary(vary, r.Request).String())
		}

		if err := ch.cache.Store(res, keys...); err != nil {
			errorf("storing resources %#v failed with error: %s", keys, err.Error())
		}

		debugf("stored resources %+v in %s", keys, Clock().Sub(t))
	}()
}

// LookupInCached finds the best matching Resource for the
// request, or nil and ErrNotFoundInCache if none is found
func (ch *Middleware) LookupInCached(req *CacheRequest) (*Resource, error) {
	res, err := ch.cache.Retrieve(req.Key.String())
	// HEAD requests can possibly be served from GET
	if err == ErrNotFoundInCache && req.Method == "HEAD" {
		res, err = ch.cache.Retrieve(req.Key.ForMethod("GET").String())
		if err != nil {
			return nil, err
		}

		if res.HasExplicitExpiration() && req.isCacheable() {
			debugf("using cached GET request for serving HEAD")
			return res, nil
		}

		return nil, ErrNotFoundInCache
	} else if err != nil {
		return res, err
	}

	// Secondary lookup for Vary
	// if vary := res.Header().Get("Vary"); vary != "" {
	//     debugf("Original retrieved key: %s", req.Key.String())
	//     debugf("Varied : %s", req.Key.Vary(vary, req.Request).String())
	// 	res, err = ch.cache.Retrieve(req.Key.Vary(vary, req.Request).String())
	// 	if err != nil {
	// 		return res, err
	// 	}
	// }

	return res, nil
}

// Freshness returns the duration that a requested resource will be fresh for
func (ch *Middleware) Freshness(res *Resource, r *CacheRequest) (time.Duration, error) {
	maxAge, err := res.MaxAge(ch.Shared)
	if err != nil {
		return time.Duration(0), err
	}

	if r.CacheControl.Has("max-age") {
		reqMaxAge, err := r.CacheControl.Duration("max-age")
		if err != nil {
			return time.Duration(0), err
		}

		if reqMaxAge < maxAge {
			debugf("using request max-age of %s", reqMaxAge.String())
			maxAge = reqMaxAge
		}
	}

	age, err := res.Age()
	if err != nil {
		return time.Duration(0), err
	}

	if res.IsStale() {
		return time.Duration(0), nil
	}

	if hFresh := res.HeuristicFreshness(); hFresh > maxAge {
		debugf("using heuristic freshness of %q", hFresh)
		maxAge = hFresh
	}

	return maxAge - age, nil
}

func (ch *Middleware) isCacheable(res *Resource, r *CacheRequest) bool {
	cc, err := res.cacheControl()
	if err != nil {
		errorf("Error parsing cache-control: %s", err.Error())
		return false
	}

	if cc.Has("no-cache") || cc.Has("no-store") {
		return false
	}

	if cc.Has("private") && len(cc["private"]) == 0 && ch.Shared {
		return false
	}

	if _, ok := storeable[res.Status()]; !ok {
		return false
	}

	if r.Header.Get("Authorization") != "" && ch.Shared {
		return false
	}

	if res.Header().Get("Authorization") != "" && ch.Shared &&
		!cc.Has("must-revalidate") && !cc.Has("s-maxage") {
		return false
	}

	if res.HasExplicitExpiration() {
		return true
	}

	if _, ok := cacheableByDefault[res.Status()]; !ok && !cc.Has("public") {
		return false
	}

	// if res.HasValidators() {
	// 	return true
	// } else if res.HeuristicFreshness() > 0 {
	// 	return true
	// }
	return true
}

// CorrectedAge adjusts the age of a resource for clock skew and travel time
func CorrectedAge(h http.Header, reqTime, respTime time.Time) (time.Duration, error) {
	date, err := timeHeader("Date", h)
	if err != nil {
		return time.Duration(0), err
	}

	apparentAge := respTime.Sub(date)
	if apparentAge < 0 {
		apparentAge = 0
	}

	respDelay := respTime.Sub(reqTime)
	ageSeconds, err := intHeader("Age", h)
	if err != nil {
		return time.Duration(0), err
	}
	age := time.Second * time.Duration(ageSeconds)
	correctedAge := age + respDelay

	if apparentAge > correctedAge {
		correctedAge = apparentAge
	}

	residentTime := Clock().Sub(respTime)
	currentAge := correctedAge + residentTime

	return currentAge, nil
}
