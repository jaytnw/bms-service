package mqtt

import (
	"log"
	"sync"
	"time"

	mqttlib "github.com/eclipse/paho.mqtt.golang"
)

// MessageHandler defines the signature for handling incoming messages
type MessageHandler func(topic string, payload []byte, retained bool)

// Client defines the interface for MQTT operations
type Client interface {
	Publish(topic string, payload string) error
	Subscribe(topic string, handler MessageHandler) error
}

// mqttClient implements the Client interface
type mqttClient struct {
	client        mqttlib.Client
	subscriptions map[string]MessageHandler
	mu            sync.Mutex
}

// NewClient initializes and connects an MQTT client with reconnect support
func NewClient(brokerURL, clientID, username, password string) Client {
	m := &mqttClient{
		subscriptions: make(map[string]MessageHandler),
	}
	opts := mqttlib.NewClientOptions().
		AddBroker(brokerURL).
		SetClientID(clientID).
		SetUsername(username).
		SetPassword(password).
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetKeepAlive(30 * time.Second).
		SetPingTimeout(10 * time.Second).
		SetConnectRetryInterval(3 * time.Second).
		SetConnectionLostHandler(func(c mqttlib.Client, err error) {
			log.Printf("⚠️ MQTT connection lost: %v", err)
		}).
		SetOnConnectHandler(func(c mqttlib.Client) {
			log.Println("🔁 MQTT reconnected")
			m.resubscribeAll()
		})

	m.client = mqttlib.NewClient(opts)
	if token := m.client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("❌ MQTT connect error: %v", token.Error())
	}

	log.Println("✅ MQTT connected")
	return m
}

// Publish sends a message to a topic
func (m *mqttClient) Publish(topic string, payload string) error {
	token := m.client.Publish(topic, 1, false, payload)
	token.WaitTimeout(5 * time.Second)
	return token.Error()
}

// Subscribe subscribes to a topic and remembers the handler for resubscription
func (m *mqttClient) Subscribe(topic string, handler MessageHandler) error {
	m.mu.Lock()
	m.subscriptions[topic] = handler
	m.mu.Unlock()

	token := m.client.Subscribe(topic, 1, func(_ mqttlib.Client, msg mqttlib.Message) {
		handler(msg.Topic(), msg.Payload(), msg.Retained())
	})
	token.WaitTimeout(5 * time.Second)
	return token.Error()
}

// resubscribeAll resubscribes to all previous topics on reconnect
func (m *mqttClient) resubscribeAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for topic, handler := range m.subscriptions {
		log.Printf("🔁 Resubscribing to topic: %s", topic)
		token := m.client.Subscribe(topic, 1, func(_ mqttlib.Client, msg mqttlib.Message) {
			handler(msg.Topic(), msg.Payload(), msg.Retained())
		})

		token.Wait()
		if err := token.Error(); err != nil {
			log.Printf("❌ Failed to resubscribe to %s: %v", topic, err)
		}
	}
}
