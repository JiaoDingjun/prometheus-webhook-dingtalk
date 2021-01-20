package notifier

import (
    "github.com/timonwong/prometheus-webhook-dingtalk/pkg/models"
    "fmt"
)

var modifierList modifiers
var prometheusSvr string

const blockHeightAlertName = "区块高度相差较大"

func init() {
	regModifier(NewBlockHeightAlarmModifier())
}

func SetPrometheusSvr(svr string) {
	prometheusSvr = svr
}

func regModifier(modifier IModifier) {
	modifierList = append(modifierList, modifier)
}

type IModifier interface {
	Modify(m *models.WebhookMessage) error
}

type modifiers []IModifier

func (mod modifiers) Do(m *models.WebhookMessage) {
	for _, modifier := range mod {
		modifier.Modify(m)
	}
}

type BlockHeightAlarmModifier struct {
}

func NewBlockHeightAlarmModifier() *BlockHeightAlarmModifier {
	modifier := &BlockHeightAlarmModifier{}
	return modifier
}

func (modifier *BlockHeightAlarmModifier) Modify(m *models.WebhookMessage) error {
	//m.ExtraInfo = "测试附加信息\nttttttttt测试"
	if m.Status != "firing" {
		return nil
	}

	for i, alert := range m.Alerts {
		if alert.Status != "firing" {
			continue
		}
		if name, ok := alert.Labels["alertname"]; !ok || name != blockHeightAlertName {
			continue
		}

		chainID, ok := alert.Labels["chainID"]
		if !ok {
			chainID, ok = alert.Labels["chainid"]
		}
		if !ok {
			continue
		}

		info, err := modifier.getBlockHeightInfo(chainID)
        fmt.Printf("getBlockHeightInfo,info:%s,err:%v\n",info,err)
		if err != nil {
			continue
		}

		m.Alerts[i].Labels["blockHeightInfo"] = info
	}
    return nil
}

func (modifier *BlockHeightAlarmModifier) getBlockHeightInfo(chainID string) (string, error) {
	return NewChainMonitor(prometheusSvr).GetChainBlockHeight(chainID)
}
