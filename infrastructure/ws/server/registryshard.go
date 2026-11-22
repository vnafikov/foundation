package server

import (
	"context"
	"sync"
)

type registryShard struct {
	connections map[string]map[*Connection]struct{}
	isClosed    bool
	mu          sync.Mutex
}

func (rs *registryShard) addLocked(key string, c *Connection) {
	if rs.connections[key] == nil {
		rs.connections[key] = make(map[*Connection]struct{})
	}
	rs.connections[key][c] = struct{}{}
}

func (rs *registryShard) deleteLocked(key string, c *Connection) {
	delete(rs.connections[key], c)
	if len(rs.connections[key]) == 0 {
		delete(rs.connections, key)
	}
}

func (rs *registryShard) send(ctx context.Context, key string, payload []byte) error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if rs.isClosed {
		return ErrClosed
	}

	connections := rs.connections[key]
	if len(connections) == 0 {
		return nil
	}

	for connection := range connections {
		_ = connection.enqueue(ctx, payload)
	}
	return nil
}

func (rs *registryShard) connectionList(key string) []*Connection {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	connections := rs.connections[key]
	list := make([]*Connection, 0, len(connections))
	for connection := range connections {
		list = append(list, connection)
	}
	return list
}

func (rs *registryShard) disconnect(ctx context.Context, key string) error {
	rs.mu.Lock()
	keyConnections := rs.connections[key]
	connections := make([]*Connection, 0, len(keyConnections))
	for connection := range keyConnections {
		connection.cancelCtx()
		connections = append(connections, connection)
	}
	rs.mu.Unlock()
	for _, connection := range connections {
		select {
		case <-connection.doneChan:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func (rs *registryShard) close() {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	rs.isClosed = true
	for _, connections := range rs.connections {
		for connection := range connections {
			connection.cancelCtx()
		}
	}
}
