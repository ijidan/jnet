package jnet

import (
	uuid "github.com/satori/go.uuid"
	"golang.org/x/net/publicsuffix"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

const HttpProxy = "http://127.0.0.1:8888" //代理
const MethodGet = "GET"                   //GET请求
const MethodPost = "POST"                 //POST请求
const MethodPut = "PUT"                   //PUT请求
const MethodDelete = "DELETE"             //DELETE请求

//请求类
type Request struct {
	useProxy  bool
	uuid      string
	url       string
	params    map[string]interface{}
	method    string
	startTime int64
	endTime   int64

	client     *http.Client  //请求client
	Req        *http.Request //请求类
	GCookies   []*http.Cookie
	GCookieJar *cookiejar.Jar
}

//设置是否使用代理
func (r *Request) SetUseProxy(useProxy bool) {
	r.useProxy = useProxy
}

//获取是否使用代理
func (r *Request) GetUseProxy() bool {
	return r.useProxy
}

//设置请求方法
func (r *Request) SetMethod(method string) {
	if method == "" {
		method = MethodGet
	}
	r.method = method
}

//获取请求方法
func (r *Request) GetMethod() string {
	return r.method
}

//设置URL
func (r *Request) SetUrl(url string) {
	r.url = url
}

//获取URL
func (r *Request) GetUrl() string {
	return r.url
}

//设置参数
func (r *Request) SetParams(params map[string]interface{}) {
	if r.method == http.MethodGet {
		params["uuid"] = r.uuid
	}
	r.params = params
}

//获取参数
func (r *Request) GetParams() map[string]interface{} {
	return r.params
}

//设置UUID
func (r *Request) SetUUID(_uuid string) {
	if _uuid == "" {
		_uuidV4 := uuid.NewV4()
		_uuid = _uuidV4.String()
	}
	r.uuid = _uuid
}

//获取UUID
func (r *Request) GetUUID() string {
	return r.uuid
}

//发送请求
func (r *Request) send() Response {
	if r.url == "" {
		return NewResponse(ReqFail, "request need host", "", "")
	}
	r.startTime = r.GetMillisecond()
	var encodeQ string
	var req *http.Request
	if r.method == MethodGet {
		req, _ = http.NewRequest(r.method, r.url, nil)
		encodeQ = r.buildGetEncodeQ(req)
		req.URL.RawQuery = encodeQ
	} else {
		encodeQ = r.buildPostEncodeQ()
		reqBody := strings.NewReader(encodeQ)
		req, _ = http.NewRequest(r.method, r.url, reqBody)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	}
	//处理cookie
	gCookieCnt := len(r.GCookies)
	if gCookieCnt > 0 {
		for idx := 0; idx < gCookieCnt; idx++ {
			currCookie := r.GCookies[idx]
			req.AddCookie(currCookie)
		}
	}
	r.Req = req
	rsp := r.computeResponse()
	return rsp
}

//构造代理
func (r *Request) buildProxy() func(*http.Request) (*url.URL, error) {
	proxy := func(*http.Request) (*url.URL, error) {
		return url.Parse(HttpProxy)
	}
	return proxy
}

//构造GET请求参数
func (r *Request) buildGetEncodeQ(req *http.Request) string {
	paramList := r.convertParams2String()
	var qEncode string
	if len(paramList) > 0 {
		q := req.URL.Query()
		for k, v := range paramList {
			q.Add(k, v)
		}
		qEncode = q.Encode()
	}
	return qEncode
}

//构造POST请求参数
func (r *Request) buildPostEncodeQ() string {
	paramList := r.convertParams2String()
	var qEncode string
	if len(paramList) > 0 {
		q := url.Values{}
		for k, v := range paramList {
			q.Set(k, v)
		}
		qEncode = q.Encode()
	}
	return qEncode

}

//参数转字符串
func (r *Request) convertParams2String() map[string]string {
	paramList := r.convertValue2String(r.params)
	return paramList
}

//其他类型转化为字符串
func (r *Request) convertValue2String(params map[string]interface{}) map[string]string {
	paramList := make(map[string]string)
	if len(params) > 0 {
		for k, v := range params {
			value := r.convertVal2String(v)
			paramList[k] = value
		}
	}
	return paramList
}

//变量转化为字符串
func (r *Request) convertVal2String(v interface{}) string {
	var value string
	t := reflect.TypeOf(v)
	switch t.Kind() {
	case reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64:
		_v := reflect.ValueOf(v).Int()
		value = strconv.FormatInt(_v, 10)
		break
	default:
		value = v.(string)
	}
	return value

}

//构造client
func (r *Request) buildClient() {
	if r.useProxy == true {
		proxy := r.buildProxy()
		httpTransport := &http.Transport{
			Proxy: proxy,
		}
		r.client = &http.Client{Transport: httpTransport, Jar: r.GCookieJar}
	} else {
		r.client = &http.Client{Jar: r.GCookieJar}
	}
}

//响应
func (r *Request) computeResponse() Response {
	r.buildClient()
	resp, err := r.client.Do(r.Req)
	if err != nil {
		return NewResponse(ReqFail, err.Error(), "", "")
	}
	respStatusCode := resp.StatusCode
	if respStatusCode != HttpStatusCodeSuccess {
		return NewResponse(JsonParseFail, "rsp status code: "+strconv.Itoa(respStatusCode), "", "")
	}
	body := resp.Body
	defer body.Close()
	respBody, _ := ioutil.ReadAll(body)
	if len(respBody) == 0 {
		return NewResponse(ResponseEmpty, "rsp content empty", "", "")
	}
	bodyString := string(respBody)
	return NewResponse(Success, "", bodyString, "")
}

//获取当前毫秒
func (r *Request) GetMillisecond() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

//获取cookie
func (r *Request) GetGCookies() []*http.Cookie {
	return r.GCookies
}

//设置cookie
func (r *Request) SetGCookies(cookies []*http.Cookie) {
	r.GCookies = cookies
}

//获取cookie jar
func (r *Request) GetGCookieJar() *cookiejar.Jar {
	return r.GCookieJar
}

//实例
var instance *Request
var once sync.Once

//获取单例
func NewRequest() *Request {
	once.Do(func() {
		instance = &Request{}
		instance.GCookies = nil
		_cookieJar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
		instance.GCookieJar = _cookieJar
	})
	return instance
}
