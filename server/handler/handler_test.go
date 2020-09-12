package handler

import (
	"io/ioutil"
	"log"
	"net"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	h := New()
	assert.NotNil(t, h)
}

func TestStartServer(t *testing.T) {
	h := New()
	li := h.StartServer("3999")
	defer li.Close()

	assert.NotNil(t, li)
}

func TestServeListener(t *testing.T) {
	tests := []struct {
		message             string
		expectedFromChannel string
		terminate           bool
		port                string
	}{
		{
			"123456789",
			"123456789",
			false,
			"4000",
		},
		{
			"terminate",
			"",
			true,
			"4001",
		},
		{
			"wrongInput",
			"",
			false,
			"4002",
		},
	}

	for _, test := range tests {
		h := New()
		li := h.StartServer(test.port)

		go func() {
			conn, err := net.Dial("tcp", "localhost:"+test.port)
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Close()

			_, err = conn.Write([]byte(test.message))
			if err != nil {
				t.Fatal(err)
			}
		}()

		defer li.Close()

		go h.ServeListener(li)

		time.Sleep(10 * time.Millisecond)

		if len(test.expectedFromChannel) > 0 {
			assert.Equal(t, <-h.c, test.expectedFromChannel)
		}
		assert.Equal(t, test.terminate, *h.terminate)
	}
}

func TestTrackInputs(t *testing.T) {
	tests := []struct {
		message                string
		expectedUniqueCount    int
		expectedDuplicateCount int
	}{
		{
			"123456789",
			1,
			0,
		},
		{
			"321654987",
			2,
			0,
		},
		{
			"123456789",
			2,
			1,
		},
	}

	h := New()
	li := h.StartServer("4005")

	go h.ServeListener(li)
	go h.TrackInputs()

	for _, tt := range tests {

		go func() {
			conn, err := net.Dial("tcp", "localhost:4005")
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Close()

			_, err = conn.Write([]byte(tt.message))
			if err != nil {
				t.Fatal(err)
			}
		}()

		defer li.Close()

		time.Sleep(10 * time.Millisecond)
		assert.Equal(t, tt.expectedUniqueCount, h.count.uniqueNumbers)
		assert.Equal(t, tt.expectedDuplicateCount, h.count.duplicateNumbers)
	}
}

func TestSaveResultToFile(t *testing.T) {

	tests := []struct {
		message        string
		expectedResult string
	}{
		{
			"123456789",
			"",
		},
		{
			"321654987",
			"123456789\n",
		},
		{
			"321654989",
			"123456789\n321654987\n",
		},
		{
			"123456789",
			"123456789\n321654987\n321654989\n",
		},
		{
			"",
			"123456789\n321654987\n321654989\n",
		},
	}

	h := New()
	li := h.StartServer("4006")

	file, err := os.Create("./testNumbers.log")
	if err != nil {
		log.Fatalln(err)
	}

	for _, test := range tests {

		go func() {
			conn, err := net.Dial("tcp", "localhost:4006")
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Close()

			_, err = conn.Write([]byte(test.message))
			if err != nil {
				t.Fatal(err)
			}
		}()

		defer li.Close()

		go h.ServeListener(li)
		go h.TrackInputs()
		h.SaveResultToFile(file)

		time.Sleep(100 * time.Millisecond)
		dat, err := ioutil.ReadFile("./testNumbers.log")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, test.expectedResult, string(dat))
	}

	err = os.Remove("./testNumbers.log")
	if err != nil {
		log.Fatal(err)
	}
}

func TestValidateInput(t *testing.T) {
	tests := []struct {
		message        string
		expectedResult bool
	}{
		{
			"123456789",
			true,
		},
		{
			"3216549872",
			false,
		},
		{
			"test",
			false,
		},
		{
			"terminate",
			true,
		},
	}

	for _, test := range tests {
		res := validateInput(test.message)
		assert.Equal(t, test.expectedResult, res)
	}
}
