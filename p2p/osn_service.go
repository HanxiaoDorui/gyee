/*
 *  Copyright (C) 2017 gyee authors
 *
 *  This file is part of the gyee library.
 *
 *  The gyee library is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  The gyee library is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *  GNU General Public License for more details.
 *
 *  You should have received a copy of the GNU General Public License
 *  along with the gyee library.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

/*
 overlay sub-network supported p2p network
*/

package p2p

import (
	"github.com/pkg/errors"
	"github.com/yeeco/gyee/p2p/config"
)

type OsnService struct {
	yeShCfg		YeShellConfig
	yeShMgr		Service
}

func OsnServiceConfig(cfg *YeShellConfig, cfgFromFie interface{}) error {
	//
	// 在本函数进行服务参数配置。一般而言，可先从配置文件中得到P2P的各项配置，具体的结构由应用决定并以
	// "cfgFromFile"为参数；其次取P2P中缺省的配置DefaultYeShellConfig之后进行适当的修改，以参数
	// "cfg"指定，调用本函数。后面是对目标配置中各个参数的说明，本函数由应用根据具体情况（cfgFromFie
	// 的结构设计）实现。完成对目标配置的设置之后，即可调用NewOsnService生成一个服务实例。
	//
	// 名称					类型					描述
	//-------------------------------------------------------------------------------------
	// AppType				config.P2pAppType	P2P包含两大部分，peer和dht，是可以单独
	//											使用的。而对我们项目的具体应用，是必须两个
	//											一起使用；
	//
	// Name					string				本次运行的P2P实例，即由函数NewOsnService
	//											生成的对象的名称；
	//
	// Validator			bool				是否为验证器；
	//
	// BootstrapNode		bool				是否为bootstrap节点，目前peer和dht的bootstrap
	//											节点身份不加区分，即一个节点要么同时是这两部分的
	//											bootstrap节点，要么都不是；
	//
	// BootstrapNodes		[]string			peer部分的bootstrap节点列表；
	//
	// DhtBootstrapNodes	[]string			dht部分的bootstrap节点列表；
	//
	// LocalNodeIp			string				本地peer部分的IP地址
	//
	// LocalUdpPort			uint16				本地peer部分的UDP端口
	//
	// LocalTcpPort			uint16				本地peer部分的TCP端口
	//
	// LocalDhtIp			string				本地dht部分的IP地址
	//
	// LocalDhtPort			uint16				本地dht部分的TCP端口
	//
	// NodeDataDir			string				本次实例的数据目录；
	//
	// NodeDatabase			string				本次实例的leveldb数据库名称（数据库所在目录）；
	//
	// SubNetMaskBits		int					子网所使用的掩码的比特数，0-15；
	//
	// EvKeepTime			time.Duration		event在dht中保留的时长；
	//
	// DedupTime			time.Duration		去重时钟管理器进行清理的周期时长；
	//
	// BootstrapTime		time.Duration		dht盲连接周期。在本实例不是bootstrap节点的情况下，
	//											周期性的从BootstrapNodes中随机挑选节点进行连接，
	//											连接成功之后，该时钟停止。
	//
	// 注：如前所述，本函数应由应用根据具体情况（cfgFromFie的结构设计）实现并调用，但这不是必须的，应用
	// 可以用任何方法构造合理的YeShellConfig结构，然后调用NewOsnService得到服务实例。
	//

	_ = cfg
	_ = cfgFromFie

	return nil
}

var _ = OsnServiceConfig

func NewOsnService(cfg *YeShellConfig)(*OsnService, error) {
	osns := OsnService{
		yeShCfg: *cfg,
	}
	if osns.yeShMgr = NewYeShellManager(&osns.yeShCfg); osns.yeShMgr == nil {
		return nil, errors.New("NewOsnService: NewYeShellManager failed")
	}
	return &osns, nil
}

func (osns *OsnService)Start() error {
	return osns.yeShMgr.Start()
}

func (osns *OsnService)Stop() {
	osns.yeShMgr.Stop()
}

func (osns *OsnService)Reconfig(reCfg *RecfgCommand) error {
	return osns.yeShMgr.Reconfig(reCfg)
}

func (osns *OsnService)BroadcastMessage(message Message) error {
	return osns.yeShMgr.BroadcastMessage(message)
}

func (osns *OsnService)BroadcastMessageOsn(message Message) error {
	return osns.yeShMgr.BroadcastMessageOsn(message)
}

func (osns *OsnService)Register(subscriber *Subscriber) {
	osns.yeShMgr.Register(subscriber)
}

func (osns *OsnService)UnRegister(subscriber *Subscriber) {
	osns.yeShMgr.UnRegister(subscriber)
}

func (osns *OsnService)DhtGetValue(key []byte) ([]byte, error) {
	return osns.yeShMgr.DhtGetValue(key)
}

func (osns *OsnService)DhtSetValue(key []byte, value []byte) error {
	return osns.yeShMgr.DhtSetValue(key, value)
}

func (osns *OsnService)GetLocalNode() *config.Node {
	return osns.yeShMgr.(*YeShellManager).GetLocalNode()
}

func (osns *OsnService)GetLocalDhtNode() *config.Node {
	return osns.yeShMgr.(*YeShellManager).GetLocalDhtNode()
}
