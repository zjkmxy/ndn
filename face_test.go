package ndn

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net"
	"sync"
	"testing"
)

var (
	testContent = bytes.Repeat([]byte("0123456789"), 100)
)

func TestConsumer(t *testing.T) {
	conn, err := net.Dial("tcp4", "aleph.ndn.ucla.edu:6363")
	if err != nil {
		t.Fatal(err)
	}
	face := NewFace(conn, nil)
	defer face.Close()
	dl, err := face.SendInterest(&Interest{
		Name: NewName("/ndn/edu/ucla"),
	})
	if err != nil {
		t.Fatal(err)
	}
	d, ok := <-dl
	if !ok {
		t.Fatal("timeout")
	}
	t.Logf("name: %v, sig: %v", d.Name, d.SignatureInfo.KeyLocator.Name)
}

func producer(id string) (err error) {
	conn, err := net.Dial("tcp", ":6363")
	if err != nil {
		return
	}
	interestIn := make(chan *Interest)
	face := NewFace(conn, interestIn)
	err = face.Register("/" + id)
	if err != nil {
		face.Close()
		return
	}
	go func() {
		for i := range interestIn {
			face.SendData(&Data{
				Name:    i.Name,
				Content: testContent,
				//MetaInfo: MetaInfo{
				//FreshnessPeriod: 3600000,
				//},
			})
		}
		face.Close()
	}()
	return
}

func consumer(id string, ch chan<- error) {
	conn, err := net.Dial("tcp", ":6363")
	if err != nil {
		ch <- err
		return
	}
	face := NewFace(conn, nil)
	defer face.Close()
	dl, err := face.SendInterest(&Interest{
		Name: NewName("/" + id),
		Selectors: Selectors{
			MustBeFresh: true,
		},
	})
	if err != nil {
		ch <- err
		return
	}
	d, ok := <-dl
	if !ok {
		ch <- fmt.Errorf("timeout %s", face.LocalAddr())
		return
	}
	if d.Name.String() != "/"+id {
		ch <- fmt.Errorf("expected %s, got %s", id, d.Name)
		return
	}
}

func TestProducer(t *testing.T) {
	key, err := ioutil.ReadFile("key/default.pri")
	if err != nil {
		t.Fatal(err)
	}
	err = SignKey.DecodePrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	id := hex.EncodeToString(newNonce())
	err = producer(id)
	if err != nil {
		t.Fatal(err)
	}
	ch := make(chan error)
	go func() {
		consumer(id, ch)
		close(ch)
	}()
	for err := range ch {
		t.Error(err)
	}
}

func BenchmarkForward(b *testing.B) {
	key, err := ioutil.ReadFile("key/default.pri")
	if err != nil {
		b.Fatal(err)
	}
	err = SignKey.DecodePrivateKey(key)
	if err != nil {
		b.Fatal(err)
	}
	var ids []string
	for i := 0; i < 64; i++ {
		ids = append(ids, hex.EncodeToString(newNonce()))
	}
	for _, id := range ids {
		err = producer(id)
		if err != nil {
			b.Fatal(err)
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch := make(chan error)
		var wg sync.WaitGroup
		for _, id := range ids {
			wg.Add(1)
			go func(id string) {
				consumer(id, ch)
				wg.Done()
			}(id)
		}
		go func() {
			wg.Wait()
			close(ch)
		}()
		for err := range ch {
			b.Error(err)
		}
	}
}
