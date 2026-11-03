package tripadvisor

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// fakeStep is the outcome of one round trip: a transport error, or a response with the given status and body.
type fakeStep struct {
	err    error
	status int
	body   string
}

// fakeTransport replays steps in order (repeating the last one) and records the request bodies it receives.
type fakeTransport struct {
	steps  []fakeStep
	bodies []string
}

func (f *fakeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	body, _ := io.ReadAll(req.Body)
	f.bodies = append(f.bodies, string(body))
	step := f.steps[min(len(f.bodies)-1, len(f.steps)-1)]
	if step.err != nil {
		return nil, step.err
	}
	return &http.Response{
		StatusCode: step.status,
		Body:       io.NopCloser(strings.NewReader(step.body)),
		Header:     make(http.Header),
		Request:    req,
	}, nil
}

const okBody = `[{"data":{"ReviewsProxy_getPublishedReviewsForLocation":[{"error":null,"result":{"totalCount":650,"reviews":[]}}]}}]`

func TestMakeRequestRetry(t *testing.T) {
	previous := retryBackoff
	retryBackoff = 0
	t.Cleanup(func() { retryBackoff = previous })

	eof := fakeStep{err: io.ErrUnexpectedEOF}
	ok := fakeStep{status: http.StatusOK, body: okBody}

	tests := []struct {
		name      string
		steps     []fakeStep
		wantCalls int
		wantErr   string
	}{
		{name: "dropped connections are retried until success", steps: []fakeStep{eof, eof, ok}, wantCalls: 3},
		{name: "server errors are retried", steps: []fakeStep{{status: http.StatusServiceUnavailable}, ok}, wantCalls: 2},
		{name: "gives up after max attempts", steps: []fakeStep{eof}, wantCalls: MaxRequestAttempts, wantErr: "unexpected EOF"},
		{name: "rate limit is not retried", steps: []fakeStep{{status: http.StatusTooManyRequests}, ok}, wantCalls: 1, wantErr: "rate Limit Detected"},
		{name: "graphql errors are not retried", steps: []fakeStep{{status: http.StatusOK, body: `[{"errors":[{"message":"PersistedQueryNotFound"}]}]`}, ok}, wantCalls: 1, wantErr: "PersistedQueryNotFound"},
		{name: "other client errors are not retried", steps: []fakeStep{{status: http.StatusForbidden}, ok}, wantCalls: 1, wantErr: "403"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := &fakeTransport{steps: tt.steps}
			client := &http.Client{Transport: transport}

			resp, err := MakeRequest(client, HotelQueryID, "HOTEL", "en", 17452639, 60, ReviewLimit)

			assert.Len(t, transport.bodies, tt.wantCalls)
			// Every attempt must send the full request body, not an already drained one.
			for _, body := range transport.bodies {
				assert.Equal(t, transport.bodies[0], body)
				assert.Contains(t, body, `"offset":60`)
			}
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				assert.Nil(t, resp)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, 650, ExtractTotalCount(resp))
		})
	}
}
