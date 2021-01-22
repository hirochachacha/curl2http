package curl2http_test

import (
	"io"
	"io/ioutil"
	"mime/multipart"
	"strings"
	"testing"

	"github.com/hirochachacha/curl2http"
)

func TestRequest(t *testing.T) {
	_, req, err := curl2http.NewRequestFromArgs([]string{"curl", "https://www.example.com/get"})
	if err != nil {
		t.Fatal(err)
	}
	if req.URL.String() != "https://www.example.com/get" {
		t.Errorf("want %q, but %q", "https://www.example.com/get", req.URL.String())
	}

	_, req, err = curl2http.NewRequestFromArgs([]string{"curl", "-u", "user:passwd", "https://www.example.com/get"})
	if err != nil {
		t.Fatal(err)
	}
	if req.URL.String() != "https://www.example.com/get" {
		t.Errorf("want %q, but %q", "https://www.example.com/get", req.URL.String())
	}
	if req.Header.Get("Authorization") != "Basic dXNlcjpwYXNzd2Q=" {
		t.Errorf("want %q, but %q", "Basic dXNlcjpwYXNzd2Q=", req.Header.Get("Authorization"))
	}

	_, req, err = curl2http.NewRequestFromArgs([]string{"curl", "-I", "-A", "myua", "-b", "session=10", "https://www.example.com/get"})
	if err != nil {
		t.Fatal(err)
	}
	if req.URL.String() != "https://www.example.com/get" {
		t.Errorf("want %q, but %q", "https://www.example.com/get", req.URL.String())
	}
	if req.Method != "HEAD" {
		t.Errorf("want %q, but %q", "HEAD", req.Method)
	}
	if req.Header.Get("User-Agent") != "myua" {
		t.Errorf("want %q, but %q", "myua", req.Header.Get("User-Agent"))
	}
	if req.Header.Get("Cookie") != "session=10" {
		t.Errorf("want %q, but %q", "session=10", req.Header.Get("Cookie"))
	}

	_, req, err = curl2http.NewRequestFromArgs([]string{"curl", "-X", "POST", "-d", "a=b", "-d", "c=d\ne\n", "https://www.example.com/post"})
	if err != nil {
		t.Fatal(err)
	}
	if req.URL.String() != "https://www.example.com/post" {
		t.Errorf("want %q, but %q", "https://www.example.com/post", req.URL.String())
	}
	if req.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
		t.Errorf("want %q, but %q", "application/x-www-form-urlencoded", req.Header.Get("Content-Type"))
	}
	body, err := ioutil.ReadAll(req.Body)
	if string(body) != "a=b&c=de" {
		t.Errorf("want %q, but %q", "a=b&c=de", string(body))
	}

	_, req, err = curl2http.NewRequestFromArgs([]string{"curl", "-X", "POST", "-d", "@testdata/data.txt", "-d", "@testdata/data2.txt", "https://www.example.com/post"})
	if err != nil {
		t.Fatal(err)
	}
	if req.URL.String() != "https://www.example.com/post" {
		t.Errorf("want %q, but %q", "https://www.example.com/post", req.URL.String())
	}
	if req.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
		t.Errorf("want %q, but %q", "application/x-www-form-urlencoded", req.Header.Get("Content-Type"))
	}
	body, err = ioutil.ReadAll(req.Body)
	if string(body) != "a=b&c=d&e=f&g=h" {
		t.Errorf("want %q, but %q", "a=b&c=d&e=f&g=h", string(body))
	}

	json := "{\n\t\"FirstName\": \"hiro\", \n\t\"Age\": 4444\n}"
	_, req, err = curl2http.NewRequestFromArgs([]string{"curl", "-X", "POST", "-H", "Content-Type: application/json", "--data-binary", json, "https://www.example.com/post"})
	if err != nil {
		t.Fatal(err)
	}
	if req.URL.String() != "https://www.example.com/post" {
		t.Errorf("want %q, but %q", "https://www.example.com/post", req.URL.String())
	}
	body, err = ioutil.ReadAll(req.Body)
	if string(body) != json {
		t.Errorf("want %q, but %q", json, string(body))
	}

	_, req, err = curl2http.NewRequestFromArgs([]string{"curl", "-X", "POST", "--data-urlencode", "=<bbb>", "--data-urlencode", "aa=<fa>", "--data-urlencode", "@testdata/data.txt", "--data-urlencode", "eg@testdata/data2.txt", "https://www.example.com/post"})
	if err != nil {
		t.Fatal(err)
	}
	if req.URL.String() != "https://www.example.com/post" {
		t.Errorf("want %q, but %q", "https://www.example.com/post", req.URL.String())
	}
	if req.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
		t.Errorf("want %q, but %q", "application/x-www-form-urlencoded", req.Header.Get("Content-Type"))
	}
	body, err = ioutil.ReadAll(req.Body)
	if string(body) != "%3Cbbb%3E&aa=%3Cfa%3E&a%3Db%26c%3Dd%0A&eg=e%3Df%0A%26%0Ag%3Dh%0A" {
		t.Errorf("want %q, but %q", "%3Cbbb%3E&aa=%3Cfa%3E&a%3Db%26c%3Dd%0A&eg=e%3Df%0A%26%0Ag%3Dh%0A", string(body))
	}

	_, req, err = curl2http.NewRequestFromArgs([]string{"curl", "-X", "POST", "-F", "aa=bbb", "-F", "cc=dddd;type=text/plain", "-F", "kk=@testdata/data.txt", "https://www.example.com/post"})
	if err != nil {
		t.Fatal(err)
	}

	if req.URL.String() != "https://www.example.com/post" {
		t.Errorf("want %q, but %q", "https://www.example.com/post", req.URL.String())
	}
	if !strings.HasPrefix(req.Header.Get("Content-Type"), "multipart/form-data") {
		t.Errorf("want %q, but %q", "multipart/form-data", req.Header.Get("Content-Type"))
	}
	var boundary string
	for _, field := range strings.Split(req.Header.Get("Content-Type"), ";")[1:] {
		kv := strings.SplitN(field, "=", 2)
		if len(kv) == 2 {
			key := strings.TrimSpace(kv[0])
			val := strings.TrimSpace(kv[1])
			if key == "boundary" {
				boundary = val
			}
		}
	}
	mr := multipart.NewReader(req.Body, boundary)
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			return
		}
		if err != nil {
			t.Fatal(err)
		}
		bs, err := ioutil.ReadAll(p)
		if err != nil {
			t.Fatal(err)
		}
		switch p.FormName() {
		case "aa":
			if string(bs) != "bbb" {
				t.Errorf("want %q, but %q", "bbb", string(bs))
			}
		case "cc":
			if string(bs) != "dddd" {
				t.Errorf("want %q, but %q", "dddd", string(bs))
			}
			if p.Header.Get("Content-Type") != "text/plain" {
				t.Errorf("want %q, but %q", "text/plain", p.Header.Get("Content-Type"))
			}
		case "kk":
			if string(bs) != "a=b&c=d\n" {
				t.Errorf("want %q, but %q", "a=b&c=d\n", string(bs))
			}
			if p.FileName() != "data.txt" {
				t.Errorf("want %q, but %q", "data.txt", p.FileName())
			}
		default:
			t.Fatal("unknown part")
		}
	}
}
