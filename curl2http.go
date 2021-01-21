package curl2http

//go:generate sh -c "curl -h > assets/help.txt"
//go:generate pkger

import (
	"net/http"
)

func NewRequestFromArgs(args []string) (client *http.Client, req *http.Request, err error) {
	fs := NewFlagSet()

	err = fs.Parse(args)
	if err != nil {
		return nil, nil, err
	}

	return NewRequestFromFlagSet(fs)
}
