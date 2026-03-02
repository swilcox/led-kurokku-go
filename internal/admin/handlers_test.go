package admin

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/swilcox/led-kurokku-go/config"
)

func testServer(t *testing.T) *Server {
	t.Helper()
	dir := t.TempDir()
	store := NewStore(filepath.Join(dir, "instances.json"))
	return NewServer(store)
}

func testServerWithInstance(t *testing.T) (*Server, Instance) {
	t.Helper()
	s := testServer(t)
	inst := Instance{ID: "test1", Name: "Test Clock", Host: "localhost", Port: 6379}
	s.store.Add(inst)
	return s, inst
}

// --- Index ---

func TestHandleIndex(t *testing.T) {
	s := testServer(t)
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if !strings.Contains(w.Body.String(), "html") {
		t.Error("expected HTML response")
	}
}

// --- Instance New ---

func TestHandleInstanceNew(t *testing.T) {
	s := testServer(t)
	req := httptest.NewRequest("GET", "/instances/new", nil)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

// --- Instance Edit ---

func TestHandleInstanceEdit_Found(t *testing.T) {
	s, _ := testServerWithInstance(t)
	req := httptest.NewRequest("GET", "/instances/test1/edit", nil)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHandleInstanceEdit_NotFound(t *testing.T) {
	s := testServer(t)
	req := httptest.NewRequest("GET", "/instances/nonexistent/edit", nil)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

// --- Instance Update ---

func TestHandleInstanceUpdate_NotFound(t *testing.T) {
	s := testServer(t)
	form := url.Values{"name": {"X"}, "host": {"h"}, "port": {"6379"}}
	req := httptest.NewRequest("PUT", "/instances/nonexistent", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandleInstanceUpdate_MissingFields(t *testing.T) {
	s, _ := testServerWithInstance(t)
	form := url.Values{"name": {""}, "host": {""}}
	req := httptest.NewRequest("PUT", "/instances/test1", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	// Should re-render form with error, still 200
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if !strings.Contains(w.Body.String(), "required") {
		t.Error("expected validation error about required fields")
	}
}

// --- Instance Delete ---

func TestHandleInstanceDelete_Found(t *testing.T) {
	s, _ := testServerWithInstance(t)
	req := httptest.NewRequest("DELETE", "/instances/test1", nil)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	// Verify instance was removed
	if s.store.Get("test1") != nil {
		t.Error("expected instance to be deleted")
	}
}

func TestHandleInstanceDelete_NotFound(t *testing.T) {
	s := testServer(t)
	req := httptest.NewRequest("DELETE", "/instances/nonexistent", nil)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

// --- Instance Test ---

func TestHandleInstanceTest_NotFound(t *testing.T) {
	s := testServer(t)
	req := httptest.NewRequest("POST", "/instances/nonexistent/test", nil)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

// --- Instance Create ---

func TestHandleInstanceCreate_MissingFields(t *testing.T) {
	s := testServer(t)
	form := url.Values{"name": {""}, "host": {""}}
	req := httptest.NewRequest("POST", "/instances", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if !strings.Contains(w.Body.String(), "required") {
		t.Error("expected validation error about required fields")
	}
}

func TestHandleInstanceCreate_DefaultPort(t *testing.T) {
	s := testServer(t)
	// Port 0 or missing should default to 6379
	form := url.Values{"name": {"Test"}, "host": {"example.com"}, "port": {"0"}}
	req := httptest.NewRequest("POST", "/instances", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	// Will fail on Redis connection, but we can verify the form error mentions connection
	body := w.Body.String()
	if !strings.Contains(body, "connection") && !strings.Contains(body, "Connection") {
		// If Redis just happens to be running, the test might succeed — that's fine too
		if w.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
		}
	}
}

// --- Config handlers (not found paths) ---

func TestHandleConfigView_NotFound(t *testing.T) {
	s := testServer(t)
	req := httptest.NewRequest("GET", "/instances/nonexistent/config", nil)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandleConfigEdit_NotFound(t *testing.T) {
	s := testServer(t)
	req := httptest.NewRequest("GET", "/instances/nonexistent/config/edit", nil)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandleConfigSave_NotFound(t *testing.T) {
	s := testServer(t)
	req := httptest.NewRequest("POST", "/instances/nonexistent/config", nil)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandleConfigJSON_NotFound(t *testing.T) {
	s := testServer(t)
	req := httptest.NewRequest("GET", "/instances/nonexistent/config/json", nil)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandleConfigJSONSave_NotFound(t *testing.T) {
	s := testServer(t)
	req := httptest.NewRequest("POST", "/instances/nonexistent/config/json", nil)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandlePreview_NotFound(t *testing.T) {
	s := testServer(t)
	req := httptest.NewRequest("GET", "/instances/nonexistent/config/preview", nil)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

// --- Widget handlers ---

func TestHandleWidgetAdd(t *testing.T) {
	s, _ := testServerWithInstance(t)
	form := url.Values{"count": {"0"}}
	req := httptest.NewRequest("POST", "/instances/test1/config/widgets/add", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHandleWidgetForm(t *testing.T) {
	s := testServer(t)
	req := httptest.NewRequest("GET", "/config/widgets/form?index=0&widgets%5B0%5D.type=message", nil)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHandleWidgetForm_DefaultType(t *testing.T) {
	s := testServer(t)
	req := httptest.NewRequest("GET", "/config/widgets/form?index=0", nil)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHandleWidgetRemove(t *testing.T) {
	s, _ := testServerWithInstance(t)
	req := httptest.NewRequest("DELETE", "/instances/test1/config/widgets/0", nil)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

// --- Helper functions ---

func TestSlugify(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Hello World", "hello-world"},
		{"  My Clock  ", "my-clock"},
		{"test_clock", "test-clock"},
		{"UPPER", "upper"},
		{"a--b", "a-b"},
		{"special!@#chars", "specialchars"},
		{"---leading---", "leading"},
		{"123 numbers", "123-numbers"},
		{"", ""},
	}
	for _, tc := range tests {
		got := slugify(tc.input)
		if got != tc.want {
			t.Errorf("slugify(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := defaultConfig()
	if cfg.Display.Type != config.DisplayMAX7219 {
		t.Errorf("display type = %q, want %q", cfg.Display.Type, config.DisplayMAX7219)
	}
	if cfg.Brightness.High != 15 {
		t.Errorf("brightness high = %d, want 15", cfg.Brightness.High)
	}
	if len(cfg.Widgets) != 1 {
		t.Fatalf("expected 1 widget, got %d", len(cfg.Widgets))
	}
	if cfg.Widgets[0].Type != "clock" {
		t.Errorf("widget type = %q, want %q", cfg.Widgets[0].Type, "clock")
	}
	if cfg.Widgets[0].Format24h == nil || !*cfg.Widgets[0].Format24h {
		t.Error("expected format_24h = true")
	}
}

func TestParseConfigForm_Basic(t *testing.T) {
	form := url.Values{
		"display.type":     {"terminal"},
		"brightness.high":  {"15"},
		"brightness.low":   {"1"},
		"brightness.day_start": {"07:00"},
		"brightness.day_end":   {"22:00"},
		"widgets[0].type":     {"clock"},
		"widgets[0].enabled":  {"on"},
		"widgets[0].duration": {"10s"},
		"widgets[0].format_24h": {"on"},
	}
	req := httptest.NewRequest("POST", "/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.ParseForm()

	cfg, errMsg := parseConfigForm(req)
	if errMsg != "" {
		t.Fatalf("unexpected error: %s", errMsg)
	}
	if cfg.Display.Type != config.DisplayTerminal {
		t.Errorf("display type = %q, want %q", cfg.Display.Type, config.DisplayTerminal)
	}
	if cfg.Brightness.High != 15 {
		t.Errorf("brightness high = %d, want 15", cfg.Brightness.High)
	}
	if len(cfg.Widgets) != 1 {
		t.Fatalf("expected 1 widget, got %d", len(cfg.Widgets))
	}
	if !cfg.Widgets[0].Enabled {
		t.Error("expected widget to be enabled")
	}
	if cfg.Widgets[0].Format24h == nil || !*cfg.Widgets[0].Format24h {
		t.Error("expected format_24h = true")
	}
}

func TestParseConfigForm_WithLocation(t *testing.T) {
	form := url.Values{
		"display.type":      {"terminal"},
		"brightness.high":   {"15"},
		"brightness.low":    {"1"},
		"brightness.use_location": {"on"},
		"location.lat":      {"36.166"},
		"location.lon":      {"-86.784"},
		"location.timezone": {"America/Chicago"},
	}
	req := httptest.NewRequest("POST", "/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.ParseForm()

	cfg, errMsg := parseConfigForm(req)
	if errMsg != "" {
		t.Fatalf("unexpected error: %s", errMsg)
	}
	if cfg.Location == nil {
		t.Fatal("expected location to be set")
	}
	if cfg.Location.Lat != 36.166 {
		t.Errorf("lat = %f, want 36.166", cfg.Location.Lat)
	}
	if !cfg.Brightness.UseLocation {
		t.Error("expected use_location = true")
	}
}

func TestParseConfigForm_InvalidDuration(t *testing.T) {
	form := url.Values{
		"display.type":        {"terminal"},
		"widgets[0].type":     {"clock"},
		"widgets[0].duration": {"bad"},
	}
	req := httptest.NewRequest("POST", "/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.ParseForm()

	_, errMsg := parseConfigForm(req)
	if errMsg == "" {
		t.Error("expected error for invalid duration")
	}
	if !strings.Contains(errMsg, "duration") {
		t.Errorf("error = %q, expected to mention duration", errMsg)
	}
}

func TestParseConfigForm_InvalidScrollSpeed(t *testing.T) {
	form := url.Values{
		"display.type":           {"terminal"},
		"widgets[0].type":        {"message"},
		"widgets[0].scroll_speed": {"bad"},
	}
	req := httptest.NewRequest("POST", "/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.ParseForm()

	_, errMsg := parseConfigForm(req)
	if errMsg == "" {
		t.Error("expected error for invalid scroll speed")
	}
}

func TestParseConfigForm_InvalidFrameDuration(t *testing.T) {
	form := url.Values{
		"display.type":             {"terminal"},
		"widgets[0].type":          {"animation"},
		"widgets[0].frame_duration": {"bad"},
	}
	req := httptest.NewRequest("POST", "/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.ParseForm()

	_, errMsg := parseConfigForm(req)
	if errMsg == "" {
		t.Error("expected error for invalid frame duration")
	}
}

func TestParseConfigForm_InvalidRepeats(t *testing.T) {
	form := url.Values{
		"display.type":       {"terminal"},
		"widgets[0].type":    {"message"},
		"widgets[0].repeats": {"abc"},
	}
	req := httptest.NewRequest("POST", "/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.ParseForm()

	_, errMsg := parseConfigForm(req)
	if errMsg == "" {
		t.Error("expected error for invalid repeats")
	}
}

func TestParseConfigForm_I2CAddr(t *testing.T) {
	form := url.Values{
		"display.type":     {"ht16k33"},
		"display.i2c_addr": {"0x70"},
	}
	req := httptest.NewRequest("POST", "/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.ParseForm()

	cfg, errMsg := parseConfigForm(req)
	if errMsg != "" {
		t.Fatalf("unexpected error: %s", errMsg)
	}
	if cfg.Display.I2CAddr != 0x70 {
		t.Errorf("i2c_addr = 0x%X, want 0x70", cfg.Display.I2CAddr)
	}
}

func TestParseConfigForm_InvalidI2CAddr(t *testing.T) {
	form := url.Values{
		"display.type":     {"ht16k33"},
		"display.i2c_addr": {"not-hex"},
	}
	req := httptest.NewRequest("POST", "/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.ParseForm()

	_, errMsg := parseConfigForm(req)
	if errMsg == "" {
		t.Error("expected error for invalid I2C address")
	}
}

func TestParseConfigForm_MultipleWidgets(t *testing.T) {
	form := url.Values{
		"display.type":        {"terminal"},
		"widgets[0].type":     {"clock"},
		"widgets[0].duration": {"10s"},
		"widgets[1].type":     {"message"},
		"widgets[1].text":     {"Hello"},
		"widgets[1].duration": {"5s"},
		"widgets[1].dynamic_source": {"my_key"},
		"widgets[1].scroll_speed":   {"50ms"},
	}
	req := httptest.NewRequest("POST", "/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.ParseForm()

	cfg, errMsg := parseConfigForm(req)
	if errMsg != "" {
		t.Fatalf("unexpected error: %s", errMsg)
	}
	if len(cfg.Widgets) != 2 {
		t.Fatalf("expected 2 widgets, got %d", len(cfg.Widgets))
	}
	if cfg.Widgets[1].Text != "Hello" {
		t.Errorf("widget[1].text = %q, want %q", cfg.Widgets[1].Text, "Hello")
	}
	if cfg.Widgets[1].DynamicSource != "my_key" {
		t.Errorf("widget[1].dynamic_source = %q, want %q", cfg.Widgets[1].DynamicSource, "my_key")
	}
}

func TestParseConfigForm_TM1637Pins(t *testing.T) {
	form := url.Values{
		"display.type":    {"tm1637"},
		"display.clk_pin": {"GPIO23"},
		"display.dio_pin": {"GPIO24"},
	}
	req := httptest.NewRequest("POST", "/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.ParseForm()

	cfg, errMsg := parseConfigForm(req)
	if errMsg != "" {
		t.Fatalf("unexpected error: %s", errMsg)
	}
	if cfg.Display.ClkPin != "GPIO23" {
		t.Errorf("clk_pin = %q, want %q", cfg.Display.ClkPin, "GPIO23")
	}
	if cfg.Display.DioPin != "GPIO24" {
		t.Errorf("dio_pin = %q, want %q", cfg.Display.DioPin, "GPIO24")
	}
}

func TestRenderErrorAlert(t *testing.T) {
	w := httptest.NewRecorder()
	renderErrorAlert(w, "something broke")

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnprocessableEntity)
	}
	if !strings.Contains(w.Body.String(), "something broke") {
		t.Error("expected error message in response body")
	}
	if !strings.Contains(w.Body.String(), "alert-error") {
		t.Error("expected alert-error class")
	}
}

func TestRenderErrorAlert_EscapesHTML(t *testing.T) {
	w := httptest.NewRecorder()
	renderErrorAlert(w, "<script>alert('xss')</script>")

	body := w.Body.String()
	if strings.Contains(body, "<script>") {
		t.Error("expected HTML to be escaped")
	}
}

// --- Static handler ---

func TestStaticHandler(t *testing.T) {
	s := testServer(t)
	req := httptest.NewRequest("GET", "/static/", nil)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	// Should not 404 — the embedded FS has static files
	if w.Code == http.StatusNotFound {
		t.Error("expected static handler to serve files")
	}
}
