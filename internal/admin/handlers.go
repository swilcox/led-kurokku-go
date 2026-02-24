package admin

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/swilcox/led-kurokku-go/config"
)

// --- Instance handlers ---

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Instances": s.store.List(),
	}
	if err := renderPage(w, "templates/index.html", data); err != nil {
		log.Printf("render index: %v", err)
	}
}

func (s *Server) handleInstanceNew(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"ID": "", "Name": "", "Host": "", "Port": 6379, "Error": "",
	}
	if err := renderPartial(w, "instance_form", data); err != nil {
		log.Printf("render instance_form: %v", err)
	}
}

func (s *Server) handleInstanceCreate(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	name := strings.TrimSpace(r.FormValue("name"))
	host := strings.TrimSpace(r.FormValue("host"))
	port, _ := strconv.Atoi(r.FormValue("port"))
	if port == 0 {
		port = 6379
	}

	if name == "" || host == "" {
		renderFormError(w, "", name, host, port, "Name and host are required.")
		return
	}

	// Verify connectivity before saving.
	if err := TestConnection(host, port); err != nil {
		renderFormError(w, "", name, host, port, fmt.Sprintf("Redis connection failed: %v", err))
		return
	}

	id := slugify(name)
	if s.store.Get(id) != nil {
		id = id + "-" + strconv.FormatInt(time.Now().UnixMilli(), 36)
	}

	inst := Instance{ID: id, Name: name, Host: host, Port: port}
	if err := s.store.Add(inst); err != nil {
		renderFormError(w, "", name, host, port, fmt.Sprintf("Save failed: %v", err))
		return
	}

	if err := renderPartial(w, "instance_row", inst); err != nil {
		log.Printf("render instance_row: %v", err)
	}
}

func (s *Server) handleInstanceEdit(w http.ResponseWriter, r *http.Request) {
	inst := s.store.Get(r.PathValue("id"))
	if inst == nil {
		http.NotFound(w, r)
		return
	}
	data := map[string]interface{}{
		"ID": inst.ID, "Name": inst.Name, "Host": inst.Host, "Port": inst.Port, "Error": "",
	}
	if err := renderPartial(w, "instance_form", data); err != nil {
		log.Printf("render instance_form: %v", err)
	}
}

func (s *Server) handleInstanceUpdate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	inst := s.store.Get(id)
	if inst == nil {
		http.NotFound(w, r)
		return
	}

	r.ParseForm()
	name := strings.TrimSpace(r.FormValue("name"))
	host := strings.TrimSpace(r.FormValue("host"))
	port, _ := strconv.Atoi(r.FormValue("port"))
	if port == 0 {
		port = 6379
	}

	if name == "" || host == "" {
		renderFormError(w, id, name, host, port, "Name and host are required.")
		return
	}

	inst.Name = name
	inst.Host = host
	inst.Port = port
	if err := s.store.Update(*inst); err != nil {
		renderFormError(w, id, name, host, port, fmt.Sprintf("Save failed: %v", err))
		return
	}

	if err := renderPartial(w, "instance_row", *inst); err != nil {
		log.Printf("render instance_row: %v", err)
	}
}

func (s *Server) handleInstanceDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.store.Remove(id); err != nil {
		http.NotFound(w, r)
		return
	}
	// Return empty body so htmx removes the row.
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleInstanceTest(w http.ResponseWriter, r *http.Request) {
	inst := s.store.Get(r.PathValue("id"))
	if inst == nil {
		http.NotFound(w, r)
		return
	}
	err := TestConnection(inst.Host, inst.Port)
	if err != nil {
		fmt.Fprintf(w, `<span id="status-%s" class="badge badge-error">offline</span>`, inst.ID)
	} else {
		fmt.Fprintf(w, `<span id="status-%s" class="badge badge-success">online</span>`, inst.ID)
	}
}

// --- Config handlers ---

func (s *Server) handleConfigView(w http.ResponseWriter, r *http.Request) {
	inst := s.store.Get(r.PathValue("id"))
	if inst == nil {
		http.NotFound(w, r)
		return
	}

	cfg, found, err := FetchConfig(inst.Host, inst.Port)
	data := map[string]interface{}{
		"Instance":  inst,
		"Config":    cfg,
		"HasConfig": found,
		"Error":     "",
	}
	if err != nil {
		data["Error"] = fmt.Sprintf("Failed to fetch config: %v", err)
	}
	if cfg == nil {
		data["Config"] = &config.Config{}
	}

	if err := renderPage(w, "templates/config_view.html", data); err != nil {
		log.Printf("render config_view: %v", err)
	}
}

func (s *Server) handleConfigEdit(w http.ResponseWriter, r *http.Request) {
	inst := s.store.Get(r.PathValue("id"))
	if inst == nil {
		http.NotFound(w, r)
		return
	}

	cfg, _, err := FetchConfig(inst.Host, inst.Port)
	if cfg == nil {
		cfg = defaultConfig()
	}

	data := map[string]interface{}{
		"Instance": inst,
		"Config":   cfg,
		"Error":    "",
		"Success":  "",
	}
	if err != nil {
		data["Error"] = fmt.Sprintf("Failed to fetch config: %v", err)
	}

	if err := renderPage(w, "templates/config_edit.html", data); err != nil {
		log.Printf("render config_edit: %v", err)
	}
}

func (s *Server) handleConfigSave(w http.ResponseWriter, r *http.Request) {
	inst := s.store.Get(r.PathValue("id"))
	if inst == nil {
		http.NotFound(w, r)
		return
	}

	r.ParseForm()
	cfg, parseErr := parseConfigForm(r)
	if parseErr != "" {
		data := map[string]interface{}{
			"Instance": inst,
			"Config":   cfg,
			"Error":    parseErr,
			"Success":  "",
		}
		renderPage(w, "templates/config_edit.html", data)
		return
	}

	if err := SaveConfig(inst.Host, inst.Port, cfg); err != nil {
		data := map[string]interface{}{
			"Instance": inst,
			"Config":   cfg,
			"Error":    fmt.Sprintf("Failed to save: %v", err),
			"Success":  "",
		}
		renderPage(w, "templates/config_edit.html", data)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/instances/%s/config", inst.ID), http.StatusSeeOther)
}

func (s *Server) handleConfigJSON(w http.ResponseWriter, r *http.Request) {
	inst := s.store.Get(r.PathValue("id"))
	if inst == nil {
		http.NotFound(w, r)
		return
	}

	raw, found, err := FetchConfigJSON(inst.Host, inst.Port)
	if !found {
		raw = "{}"
	}
	// Pretty-print if we can.
	if found {
		var v interface{}
		if json.Unmarshal([]byte(raw), &v) == nil {
			if pretty, err := json.MarshalIndent(v, "", "  "); err == nil {
				raw = string(pretty)
			}
		}
	}

	data := map[string]interface{}{
		"Instance": inst,
		"JSON":     raw,
		"Error":    "",
		"Success":  "",
	}
	if err != nil {
		data["Error"] = fmt.Sprintf("Failed to fetch config: %v", err)
	}

	if renderErr := renderPage(w, "templates/config_json.html", data); renderErr != nil {
		log.Printf("render config_json: %v", renderErr)
	}
}

func (s *Server) handleConfigJSONSave(w http.ResponseWriter, r *http.Request) {
	inst := s.store.Get(r.PathValue("id"))
	if inst == nil {
		http.NotFound(w, r)
		return
	}

	r.ParseForm()
	jsonStr := r.FormValue("json")

	// Validate JSON.
	if _, err := config.Parse([]byte(jsonStr)); err != nil {
		data := map[string]interface{}{
			"Instance": inst,
			"JSON":     jsonStr,
			"Error":    fmt.Sprintf("Invalid JSON: %v", err),
			"Success":  "",
		}
		renderPage(w, "templates/config_json.html", data)
		return
	}

	if err := SaveConfigJSON(inst.Host, inst.Port, jsonStr); err != nil {
		data := map[string]interface{}{
			"Instance": inst,
			"JSON":     jsonStr,
			"Error":    fmt.Sprintf("Failed to save: %v", err),
			"Success":  "",
		}
		renderPage(w, "templates/config_json.html", data)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/instances/%s/config", inst.ID), http.StatusSeeOther)
}

func (s *Server) handleWidgetAdd(w http.ResponseWriter, r *http.Request) {
	countStr := r.URL.Query().Get("count")
	idx, _ := strconv.Atoi(countStr)

	wc := config.WidgetConfig{
		Type:    "clock",
		Enabled: true,
		Duration: config.Duration(10 * time.Second),
	}
	data := map[string]interface{}{
		"Index":  idx,
		"Widget": wc,
	}
	if err := renderPartial(w, "widget_form", data); err != nil {
		log.Printf("render widget_form: %v", err)
	}
}

func (s *Server) handleWidgetRemove(w http.ResponseWriter, r *http.Request) {
	// Return empty response so htmx removes the element.
	w.WriteHeader(http.StatusOK)
}

// --- Helpers ---

func renderFormError(w http.ResponseWriter, id, name, host string, port int, errMsg string) {
	data := map[string]interface{}{
		"ID": id, "Name": name, "Host": host, "Port": port, "Error": errMsg,
	}
	renderPartial(w, "instance_form", data)
}

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		} else if r == ' ' || r == '-' || r == '_' {
			b.WriteRune('-')
		}
	}
	result := b.String()
	// Collapse consecutive hyphens.
	for strings.Contains(result, "--") {
		result = strings.ReplaceAll(result, "--", "-")
	}
	return strings.Trim(result, "-")
}

func defaultConfig() *config.Config {
	t := true
	return &config.Config{
		Display: config.DisplayConfig{Type: config.DisplayMAX7219},
		Brightness: config.BrightnessConfig{
			High: 15,
			Low:  1,
		},
		Widgets: []config.WidgetConfig{
			{
				Type:      "clock",
				Enabled:   true,
				Duration:  config.Duration(10 * time.Second),
				Format24h: &t,
			},
		},
	}
}

func parseConfigForm(r *http.Request) (*config.Config, string) {
	cfg := &config.Config{}

	// Display
	cfg.Display.Type = config.DisplayType(r.FormValue("display.type"))
	cfg.Display.ClkPin = r.FormValue("display.clk_pin")
	cfg.Display.DioPin = r.FormValue("display.dio_pin")
	cfg.Display.I2CBus = r.FormValue("display.i2c_bus")
	cfg.Display.Layout = r.FormValue("display.layout")
	if addrStr := r.FormValue("display.i2c_addr"); addrStr != "" {
		addrStr = strings.TrimPrefix(addrStr, "0x")
		addrStr = strings.TrimPrefix(addrStr, "0X")
		val, err := strconv.ParseUint(addrStr, 16, 16)
		if err != nil {
			return cfg, fmt.Sprintf("Invalid I2C address: %v", err)
		}
		cfg.Display.I2CAddr = uint16(val)
	}

	// Brightness
	high, _ := strconv.Atoi(r.FormValue("brightness.high"))
	low, _ := strconv.Atoi(r.FormValue("brightness.low"))
	cfg.Brightness.High = byte(high)
	cfg.Brightness.Low = byte(low)
	cfg.Brightness.DayStart = r.FormValue("brightness.day_start")
	cfg.Brightness.DayEnd = r.FormValue("brightness.day_end")
	cfg.Brightness.UseLocation = r.FormValue("brightness.use_location") == "on"

	// Location
	latStr := r.FormValue("location.lat")
	lonStr := r.FormValue("location.lon")
	tz := r.FormValue("location.timezone")
	if latStr != "" || lonStr != "" || tz != "" {
		lat, _ := strconv.ParseFloat(latStr, 64)
		lon, _ := strconv.ParseFloat(lonStr, 64)
		cfg.Location = &config.LocationConfig{
			Lat:      lat,
			Lon:      lon,
			Timezone: tz,
		}
	}

	// Widgets
	for i := 0; ; i++ {
		prefix := fmt.Sprintf("widgets[%d].", i)
		wType := r.FormValue(prefix + "type")
		if wType == "" {
			break
		}
		wc := config.WidgetConfig{
			Type:          wType,
			Enabled:       r.FormValue(prefix+"enabled") == "on",
			Text:          r.FormValue(prefix + "text"),
			DynamicSource: r.FormValue(prefix + "dynamic_source"),
			Cron:          r.FormValue(prefix + "cron"),
			AnimationType: r.FormValue(prefix + "animation_type"),
		}

		if durStr := r.FormValue(prefix + "duration"); durStr != "" {
			dur, err := time.ParseDuration(durStr)
			if err != nil {
				return cfg, fmt.Sprintf("Widget #%d: invalid duration %q", i+1, durStr)
			}
			wc.Duration = config.Duration(dur)
		}
		if ssStr := r.FormValue(prefix + "scroll_speed"); ssStr != "" {
			dur, err := time.ParseDuration(ssStr)
			if err != nil {
				return cfg, fmt.Sprintf("Widget #%d: invalid scroll speed %q", i+1, ssStr)
			}
			wc.ScrollSpeed = config.Duration(dur)
		}
		if fdStr := r.FormValue(prefix + "frame_duration"); fdStr != "" {
			dur, err := time.ParseDuration(fdStr)
			if err != nil {
				return cfg, fmt.Sprintf("Widget #%d: invalid frame duration %q", i+1, fdStr)
			}
			wc.FrameDuration = config.Duration(dur)
		}
		if r.FormValue(prefix+"format_24h") == "on" {
			t := true
			wc.Format24h = &t
		}
		if repStr := r.FormValue(prefix + "repeats"); repStr != "" {
			rep, err := strconv.Atoi(repStr)
			if err != nil {
				return cfg, fmt.Sprintf("Widget #%d: invalid repeats %q", i+1, repStr)
			}
			wc.Repeats = &rep
		}

		cfg.Widgets = append(cfg.Widgets, wc)
	}

	return cfg, ""
}
