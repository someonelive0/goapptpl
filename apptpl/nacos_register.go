package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3/client"
	log "github.com/sirupsen/logrus"

	"goapptol/utils"
)

/*
 * only support nacos 2.*
 */
const NACOS_LOOP_INTERVAL = 10

type NacosRegister struct {
	Nacosconfig *NacosConfig
	Myconfig    *MyConfig
	localip     string
	shutdown    bool
	chtimeout   chan bool
}

// loop to regist service
func (p *NacosRegister) Regist() error {
	if !p.Nacosconfig.Enable {
		log.Infof("nacos register disabled by config, so not regist to nacos")
		return nil
	}
	p.chtimeout = make(chan bool, 1)

	cc := client.New()
	cc.SetTimeout(time.Duration(p.Nacosconfig.Timeout) * time.Second)
	cc.AddHeader("username", p.Nacosconfig.User)
	cc.AddHeader("password", p.Nacosconfig.Password)

	p.localip, _ = utils.GetOutBoundIP4()

	for {
		if p.shutdown {
			break
		}

		// first login
		url := fmt.Sprintf("%s/nacos/v1/auth/login?username=%s&password=%s",
			p.Nacosconfig.Addr, p.Nacosconfig.User, p.Nacosconfig.Password)
		resp, err := cc.Post(url)
		if err != nil {
			log.Errorf("nacos login failed: %v", err)

			go func() {
				time.Sleep(time.Duration(p.Nacosconfig.Timeout) * time.Second)
				p.chtimeout <- true
			}()
			<-p.chtimeout

			continue
		}
		log.Infof("nacos login resp: %v", resp)
		resp.Close()

		// then regist service
		url = fmt.Sprintf("%s/nacos/v2/ns/instance?serviceName=goapptpl&ip=%s&port=%d",
			p.Nacosconfig.Addr, p.localip, p.Myconfig.Port)
		for {
			if p.shutdown {
				break
			}

			resp, err := cc.Post(url)
			if err != nil {
				log.Errorf("nacos registe service failed: %v", err)
				continue
			}
			log.Tracef("nacos register resp: %v", resp)
			resp.Close()

			go func() {
				time.Sleep(NACOS_LOOP_INTERVAL * time.Second)
				p.chtimeout <- true
			}()
			<-p.chtimeout
		}
	}

	log.Debug("nacos reister stop")
	return nil
}

func (p *NacosRegister) Stop() error {
	p.shutdown = true
	close(p.chtimeout)

	// then delete nacos service
	url := fmt.Sprintf("%s/nacos/v2/ns/instance?serviceName=goapptpl&ip=%s&port=%d",
		p.Nacosconfig.Addr, p.localip, p.Myconfig.Port)
	cc := client.New()
	cc.SetTimeout(time.Duration(p.Nacosconfig.Timeout) * time.Second)
	cc.AddHeader("username", p.Nacosconfig.User)
	cc.AddHeader("password", p.Nacosconfig.Password)
	resp, err := cc.Delete(url)
	if err != nil {
		log.Errorf("nacos unregist service failed: %v", err)
		return err
	}
	log.Tracef("nacos unregist service resp: %v", resp)
	resp.Close()

	return nil
}

func (p *NacosRegister) GetConfig(dataId, group, namespaceId string) (string, error) {
	url := fmt.Sprintf("%s/nacos/v2/cs/config?dataId=%s&group=%s&namespaceId=%s",
		p.Nacosconfig.Addr, dataId, group, namespaceId)
	cc := client.New()
	cc.SetTimeout(time.Duration(p.Nacosconfig.Timeout) * time.Second)
	cc.AddHeader("username", p.Nacosconfig.User)
	cc.AddHeader("password", p.Nacosconfig.Password)
	resp, err := cc.Get(url)
	if err != nil {
		log.Errorf("nacos GetConfig '%s'.'%s' failed: %v", group, dataId, err)
		return "", err
	}
	defer resp.Close()

	m := make(map[string]interface{})
	if err = json.Unmarshal(resp.Body(), &m); err != nil {
		log.Errorf("nacos parse nacos response failed '%s': %v", resp.Body(), err)
		return "", err
	}
	if m["code"].(float64) == 0 {
		if data, ok := m["data"]; ok && data != nil {
			return data.(string), nil
		}
	}

	return "", fmt.Errorf("nacos GetConfig '%s'.'%s': not found ", group, dataId)
}
