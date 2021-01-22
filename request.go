package curl2http

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"path"
	"strings"
)

var SupportedRequestFlags = map[string]bool{
	"-I":               true,
	"--head":           true,
	"-X":               true,
	"--request":        true,
	"-H":               true,
	"--header":         true,
	"-b":               true,
	"--cookie":         true,
	"-A":               true,
	"--user-agent":     true,
	"-u":               true,
	"--user":           true,
	"-k":               true,
	"--insecure":       true,
	"-L":               true,
	"--location":       true,
	"-d":               true,
	"--data":           true,
	"--data-ascii":     true,
	"--data-raw":       true,
	"--data-urlencode": true,
	"--data-binary":    true,
	"-G":               true,
	"--get":            true,
	"-F":               true,
	"--form":           true,
	"--url":            true,

	// ignore output options
	"--compressed": true,
	"-i":           true,
	"--include":    true,
	"-S":           true,
	"--show-error": true,
	"-s":           true,
	"--silent":     true,
}

func NewRequestFromFlagSet(fs *FlagSet) (client *http.Client, req *http.Request, err error) {
	client = new(http.Client)

	var (
		method      string
		_url        string
		rawQuery    string
		body        io.Reader
		headers     [][2]string
		host        string
		contentType string
	)

	if arg := fs.Arg(-1); arg != "curl" {
		_url = arg
	}

	if flg := fs.LongFlag("head"); flg.IsSet {
		method = "HEAD"
	}
	if flg := fs.LongFlag("request"); flg.IsSet {
		method = flg.Value(-1)
	}
	if flg := fs.LongFlag("header"); flg.IsSet {
		for _, val := range flg.Values {
			kv := strings.SplitN(val, ":", 2)
			if len(kv) != 2 {
				return nil, nil, fmt.Errorf("header value must contain ':'")
			}
			key := strings.Title(strings.TrimSpace(kv[0]))
			val := strings.TrimSpace(kv[1])

			switch key {
			case "Host":
				host = val
			case "Content-Type":
				contentType = val
			default:
				headers = append(headers, [2]string{key, val})
			}
		}
	}
	if flg := fs.LongFlag("cookie"); flg.IsSet {
		val := flg.Value(-1)
		if strings.Contains(val, "=") {
			headers = append(headers, [2]string{"Cookie", val})
		} else {
			bs, err := ioutil.ReadFile(val)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to read cookie file: %w", err)
			}
			headers = append(headers, [2]string{"Cookie", string(bs)})
		}
	}
	if flg := fs.LongFlag("user-agent"); flg.IsSet {
		headers = append(headers, [2]string{"User-Agent", flg.Value(-1)})
	}
	if flg := fs.LongFlag("user"); flg.IsSet {
		headers = append(headers, [2]string{
			"Authorization",
			"Basic " + base64.StdEncoding.EncodeToString([]byte(flg.Value(-1))),
		})
	}
	if flg := fs.LongFlag("insecure"); flg.IsSet {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}
	if flg := fs.LongFlag("location"); !flg.IsSet {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}
	for _, name := range []string{"data", "data-ascii", "data-raw", "data-urlencode", "data-binary"} {
		if flg := fs.LongFlag(name); flg.IsSet {
			if body != nil {
				return nil, nil, fmt.Errorf("request body is already filled")
			}

			if method == "" {
				method = "POST"
			}

			if name != "data-binary" {
				if contentType == "" {
					contentType = "application/x-www-form-urlencoded"
				}
			}

			for _, val := range flg.Values {
				var r io.Reader
				if name == "data-urlencode" {
					switch {
					case strings.HasPrefix(val, "@"):
						bs, err := ioutil.ReadFile(val[1:])
						if err != nil {
							return nil, nil, fmt.Errorf("failed to read %s file: %w", name, err)
						}
						r = strings.NewReader(url.QueryEscape(string(bs)))
					case strings.Contains(val, "@"):
						kv := strings.SplitN(val, "@", 2)
						bs, err := ioutil.ReadFile(kv[1])
						if err != nil {
							return nil, nil, fmt.Errorf("failed to read %s file: %w", name, err)
						}
						r = strings.NewReader(kv[0] + "=" + url.QueryEscape(string(bs)))
					case strings.HasPrefix(val, "="):
						r = strings.NewReader(url.QueryEscape(val[1:]))
					case strings.Contains(val, "="):
						kv := strings.SplitN(val, "=", 2)
						r = strings.NewReader(kv[0] + "=" + url.QueryEscape(kv[1]))
					default:
						r = strings.NewReader(url.QueryEscape(val))
					}
				} else {
					if name != "data-raw" && strings.HasPrefix(val, "@") {
						bs, err := ioutil.ReadFile(val[1:])
						if err != nil {
							return nil, nil, fmt.Errorf("failed to read %s file: %w", name, err)
						}
						r = bytes.NewReader(bs)
					} else {
						r = strings.NewReader(val)
					}
					if name != "data-binary" {
						r = newAsciiReader(r)
					}
				}
				if body == nil {
					body = r
				} else {
					if name == "data-binary" {
						body = io.MultiReader(body, r)
					} else {
						body = io.MultiReader(body, strings.NewReader("&"), r)
					}
				}
			}
		}
	}

	if flg := fs.LongFlag("get"); flg.IsSet {
		if flg = fs.LongFlag("data-binary"); flg.IsSet {
			return nil, nil, fmt.Errorf("cannot use --get with --data-binary")
		}

		method = "GET"
		contentType = ""

		if body != nil {
			bs, err := ioutil.ReadAll(body)
			if err != nil {
				return nil, nil, err
			}
			rawQuery = string(bs)
			body = nil
		}
	}

	if flg := fs.LongFlag("form"); flg.IsSet {
		if body != nil {
			return nil, nil, fmt.Errorf("request body is already filled")
		}

		if method == "" {
			method = "POST"
		}

		buf := new(bytes.Buffer)

		w := multipart.NewWriter(buf)

		if contentType == "" {
			contentType = w.FormDataContentType()
		}

		for _, val := range flg.Values {
			fields := strings.Split(val, ";")
			kv := strings.SplitN(fields[0], "=", 2)
			if len(kv) != 2 {
				return nil, nil, fmt.Errorf("form value must contain '='")
			}
			fieldName := strings.TrimSpace(kv[0])
			fieldContent := strings.TrimSpace(kv[1])

			var fieldType string
			if len(fields) > 1 {
				for _, field := range fields[1:] {
					kv := strings.SplitN(field, "=", 2)
					if len(kv) == 2 {
						key := strings.TrimSpace(kv[0])
						if key == "type" {
							fieldType = strings.TrimSpace(kv[1])
						}
					}
				}
			}

			h := make(textproto.MIMEHeader)

			if strings.HasPrefix(fieldContent, "@") {
				fileName := path.Base(fieldContent)
				bs, err := ioutil.ReadFile(fieldContent[1:])
				if err != nil {
					return nil, nil, fmt.Errorf("failed to read form-data file: %w", err)
				}
				fieldContent = string(bs)
				h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, escapeQuotes(fieldName), escapeQuotes(fileName)))
			} else {
				h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"`, escapeQuotes(fieldName)))
			}

			if fieldType != "" {
				h.Set("Content-Type", escapeQuotes(fieldType))
			}

			part, err := w.CreatePart(h)
			if err != nil {
				return nil, nil, err
			}

			_, err = part.Write([]byte(fieldContent))
			if err != nil {
				return nil, nil, err
			}
		}

		err = w.Close()
		if err != nil {
			return nil, nil, err
		}

		body = buf
	}

	if flg := fs.LongFlag("url"); flg.IsSet {
		_url = flg.Value(-1)
	}

	fs.Visit(func(flg *Flag) {
		if flg.ShortName != "" && !SupportedRequestFlags["-"+flg.ShortName] {
			if err == nil {
				err = fmt.Errorf("unsupported flag: %s", "-"+flg.ShortName)
			}
		}
		if flg.LongName != "" && !SupportedRequestFlags["--"+flg.LongName] {
			if err == nil {
				err = fmt.Errorf("unsupported flag: %s", "--"+flg.LongName)
			}
		}
	})
	if err != nil {
		return nil, nil, err
	}

	if method == "" {
		method = "GET"
	}

	if _url == "" {
		return nil, nil, fmt.Errorf("url is missing")
	}

	u, err := url.Parse(_url)
	if err != nil {
		return nil, nil, err
	}

	if rawQuery != "" {
		if u.RawQuery != "" {
			u.RawQuery += "&" + rawQuery
		} else {
			u.RawQuery = rawQuery
		}
	}

	if u.Scheme == "" {
		u.Scheme = "http"
	}

	req, err = http.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, nil, err
	}

	if host != "" {
		req.Host = host
	}

	if contentType != "" {
		req.Header.Add("Content-Type", contentType)
	}

	for _, header := range headers {
		req.Header.Add(header[0], header[1])
	}

	return client, req, nil
}
