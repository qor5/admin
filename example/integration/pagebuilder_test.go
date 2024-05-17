package integration_test

import (
	"net/http/httptest"
	"testing"

	"github.com/qor5/admin/v3/example/admin"
	"github.com/theplant/testenv"
	"gorm.io/gorm"
)

var TestDB *gorm.DB

func TestMain(m *testing.M) {
	env, err := testenv.New().DBEnable(true).SetUp()
	if err != nil {
		panic(err)
	}
	defer env.TearDown()
	TestDB = env.DB
	m.Run()
}

func TestPageBuilder(t *testing.T) {
	h := admin.TestHandler(TestDB)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/pages", nil))
	t.Log(w.Body.String())
}
