package network

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"net"
)

// Encode marshals an interface into a JSON byte array
func Encode(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// Decode unmarshals a JSON byte array into the provided interface
func Decode(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// WriteMessage writes a message to the connection using length-prefixed framing
// It first encodes the message as JSON, then prefixes it with its length as a 4-byte header
func WriteMessage(conn net.Conn, v interface{}) error {
	// Encode the message to JSON
	data, err := Encode(v)
	if err != nil {
		return err
	}

	// Create a buffer for the length prefix (4 bytes)
	lengthBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBuf, uint32(len(data)))

	// Write the length prefix
	_, err = conn.Write(lengthBuf)
	if err != nil {
		return err
	}

	// Write the actual data
	_, err = conn.Write(data)
	return err
}

// ReadMessage reads a message from the connection using length-prefixed framing
// It first reads the 4-byte length header, then reads that many bytes for the JSON payload
func ReadMessage(conn net.Conn, v interface{}) error {
	// Read the length prefix (4 bytes)
	lengthBuf := make([]byte, 4)
	_, err := io.ReadFull(conn, lengthBuf)
	if err != nil {
		return err
	}

	// Parse the length
	messageLength := binary.BigEndian.Uint32(lengthBuf)
	if messageLength > 1024*1024 { // 1MB limit to prevent malicious attacks
		return errors.New("message too large")
	}

	// Read the message payload
	messageBuf := make([]byte, messageLength)
	_, err = io.ReadFull(conn, messageBuf)
	if err != nil {
		return err
	}

	// Decode the message
	return Decode(messageBuf, v)
}
