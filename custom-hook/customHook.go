package hook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"message-core/pkg/xservice/platform"
	"message-core/websocket"

	"github.com/mochi-co/mqtt/v2"
	"github.com/mochi-co/mqtt/v2/packets"
	"github.com/spf13/cast"
)

// configuration for the broker
const (
	ACLWriteonly = "writeonly"
)

type CustomHook struct {
	mqtt.HookBase
}

func (h *CustomHook) ID() string {
	return "custom-hook-acl-auth"
}

func (h *CustomHook) Provides(b byte) bool {
	return bytes.Contains([]byte{
		mqtt.OnConnect,
		mqtt.OnDisconnect,
		mqtt.OnSubscribed,
		mqtt.OnUnsubscribed,
		mqtt.OnPublished,
		mqtt.OnPublish,
		mqtt.OnACLCheck,
		mqtt.OnConnectAuthenticate,
	}, []byte{b})
}

func (h *CustomHook) Init(config any) error {
	h.Log.Info().Msg("initialised")
	return nil
}

func (h *CustomHook) OnConnectAuthenticate(cl *mqtt.Client, pk packets.Packet) bool {
	return true
}

func (h *CustomHook) OnConnect(cl *mqtt.Client, pk packets.Packet) error {
	err := h.TopicVerifyConnect(cl, pk)
	if err != nil {
		h.Log.Error().Err(err).
			Str("username", string(pk.Connect.Username)).
			Str("password", string(pk.Connect.Password)).
			Str("client", cl.ID).
			Msg("Client disconnected")
		return err
	}
	h.Log.Info().
		Str("username", string(pk.Connect.Username)).
		Str("password", string(pk.Connect.Password)).
		Str("client", cl.ID).
		Msg("Client connected")
	return nil
}

func (h *CustomHook) OnACLCheck(cl *mqtt.Client, topic string, write bool) bool {
	ACL, err := h.TopicVerifyACL(cl, topic)
	if err != nil {
		h.Log.Error().
			Err(err).
			Str("client", cl.ID).
			Str("topic", topic).
			Msg("OnACLCheck Error")
		return false
	}
	if ACL == ACLWriteonly {
		if !write {
			h.Log.Error().
				Err(err).
				Str("client", cl.ID).
				Str("topic", topic).
				Msg("Deny message subscribe because topic writeonly")
			return false
		}
		return true
	}

	return true
}

func (h *CustomHook) OnDisconnect(cl *mqtt.Client, err error, expire bool) {
	h.Log.Info().Str("client", cl.ID).Bool("expire", expire).Err(err).Msg("client disconnected")
}

func (h *CustomHook) OnSubscribed(cl *mqtt.Client, pk packets.Packet, reasonCodes []byte) {
	h.Log.Info().Str("client", cl.ID).Interface("filters", pk.Filters).Msgf("subscribed qos=%v", reasonCodes)
}

func (h *CustomHook) OnUnsubscribed(cl *mqtt.Client, pk packets.Packet) {
	h.Log.Info().Str("client", cl.ID).Interface("filters", pk.Filters).Msg("unsubscribed")
}

func (h *CustomHook) TopicVerifyConnect(cl *mqtt.Client, pk packets.Packet) (err error) {
	err = platform.ValidationUser(
		context.Background(),
		platform.GatewayValidationRequest{
			UserName: string(pk.Connect.Username),
			Password: string(pk.Connect.Password),
		})
	if err != nil {
		return err
	}

	return
}

func (h *CustomHook) TopicVerifyACL(cl *mqtt.Client, topic string) (ACL string, err error) {

	topicActual, ACL, err := SplitTopicACL(topic)
	if err != nil {
		return
	}

	if string(cl.Properties.Username) != topicActual {
		err = fmt.Errorf("Not meet username and topic: %s <> %s",
			string(cl.Properties.Username),
			topicActual)
		return
	}

	return
}

func (h *CustomHook) OnPublish(cl *mqtt.Client, pk packets.Packet) (packets.Packet, error) {
	topicActual, _, err := SplitTopicACL(pk.TopicName)
	if err != nil {
		return packets.Packet{}, nil
	}
	websocket.GetServerConn().Publish(topicActual, pk.Payload)

	npk := h.ApplyRuleForPacket(pk, pk.TopicName)
	// apply rules here.
	// first - check username to get the topic name.
	// second - get rule from redis if exists.
	// finnal - modify the message if rule exists.

	return npk, nil
}

func (h *CustomHook) OnPublished(cl *mqtt.Client, pk packets.Packet) {
	// send to websocket server.
	h.Log.Info().Str("client", cl.ID).Str("payload", string(pk.Payload)).Msg("published to client")
}

func (h *CustomHook) ApplyRuleForPacket(pk packets.Packet, userName string) (npk packets.Packet) {
	dataPacket := make(map[string]interface{})
	err := json.Unmarshal(pk.Payload, &dataPacket)
	if err != nil {
		// if error when unmarshalling => return original packet
		return pk
	}

	dataRules, err := platform.GetRuleCache(context.Background(), userName)
	if err != nil {
		// if error when unmarshalling => return original packet
		return pk
	}

	if len(dataRules.Rules) == 0 {
		return pk
	}

	mapRule := make(map[string]platform.RulesDevices)
	for _, rule := range dataRules.Rules {
		mapRule[rule.Atribute] = rule
	}

	for atribute, value := range dataPacket {
		if _, exist := mapRule[atribute]; exist {
			switch mapRule[atribute].Comparison {
			case "EQUAL":
				if cast.ToFloat32(value) != cast.ToFloat32(mapRule[atribute].RuleValue) {
					return packets.Packet{}
				}
				continue
			case "NOT EQUAL":
				if cast.ToFloat32(value) == cast.ToFloat32(mapRule[atribute].RuleValue) {
					return packets.Packet{}
				}
				continue
			case "GREATER THAN":
				if cast.ToFloat32(value) < cast.ToFloat32(mapRule[atribute].RuleValue) {
					return packets.Packet{}
				}
				continue
			case "LESS THAN":
				if cast.ToFloat32(value) > cast.ToFloat32(mapRule[atribute].RuleValue) {
					return packets.Packet{}
				}
				continue
			}
		}

	}
	return pk
}

// const options = ["EQUAL", "NOT EQUAL", "GREATER THAN", "LESS THAN"];
