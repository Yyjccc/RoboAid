package core

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"net"
	"net/http"
	"time"
)

// MaxIdleConns 默认指定空闲连接池大小
const MaxIdleConns = 10000

var (
	Http *HttpApi
)

type HttpApi struct {
	RestyClient *resty.Client // resty client 复用
}

func init() {

	client := resty.New().
		SetTransport(CreateTransport(nil, MaxIdleConns)). // 自定义 transport
		SetLogger(DefaultLogger).
		SetHeader("User-Agent", "golang-sdk") //.
	Http = &HttpApi{
		RestyClient: client,
	}

}

func (o *HttpApi) Transport(ctx context.Context, method, url string, body interface{}) ([]byte, error) {
	resp, err := o.Request(ctx).SetBody(body).Execute(method, url)
	return resp.Body(), err
}

// request 每个请求，都需要创建一个 request
func (o *HttpApi) Request(ctx context.Context) *resty.Request {
	return o.RestyClient.R().SetContext(ctx)
}

func RespInfo(resp *resty.Response) string {
	bodyJSON, _ := json.Marshal(resp.Request.Body)
	return fmt.Sprintf(
		resp.Status(),
		resp.Time(),
		string(bodyJSON),
		string(resp.Body()),
	)
}

func CreateTransport(localAddr net.Addr, idleConns int) *http.Transport {
	dialer := &net.Dialer{
		Timeout:   60 * time.Second,
		KeepAlive: 60 * time.Second,
	}
	if localAddr != nil {
		dialer.LocalAddr = localAddr
	}
	return &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          idleConns,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConnsPerHost:   idleConns,
		MaxConnsPerHost:       idleConns,
	}
}
