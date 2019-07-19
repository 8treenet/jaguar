package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/8treenet/jaguar/chat_examples/mock_client/client"
	//"github.com/8treenet/jaguar/chat_examples/chat_examples/client"
)

var _client *client.TcpConn

func main() {
	rand.Seed(time.Now().UnixNano())
	_client = client.NewMockConn("127.0.0.1:9000", readcall)
	go login()
	_client.Start()
}

func readcall(id uint16, buffer *bytes.Buffer) {
	switch id {
	case 100:
		loginReceipt(buffer)
	case 300:
		newInformation(buffer)
	}
}

func login() {
	time.Sleep(1 * time.Second)
	packet := new(bytes.Buffer)
	id := make([]byte, 2)
	binary.BigEndian.PutUint16(id, 100)

	token := fmt.Sprint(1000 + rand.Intn(8000))
	tlen := byte(len(token))

	packet.Write(id)
	packet.WriteByte(tlen)
	packet.WriteString(token)
	_client.Send(packet.Bytes())
}

func loginReceipt(buffer *bytes.Buffer) {
	success, _ := buffer.ReadByte()
	if int(success) != 1 {
		fmt.Println("Exit")
		os.Exit(-1)
	}
	var uid uint32
	binary.Read(buffer, binary.BigEndian, &uid)
	namen, _ := buffer.ReadByte()
	username := string(buffer.Next(int(namen)))
	fmt.Println(fmt.Sprintf("Access success user_id:%d, user_name: %s", uid, username))
	go chatterbox()
}

func newInformation(buffer *bytes.Buffer) {
	var row uint32
	binary.Read(buffer, binary.BigEndian, &row)
	var sendersize uint8
	binary.Read(buffer, binary.BigEndian, &sendersize)
	sender := string(buffer.Next(int(sendersize)))
	var contentsize uint16
	binary.Read(buffer, binary.BigEndian, &contentsize)
	content := string(buffer.Next(int(contentsize)))
	fmt.Println(fmt.Sprintf("sender :%s, row:%d content:%s", sender, row, content))
}

func chatterbox() {
	msgs := []string{
		"I’m so sorry to have bothered you.",
		"You look tiredHad a big night?",
		"Did you sleep soundly last night?",
		"I haven’t seen you in years.",
		"What’s your nationality?",
		"I hope we will become good friends.",
		"Where did you go over the weekend?",
	}
	index := uint32(1)
	for {
		packet := new(bytes.Buffer)
		id := make([]byte, 2)
		binary.BigEndian.PutUint16(id, 101)
		packet.Write(id)
		row := make([]byte, 4)
		binary.BigEndian.PutUint32(row, index)
		packet.Write(row)
		msg := msgs[rand.Intn(len(msgs))]
		msglen := uint16(len(msg))
		msglenByte := make([]byte, 2)
		binary.BigEndian.PutUint16(msglenByte, msglen)
		packet.Write(msglenByte)
		packet.Write([]byte(msg))

		_client.Send(packet.Bytes())
		time.Sleep(time.Duration(1+rand.Intn(3)) * time.Second)
		index += 1
	}
}
