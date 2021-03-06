// proxy project proxy.go
package proxy

import (
	"io"
	"net"
	"net/http"
	"strings"
)

var proxyHeaders = []string{
	"Connection",
	"Proxy-Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te",
	"Trailer",
	"Transfer-Encoding",
	"Upgrade",
}

type ProxySvr struct {
	Trans *http.Transport
}

//重写ServerHttp接口
func (this *ProxySvr) ServerHttp(rw http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "CONNECT":

	case "GET":

	default:
		this.ProxyHttpHandler(rw, req)
	}
}

//转发http
func (this *ProxySvr) ProxyHttpHandler(rw http.ResponseWriter, req *http.Request) {
	this.DelHeads(req)
	addXForwardIpToHead(req)

	res, err := this.Trans.RoundTrip(req)
	if err != nil {
		rw.WriteHeader(http.StatusBadGateway)
		return
	}
	defer res.Body.Close()

	this.RewriteHead(rw.Header(), res.Header)
	rw.WriteHeader(res.StatusCode)
	_, err = io.Copy(rw, res.Body)
	if err != nil {
		if err != io.EOF {
			rw.WriteHeader(http.StatusBadGateway)
			return
		}
	}

}

//添加X-Forwarded-For
func addXForwardIpToHead(req *http.Request) {
	if ip, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
		if proxyip, ok := req.Header["X-Forwarded-For"]; ok {
			ip = strings.Join(proxyip, ", ") + ", " + ip
		}
		req.Header.Set("X-Forwarded-For", ip)
	}
}

//重写头部
func (this *ProxySvr) RewriteHead(dst, src http.Header) {
	for dstkey, _ := range dst {
		dst.Del(dstkey)
	}
	for srckey, srcv := range src {
		for _, v := range srcv {
			dst.Add(srckey, v)
		}
	}
}

//删除hop-to-hop头部
func (this *ProxySvr) DelHeads(req *http.Request) {
	for _, head := range proxyHeaders {
		if req.Header.Get(head) != "" {
			req.Header.Del(head)
		}
	}
}
