curl2http
====

Description
-----------

This library parses `curl` arguments and able to generate corresponding `http.Client` and `http.Request`.

Why?
-----------

Modern web browsers have `copy as cURL` feature in development console.

To take advantage of that, CLI tools have to support cURL compatible options.

This library aims to cut the boilerplate.

Example
-----------

### simple usage

```go
package main

import (
	"io"
	"os"

	"github.com/hirochachacha/curl2http"
)

func main() {
	client, req, err := curl2http.NewRequestFromArgs(os.Args[1:])
	if err != nil {
		panic(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	err = io.Copy(os.Stdout, resp.Body)
	if err != nil {
		panic(err)
	}
}
```

```
go run main.go curl -X POST www.example.com/post -d a=b -d c=d
```

Supported flags
-----------

`curl2http` currently supports following flags:

```
 -b, --cookie <data|filename> Send cookies from string/file
 -d, --data <data>   HTTP POST data
     --data-ascii <data> HTTP POST ASCII data
     --data-binary <data> HTTP POST binary data
     --data-raw <data> HTTP POST data, '@' allowed
     --data-urlencode <data> HTTP POST data url encoded
 -F, --form <name=content> Specify multipart MIME data
 -I, --head          Show document info only
 -H, --header <header/@file> Pass custom header(s) to server
 -k, --insecure      Allow insecure server connections when using SSL
 -L, --location      Follow redirects
 -X, --request <command> Specify request command to use
     --url <url>     URL to work with
 -u, --user <user:password> Server user and password
 -A, --user-agent <name> Send User-Agent <name> to server
```

Following output flags will be skipped during the compilation:

```
     --compressed    Request compressed response
 -i, --include       Include protocol response headers in the output
 -S, --show-error    Show error even when -s is used
 -s, --silent        Silent mode
```

Also, initial "curl" command name is optional.

Similar project
---------------

https://mholt.github.io/curl-to-go/ statically generates Go code.
