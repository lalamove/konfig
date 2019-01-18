package klhttp

import (
	"fmt"
	"io"
	"net/http"
)

// Do makes an http request and sends the body to the parser
func (s Source) Do(c Client) (io.Reader, error) {
	var req, err = http.NewRequest(
		s.Method,
		s.URL,
		s.Body,
	)
	if err != nil {
		return nil, err
	}

	// call the prepare method if there is one
	if s.Prepare != nil {
		s.Prepare(req)
	}

	// make the request
	var res *http.Response
	res, err = c.Do(req)
	if err != nil {
		return nil, err
	}

	// check status code
	if (s.StatusCode != 0 && res.StatusCode != s.StatusCode) ||
		(res.StatusCode != http.StatusOK) {

		return nil, fmt.Errorf(
			"Error while fetching config at %s, status code: %d",
			s.URL,
			res.StatusCode,
		)
	}

	return res.Body, nil
}
