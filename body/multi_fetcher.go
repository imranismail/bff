package body

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/google/martian/v3"
	"github.com/google/martian/v3/log"
	"github.com/google/martian/v3/parse"
	"github.com/google/martian/v3/verify"
)

func init() {
	parse.Register("body.MultiFetcher", multiFetcherFromJSON)
}

type multiFetcherJSON struct {
	Scope     []parse.ModifierType `json:"scope"`
	Resources []json.RawMessage    `json:"resources"`
}

// ResourceFetcher WIP
type ResourceFetcher interface {
	verify.ResponseVerifier
	FetchResource() (martian.ResponseModifier, error)
}

// MultiFetcher let you change the name of the fields of the generated responses
type MultiFetcher struct {
	fetchers []ResourceFetcher
}

// NewMultiFetcher constructs and returns a body.JSONDataSourceModifier.
func NewMultiFetcher(fetchers []ResourceFetcher) *MultiFetcher {
	return &MultiFetcher{fetchers: fetchers}
}

// ModifyResponse patches the response body.
func (m *MultiFetcher) ModifyResponse(res *http.Response) error {
	log.Debugf("body.MultiFetcher.ModifyResponse: request: %s", res.Request.URL)

	resources := make(map[int]martian.ResponseModifier)
	resmu := sync.Mutex{}
	merrmu := sync.Mutex{}
	merr := martian.NewMultiError()
	wg := sync.WaitGroup{}

	for i, fetcher := range m.fetchers {
		wg.Add(1)

		go func(i int, fetcher ResourceFetcher) {
			defer wg.Done()

			resource, err := fetcher.FetchResource()

			if err != nil {
				merrmu.Lock()
				merr.Add(err)
				merrmu.Unlock()
			}

			resmu.Lock()
			resources[i] = resource
			resmu.Unlock()
		}(i, fetcher)
	}

	wg.Wait()

	merrmu.Lock()
	if !merr.Empty() {
		return merr
	}
	merrmu.Unlock()

	resmu.Lock()
	for i := 0; i < len(resources); i++ {
		err := resources[i].ModifyResponse(res)

		if err != nil {
			return err
		}
	}
	resmu.Unlock()

	return nil
}

// ResetResponseVerifications clears all failed response verifications.
func (m *MultiFetcher) ResetResponseVerifications() {
	log.Debugf("body.MultiFetcher.ResetResponseVerifications")

	for _, resmod := range m.fetchers {
		resmod.ResetResponseVerifications()
	}
}

// VerifyResponses returns a MultiError containing all the
// verification errors returned by response verifiers.
func (m *MultiFetcher) VerifyResponses() error {
	log.Debugf("body.MultiFetcher.VerifyResponse")

	merr := martian.NewMultiError()
	for _, fetcher := range m.fetchers {

		if err := fetcher.VerifyResponses(); err != nil {
			merr.Add(err)
		}
	}

	if merr.Empty() {
		return nil
	}

	return merr
}

func multiFetcherFromJSON(b []byte) (*parse.Result, error) {
	msg := &multiFetcherJSON{}
	fetchers := make([]ResourceFetcher, 0)

	if err := json.Unmarshal(b, msg); err != nil {
		return nil, err
	}

	for _, resource := range msg.Resources {
		r, err := parse.FromJSON(resource)

		if err != nil {
			return nil, err
		}

		resmod := r.ResponseModifier()

		if fetcher, ok := resmod.(ResourceFetcher); ok {
			fetchers = append(fetchers, fetcher)
		}
	}

	mod := NewMultiFetcher(fetchers)

	return parse.NewResult(mod, msg.Scope)
}
