package handler

import (
	"log"
	"sync"

	"go-sapa-kami/internal/mapper"
	"go-sapa-kami/internal/model"
)

type SSEMessage struct {
	Event string
	Data  interface{}
}

type SSEClient struct {
	UserID      uint
	UserRole    model.RoleEnum
	MessageChan chan SSEMessage
}

type SSEManager struct {
	clients map[uint]map[*SSEClient]bool
	mutex   sync.RWMutex
}

func NewSSEManager() *SSEManager {
	return &SSEManager{clients: make(map[uint]map[*SSEClient]bool)}
}

// Add a New Clients to the Manager
func (m *SSEManager) AddClient(aspirationID, userID uint, userRole model.RoleEnum) *SSEClient {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.clients[aspirationID] == nil {
		m.clients[aspirationID] = make(map[*SSEClient]bool)
	}

	client := &SSEClient{
		UserID:      userID,
		UserRole:    userRole,
		MessageChan: make(chan SSEMessage, 100),
	}
	m.clients[aspirationID][client] = true

	return client
}

// Remove a Client from the Manager
func (m *SSEManager) RemoveClient(aspirationID uint, client *SSEClient) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if room, exists := m.clients[aspirationID]; exists {
		if _, ok := room[client]; ok {
			delete(room, client)
			close(client.MessageChan)

			if len(room) == 0 { delete(m.clients, aspirationID) }
		}
	}
}

// Broadcast a Generic Message to All Clients in an Aspiration
func (m *SSEManager) BroadcastToAspiration(aspirationID uint, eventName string, data interface{}) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	room, exists := m.clients[aspirationID]
	if !exists || len(room) == 0 { return }

	msg := SSEMessage{
		Event: eventName,
		Data:  data,
	}

	for client := range room {
		select {
		case client.MessageChan <- msg:
		default: log.Printf("[SSE] Warning: Dropping message for stuck client in aspiration %d", aspirationID)
		}
	}
}

// Broadcast a Discussion Message Per-Client to Guarantee Privacy & Correct is_you
func (m *SSEManager) BroadcastDiscussion(aspiration *model.Aspiration, discussion *model.Discussion) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	room, exists := m.clients[aspiration.ID]
	if !exists || len(room) == 0 { return }

	for client := range room {
		clientData := mapper.ToDiscussionResponse(aspiration, client.UserID, client.UserRole, discussion)
		msg := SSEMessage{
			Event: "new_discussion",
			Data:  clientData,
		}

		select {
		case client.MessageChan <- msg:
		default: log.Printf("[SSE] Warning: Dropping discussion for stuck client in aspiration %d", aspiration.ID)
		}
	}
}
