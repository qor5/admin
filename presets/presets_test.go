package presets

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/hook"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsMenuItemActive(t *testing.T) {
	cases := []struct {
		// path means current url path
		path string
		// link means menu item link
		link string

		excepted bool
	}{
		{"", "/", true},
		{"/", "/", true},
		{"/", "/order", false},
		{"/order", "/order", true},
		{"/order/1", "/order", true},
		{"/order#", "/order", true},
		{"/product", "/order", false},
		{"/product", "/", false},
	}

	type io struct {
		ctx      *web.EventContext
		m        *ModelBuilder
		excepted bool
	}

	var toIO []io
	b := New()
	for _, c := range cases {
		toIO = append(toIO, io{
			ctx: &web.EventContext{
				R: &http.Request{
					URL: &url.URL{
						Path: c.path,
					},
				},
			},
			m: &ModelBuilder{
				link:      c.link,
				modelInfo: &ModelInfo{mb: NewModelBuilder(b, &struct{}{})},
			},
			excepted: c.excepted,
		})
	}

	for i, io := range toIO {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if b.menuOrder.isMenuItemActive(io.m, io.ctx) != io.excepted {
				t.Errorf("isMenuItemActive() = %v, excepted %v", b.menuOrder.isMenuItemActive(io.m, io.ctx), io.excepted)
			}
		})
	}
}

func TestLookUpModelBuilder(t *testing.T) {
	type Order struct {
		ID      uint
		Product string
	}
	type Customer struct {
		ID   uint
		Name string
	}

	pb := New()
	mb0 := pb.Model(&Order{})
	mb1 := pb.Model(&Customer{})
	assert.Equal(t, mb0, pb.LookUpModelBuilder(mb0.Info().URIName()))
	assert.Equal(t, mb1, pb.LookUpModelBuilder(mb1.Info().URIName()))

	mb3 := pb.Model(&Customer{})
	assert.PanicsWithValue(t, `Duplicated model names registered "customers"`, func() {
		pb.LookUpModelBuilder(mb3.Info().URIName())
	})
	mb3.URIName(mb3.Info().URIName() + "-version-list-dialog")
	assert.Equal(t, mb3, pb.LookUpModelBuilder(mb3.Info().URIName()))
}

func TestCloneFieldsLayout(t *testing.T) {
	src := []any{
		"foo",
		[]string{"bar"},
		&FieldsSection{
			Title: "title",
			Rows: [][]string{
				{"a", "b"},
				{"c", "d"},
			},
		},
	}
	require.Equal(t, src, CloneFieldsLayout(src))
}

func TestHumanizeString(t *testing.T) {
	assert.Equal(t, "Hello World", humanizeString("HelloWorld"))
	assert.Equal(t, "Hello World", humanizeString("helloWorld"))
	assert.Equal(t, "Order Item", humanizeString("OrderItem"))
	assert.Equal(t, "CNN Name", humanizeString("CNNName"))
}

func TestWithHandlerHook(t *testing.T) {
	b := New()

	// Create hooks
	hook1 := hook.Hook[http.Handler](func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Hook-1", "first")
			next.ServeHTTP(w, r)
		})
	})

	hook2 := hook.Hook[http.Handler](func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Hook-2", "second")
			next.ServeHTTP(w, r)
		})
	})

	// Test chaining
	result := b.WithHandlerHook(hook1, hook2)
	assert.Equal(t, b, result)

	// Test hooks are applied
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	finalHandler := b.handlerHook(testHandler)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", http.NoBody)
	finalHandler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "test response", w.Body.String())
	assert.Equal(t, "first", w.Header().Get("X-Hook-1"))
	assert.Equal(t, "second", w.Header().Get("X-Hook-2"))
}

func TestNewMuxHook(t *testing.T) {
	b := New()

	// Create a mux with a route
	mux := http.NewServeMux()
	mux.HandleFunc("/api/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("mux response"))
	})

	// Create the hook
	muxHook := b.NewMuxHook(mux)
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("next handler"))
	})
	wrappedHandler := muxHook(nextHandler)

	// Test mux route
	w1 := httptest.NewRecorder()
	r1 := httptest.NewRequest("GET", "/api/test", http.NoBody)
	wrappedHandler.ServeHTTP(w1, r1)
	assert.Equal(t, http.StatusOK, w1.Code)
	assert.Equal(t, "mux response", w1.Body.String())

	// Test fallback to next handler
	w2 := httptest.NewRecorder()
	r2 := httptest.NewRequest("GET", "/other", http.NoBody)
	wrappedHandler.ServeHTTP(w2, r2)
	assert.Equal(t, http.StatusOK, w2.Code)
	assert.Equal(t, "next handler", w2.Body.String())
}

func TestServeHTTP_HandlerNilAfterOnce_Returns500(t *testing.T) {
	b := New()

	// Mark warmupOnce as done to skip Build() so that handler stays nil.
	b.warmupOnce.Do(func() {})

	// Capture log output.
	var buf bytes.Buffer
	orig := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(orig)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/foo", http.NoBody)

	b.ServeHTTP(w, r)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Internal Server Error")
	assert.Contains(t, buf.String(), "Builder.handler is nil after Build")
	assert.Contains(t, buf.String(), "/foo")
}
