package ndn

import (
	"bufio"
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"testing"
)

func BenchmarkDataEncodeRsa(b *testing.B) {
	b.StopTimer()
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		b.Fatal(err)
	}
	SignKey, err = NewKey("/testing/key", rsaKey)
	if err != nil {
		b.Fatal(err)
	}

	packet := NewData("/testing/ndn")
	packet.SignatureInfo.SignatureType = SignatureTypeSha256WithRsa
	buf := new(bytes.Buffer)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		err := packet.WriteTo(buf)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDataEncodeEcdsa(b *testing.B) {
	b.StopTimer()
	ecdsaKey, err := ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
	if err != nil {
		b.Fatal(err)
	}
	SignKey, err = NewKey("/testing/key", ecdsaKey)
	if err != nil {
		b.Fatal(err)
	}

	packet := NewData("/testing/ndn")
	packet.SignatureInfo.SignatureType = SignatureTypeSha256WithEcdsa
	buf := new(bytes.Buffer)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		err := packet.WriteTo(buf)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDataEncode(b *testing.B) {
	b.StopTimer()
	packet := NewData("/testing/ndn")
	buf := new(bytes.Buffer)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		err := packet.WriteTo(buf)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDataDecode(b *testing.B) {
	b.StopTimer()
	packet := NewData("/testing/ndn")
	buf := new(bytes.Buffer)
	for i := 0; i < b.N; i++ {
		packet.WriteTo(buf)
		b.StartTimer()
		err := new(Data).ReadFrom(bufio.NewReader(buf))
		if err != nil {
			b.Fatal(err)
		}
		b.StopTimer()
	}
}

func BenchmarkInterestEncode(b *testing.B) {
	b.StopTimer()
	packet := NewInterest("/testing/ndn")
	buf := new(bytes.Buffer)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		err := packet.WriteTo(buf)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkInterestDecode(b *testing.B) {
	b.StopTimer()
	packet := NewInterest("/testing/ndn")
	buf := new(bytes.Buffer)
	for i := 0; i < b.N; i++ {
		packet.WriteTo(buf)
		b.StartTimer()
		err := new(Interest).ReadFrom(bufio.NewReader(buf))
		if err != nil {
			b.Fatal(err)
		}
		b.StopTimer()
	}
}