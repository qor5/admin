package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"testing"

	"github.com/qor5/admin/v3/autocomplete"

	"github.com/theplant/gofixtures"
	"github.com/theplant/testenv"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var TestDB *gorm.DB

func TestMain(m *testing.M) {
	env, err := testenv.New().DBEnable(true).SetUp()
	if err != nil {
		panic(err)
	}
	defer env.TearDown()
	TestDB = env.DB
	TestDB.Logger = TestDB.Logger.LogMode(logger.Info)
	m.Run()
}

var autocompleteData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.categories (id, created_at, updated_at, deleted_at, name, path) VALUES 
   (1, '2024-05-17 15:25:31.134801 +00:00', '2024-05-17 15:25:31.134801 +00:00', null, '1', '/1'),
   (2, '2024-05-17 15:25:31.134801 +00:00', '2024-05-17 15:25:31.134801 +00:00', null, '2', '/2'),
   (3, '2024-05-17 15:25:31.134801 +00:00', '2024-05-17 15:25:31.134801 +00:00', null, '3', '/3'),
   (4, '2024-05-17 15:25:31.134801 +00:00', '2024-05-17 15:25:31.134801 +00:00', null, '4', '/4'),
   (5, '2024-05-17 15:25:31.134801 +00:00', '2024-05-17 15:25:31.134801 +00:00', null, '5', '/5');
INSERT INTO public.users (id, created_at, updated_at, deleted_at, name, age) VALUES 
   (1, '2024-05-17 15:25:31.134801 +00:00', '2024-05-17 15:25:31.134801 +00:00', null, 'k1', 20),
   (2, '2024-05-17 15:25:31.134801 +00:00', '2024-05-17 15:25:31.134801 +00:00', null, 'k2', 21),
   (3, '2024-05-17 15:25:31.134801 +00:00', '2024-05-17 15:25:31.134801 +00:00', null, 'b3', 22),
   (4, '2024-05-17 15:25:31.134801 +00:00', '2024-05-17 15:25:31.134801 +00:00', null, 'b4', 23),
   (5, '2024-05-17 15:25:31.134801 +00:00', '2024-05-17 15:25:31.134801 +00:00', null, 'b5', 24);
`, []string{"categories", "users"}))

type (
	Category struct {
		gorm.Model
		Name string `json:"name"`
		Path string `json:"path"`
	}
	User struct {
		gorm.Model
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
)

func Handler(db *gorm.DB) http.Handler {
	_ = db.AutoMigrate(Category{}, User{})
	mux := http.NewServeMux()
	b := autocomplete.New().DB(db).Prefix("/complete").AllowCrossOrigin(true)
	b.Model(&Category{}).Columns("id", "name", "path").OrderBy("id desc")
	b.Model(&User{}).Columns("id", "name", "age").SQLCondition("name ilike ?")
	mux.Handle("/complete/", b)
	return b
}

func runTest(t *testing.T, r *http.Request, handler http.Handler) *bytes.Buffer {
	w := httptest.NewRecorder()
	bs, _ := httputil.DumpRequest(r, true)
	t.Log("======== Request ========")
	t.Log(string(bs))
	handler.ServeHTTP(w, r)
	t.Log("======== Response ========")
	t.Log(w.Header())
	t.Log(w.Body.String())
	return w.Body
}

func TestCategory(t *testing.T) {
	handler := Handler(TestDB)
	dbr, _ := TestDB.DB()
	t.Log("Test Categories")
	autocompleteData.TruncatePut(dbr)
	var (
		err      error
		response autocomplete.Response
	)
	bytes := runTest(t, httptest.NewRequest("GET", "/complete/categories", http.NoBody), handler)
	if err = json.Unmarshal(bytes.Bytes(), &response); err != nil {
		t.Fatalf("json unmarshal faield :%v", err)
		return
	}
	count := 5
	if response.Total != int64(count) {
		t.Fatalf("except get %v but get %v", count, response.Total)
		return
	}
	if response.Data[0]["path"] == nil {
		t.Fatalf("unexcpet value")
		return
	}
}

func TestUser(t *testing.T) {
	handler := Handler(TestDB)
	dbr, _ := TestDB.DB()
	t.Log("Test Users")
	autocompleteData.TruncatePut(dbr)
	var (
		err      error
		response autocomplete.Response
	)

	bytes := runTest(t, httptest.NewRequest("GET", "/complete/users", http.NoBody), handler)
	if err = json.Unmarshal(bytes.Bytes(), &response); err != nil {
		t.Fatalf("json unmarshal faield :%v", err)
		return
	}
	count := 5
	if response.Total != int64(count) {
		t.Fatalf("except get %v but get %v", count, response.Total)
		return
	}
	if response.Data[0]["age"] == nil {
		t.Fatalf("unexcpet value")
		return
	}
}

func TestUserSearch(t *testing.T) {
	handler := Handler(TestDB)
	dbr, _ := TestDB.DB()
	t.Log("Test Users Search")
	autocompleteData.TruncatePut(dbr)
	var (
		err      error
		response autocomplete.Response
	)

	bytes := runTest(t, httptest.NewRequest("GET", "/complete/users?search=k", http.NoBody), handler)
	if err = json.Unmarshal(bytes.Bytes(), &response); err != nil {
		t.Fatalf("json unmarshal faield :%v", err)
		return
	}
	count := 2
	if response.Total != int64(count) {
		t.Fatalf("except get %v but get %v", count, response.Total)
		return
	}
	if response.Data[0]["age"] == nil {
		t.Fatalf("unexcpet value")
		return
	}
}

func TestCategoryPage(t *testing.T) {
	handler := Handler(TestDB)
	dbr, _ := TestDB.DB()
	t.Log("Test Category Page")
	autocompleteData.TruncatePut(dbr)
	var (
		err      error
		response autocomplete.Response
	)

	bytes := runTest(t, httptest.NewRequest("GET", "/complete/categories?page=2&pageSize=3", http.NoBody), handler)
	if err = json.Unmarshal(bytes.Bytes(), &response); err != nil {
		t.Fatalf("json unmarshal faield :%v", err)
		return
	}
	count := 2
	if len(response.Data) != count {
		t.Fatalf("except get %v but get %v", count, len(response.Data))
		return
	}
	if response.Data[0]["path"] == nil {
		t.Fatalf("unexcpet value")
		return
	}
}
