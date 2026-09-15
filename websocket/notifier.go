// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Publish-subscribe model for nonblocking notification of server events to websocket clients.

package websocket

import (
	"log"
	"sync"
)

// Allow the listeners to buffer a small number of notifications to streamline delivery.
const notifyBufferSize = 5

type Notifier struct {
	messageType     string
	messageProducer func() any
	listeners       map[chan messageEnvelope]bool // True means disconnect the listener when its buffer overflows.
	mutex           sync.Mutex
	sequence        uint64
}

type messageEnvelope struct {
	messageType string
	messageBody any
	sequence    uint64
}

func NewNotifier(messageType string, messageProducer func() any) *Notifier {
	notifier := &Notifier{messageType: messageType, messageProducer: messageProducer}
	notifier.listeners = make(map[chan messageEnvelope]bool)
	return notifier
}

// Calls the messageProducer function and sends a message containing the results to all registered listeners, and cleans
// up any listeners that have closed.
func (notifier *Notifier) Notify() {
	notifier.mutex.Lock()
	defer notifier.mutex.Unlock()
	notifier.notifyWithMessageLocked(notifier.getMessageBody())
}

// Sends the given message to all registered listeners, and cleans up any listeners that have closed. If there is a
// messageProducer function defined it is ignored.
func (notifier *Notifier) NotifyWithMessage(messageBody any) {
	notifier.mutex.Lock()
	defer notifier.mutex.Unlock()
	notifier.notifyWithMessageLocked(messageBody)
}

func (notifier *Notifier) notifyWithMessageLocked(messageBody any) {
	notifier.sequence++
	message := messageEnvelope{messageType: notifier.messageType, messageBody: messageBody, sequence: notifier.sequence}
	for listener, disconnectOnOverflow := range notifier.listeners {
		notifier.notifyListener(listener, message, disconnectOnOverflow)
	}
}

// listenWithSnapshot registers a listener and captures a producer snapshot and sequence under the same lock. This
// guarantees that v1 clients receive either the bootstrap value or a subsequent notification for every state change.
func (notifier *Notifier) listenWithSnapshot() (chan messageEnvelope, any, uint64, bool) {
	notifier.mutex.Lock()
	defer notifier.mutex.Unlock()

	listener := make(chan messageEnvelope, notifyBufferSize)
	notifier.listeners[listener] = true
	if notifier.messageProducer == nil {
		return listener, nil, notifier.sequence, false
	}
	return listener, notifier.messageProducer(), notifier.sequence, true
}

func (notifier *Notifier) notifyListener(listener chan messageEnvelope, message messageEnvelope, disconnectOnOverflow bool) {
	defer func() {
		// If channel is closed sending to it will cause a panic; recover and remove it from the list.
		if r := recover(); r != nil {
			delete(notifier.listeners, listener)
		}
	}()

	// Do a non-blocking send. This guarantees that sending notifications won't interrupt the main event loop,
	// at the risk of clients missing some messages if they don't read them all promptly.
	select {
	case listener <- message:
		// The notification was sent and received successfully.
	default:
		log.Printf("Failed to send a '%s' notification due to blocked listener.", notifier.messageType)
		if disconnectOnOverflow {
			close(listener)
			delete(notifier.listeners, listener)
		}
	}
}

// Registers and returns a channel that can be read from to receive notification messages. The caller is
// responsible for closing the channel, which will cause it to be reaped from the list of listeners.
func (notifier *Notifier) listen() chan messageEnvelope {
	notifier.mutex.Lock()
	defer notifier.mutex.Unlock()

	listener := make(chan messageEnvelope, notifyBufferSize)
	notifier.listeners[listener] = false
	return listener
}

func (notifier *Notifier) stopListening(listener chan messageEnvelope) {
	notifier.mutex.Lock()
	defer notifier.mutex.Unlock()
	if _, ok := notifier.listeners[listener]; ok {
		delete(notifier.listeners, listener)
		close(listener)
	}
}

// Invokes the message producer to get the message, or returns nil if no producer is defined.
func (notifier *Notifier) getMessageBody() any {
	if notifier.messageProducer == nil {
		return nil
	} else {
		return notifier.messageProducer()
	}
}
