package mqtt

import (
	"log"
	"time"

	mqttlib "github.com/eclipse/paho.mqtt.golang"
)

// MessageHandler defines the signature for handling incoming messages
type MessageHandler func(topic string, payload []byte)

// Client defines the interface for MQTT operations
type Client interface {
	Publish(topic string, payload string) error
	Subscribe(topic string, handler MessageHandler) error
}

// mqttClient implements the Client interface
type mqttClient struct {
	client mqttlib.Client
}

// NewClient initializes and connects an MQTT client
func NewClient(brokerURL, clientID, username, password string) Client {
	opts := mqttlib.NewClientOptions().
		AddBroker(brokerURL).
		SetClientID(clientID).
		SetUsername(username).
		SetPassword(password)

	client := mqttlib.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("❌ MQTT connect error: %v", token.Error())
	}

	log.Println("✅ MQTT connected")
	return &mqttClient{client: client}
}

// Publish sends a message to a topic
func (m *mqttClient) Publish(topic string, payload string) error {
	token := m.client.Publish(topic, 1, false, payload)
	token.WaitTimeout(5 * time.Second)
	return token.Error()
}

// Subscribe subscribes to a topic with a custom message handler
func (m *mqttClient) Subscribe(topic string, handler MessageHandler) error {
	token := m.client.Subscribe(topic, 1, func(_ mqttlib.Client, msg mqttlib.Message) {
		handler(msg.Topic(), msg.Payload())
	})
	token.WaitTimeout(5 * time.Second)
	return token.Error()
}
