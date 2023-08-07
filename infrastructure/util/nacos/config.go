package nacos

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type api struct {
	loginApi  string
	getApi    string
	putApi    string
	listenApi string
}

const (
	contentType      = "application/x-www-form-urlencoded;charset=utf-8"
	splitConfig      = string(rune(1))
	splitConfigInner = string(rune(2))
)

type NacosConfig struct {
	HttpClient  *http.Client
	Endpoint    string
	ServerAddr  string
	accessToken string
	tokenTTL    int
	Username    string
	Password    string
	Logger      logger
	PollTime    time.Duration
	api         *api
}

type LoginResponse struct {
	AccessToken string `json:"accessToken"`
	TokenTTL    int    `json:"tokenTtl"`
	GlobalAdmin bool   `json:"globalAdmin"`
}

func NewNacosConfig(options ...func(c *NacosConfig)) *NacosConfig {
	nc := &NacosConfig{
		HttpClient: http.DefaultClient,
		Logger:     &defualtLogger{},
		PollTime:   10 * time.Second,
	}

	for _, option := range options {
		option(nc)
	}

	if nc.Username == "" {
		panic("Username empty")
	}

	if nc.Username != "" && nc.Password != "" {
		nc.api = &api{
			loginApi:  "/nacos/v1/auth/login",
			getApi:    "/nacos/v1/cs/configs",
			putApi:    "/nacos/v1/cs/configs",
			listenApi: "/nacos/v1/cs/configs/listener",
		}
		if err := nc.login(); err != nil {
			panic(err)
		}
	}

	return nc
}

func (n *NacosConfig) login() error {
	n.Logger.Info(fmt.Sprintf("nacos login server:[%s:%s]", n.ServerAddr, n.Username))

	v := url.Values{}
	v.Add("username", n.Username)
	v.Add("password", n.Password)

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", n.ServerAddr, n.api.loginApi), strings.NewReader(v.Encode()))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", contentType)

	resp, err := n.HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bb, err := io.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("nacos login fail:%s", string(bb))
	}

	loginResp := &LoginResponse{}

	if err := json.Unmarshal(bb, loginResp); err != nil {
		return err
	}
	n.accessToken = loginResp.AccessToken
	n.tokenTTL = loginResp.TokenTTL - 600

	return nil
}

func (n *NacosConfig) getServer() error {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/diamond-server/diamond", n.Endpoint), nil)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", contentType)

	resp, err := n.HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bb, err := io.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	if resp.StatusCode != 200 || len(bb) == 0 {
		return fmt.Errorf("acm get server fail:%s", string(bb))
	}

	addrs := strings.Split(string(bb), "\n")

	n.ServerAddr = fmt.Sprintf("http://%s:8080", strings.TrimSpace(addrs[0]))
	return nil
}

func (n *NacosConfig) Get(namespace, group, dataId string) (string, error) {
	n.Logger.Info(fmt.Sprintf("nacos get config:[namespace:%s,group:%s,dataId:%s]", namespace, group, dataId))

	v := url.Values{}
	v.Add("tenant", namespace)
	v.Add("group", group)
	v.Add("dataId", dataId)
	if n.accessToken != "" {
		v.Add("accessToken", n.accessToken)
	}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s?", n.ServerAddr, n.api.getApi)+v.Encode(), nil)
	if err != nil {
		return "", err
	}

	timeStamp := strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
	req.Header.Add("Content-Type", contentType)
	req.Header.Add("timeStamp", timeStamp)

	resp, err := n.HttpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bb, err := io.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("nacos get fail:%s", string(bb))
	}

	return string(bb), nil
}

func (n *NacosConfig) ListenAsync(namespace, group, dataId string, fn func(cnf string)) {
	ret, err := n.Get(namespace, group, dataId)
	if err != nil {
		panic(err)
	}

	contentMd5 := md5string(ret)

	go func() {
		t1 := time.NewTicker(time.Duration(n.tokenTTL) * time.Second)
		t2 := time.NewTicker(n.PollTime)
		for {
			select {
			// token到期刷新
			case <-t1.C:
				if n.Username != "" {
					if err := n.login(); err != nil {
						n.Logger.Error(err)
					}
				}
			// 每10秒监听配置
			case <-t2.C:
				update, err := n.Listen(namespace, group, dataId, contentMd5)
				if err != nil {
					n.Logger.Error(err)
					continue
				}
				if update {
					ret, err := n.Get(namespace, group, dataId)
					if err != nil {
						n.Logger.Error(err)
						continue
					}

					contentMd5 = md5string(ret)
					n.Logger.Info(fmt.Sprintf("nacos listen refresh:[namespace:%s,group:%s,dataId:%s,md5:%s]", namespace, group, dataId, contentMd5))
					fn(ret)
				}
			}
		}
	}()
}

func (n *NacosConfig) Listen(namespace, group, dataId, md5 string) (bool, error) {
	n.Logger.Info(fmt.Sprintf("nacos listen start:[namespace:%s,group:%s,dataId:%s,md5:%s]", namespace, group, dataId, md5))

	content := dataId + splitConfigInner + group + splitConfigInner + md5 + splitConfigInner + namespace + splitConfig

	v := url.Values{}
	if n.Username != "" {
		v.Add("Listening-Configs", content)
	} else {
		v.Add("Probe-Modify-Request", content)
	}
	v.Add("tenant", namespace)
	if n.accessToken != "" {
		v.Add("accessToken", n.accessToken)
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", n.ServerAddr, n.api.listenApi), strings.NewReader(v.Encode()))
	if err != nil {
		return false, err
	}

	timeStamp := strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
	req.Header.Add("Long-Pulling-Timeout", "3000")
	req.Header.Add("User-Agent", "Nacos-go-client/v1.0.1")
	req.Header.Add("exConfigInfo", "true")
	req.Header.Add("Content-Type", contentType)
	req.Header.Add("timeStamp", timeStamp)

	resp, err := n.HttpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	bb, err := io.ReadAll(resp.Body)

	if err != nil {
		return false, err
	}

	if resp.StatusCode != 200 {
		return false, fmt.Errorf("nacos listen response error:%s", string(bb))
	}

	str := strings.Split(string(bb), "%02")

	// 如果返回数据不为空则代表有变化的文件
	if len(str) > 0 && str[0] == dataId {
		return true, nil
	}

	return false, nil
}

func md5string(text string) string {
	algorithm := md5.New()
	algorithm.Write([]byte(text))
	return hex.EncodeToString(algorithm.Sum(nil))
}
