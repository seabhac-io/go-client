package seabhac_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	seabhac "github.com/seabhac-io/go-client"
)

// serve returns a test server that responds with the given payload and records
// the last request for inspection.
func serve(t *testing.T, payload any) (client *seabhac.Client, lastReq func() *http.Request) {
	t.Helper()
	var req *http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req = r
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(payload)
	}))
	t.Cleanup(srv.Close)
	c := seabhac.New("test-key").WithBaseURL(srv.URL)
	return c, func() *http.Request { return req }
}

// serveError returns a test server that responds with a 4xx error.
func serveError(t *testing.T, status int, msg string) *seabhac.Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(map[string]string{"error": msg})
	}))
	t.Cleanup(srv.Close)
	return seabhac.New("test-key").WithBaseURL(srv.URL)
}

func mustTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

// --- auth ---

func TestAPIKeyHeader(t *testing.T) {
	c, lastReq := serve(t, map[string]any{"data": []any{}})
	c.ListSchedules()
	if got := lastReq().Header.Get("X-Api-Key"); got != "test-key" {
		t.Fatalf("expected X-Api-Key: test-key, got %q", got)
	}
}

// --- schedules ---

func TestListSchedules(t *testing.T) {
	payload := map[string]any{
		"data": []any{
			map[string]any{
				"id": "sched-1", "user_id": "user-1", "name": "My Schedule",
				"type": "http", "status": "active", "cron_expr": "*/5 * * * *",
				"created_at": "2024-01-01T00:00:00Z", "updated_at": "2024-01-01T00:00:00Z",
				"config": map[string]any{}, "tags": []any{},
			},
		},
	}
	c, lastReq := serve(t, payload)
	schedules, err := c.ListSchedules()
	if err != nil {
		t.Fatal(err)
	}
	if len(schedules) != 1 {
		t.Fatalf("expected 1 schedule, got %d", len(schedules))
	}
	s := schedules[0]
	if s.ID != "sched-1" {
		t.Errorf("ID: want sched-1, got %s", s.ID)
	}
	if s.Type != seabhac.ScheduleTypeHTTP {
		t.Errorf("Type: want http, got %s", s.Type)
	}
	if s.Status != seabhac.ScheduleStatusActive {
		t.Errorf("Status: want active, got %s", s.Status)
	}
	if lastReq().URL.Path != "/v1/schedules" {
		t.Errorf("path: want /v1/schedules, got %s", lastReq().URL.Path)
	}
}

func TestGetSchedule(t *testing.T) {
	payload := map[string]any{
		"data": map[string]any{
			"id": "sched-abc", "user_id": "user-1", "name": "DNS Check",
			"type": "dns", "status": "paused", "cron_expr": "@hourly",
			"created_at": "2024-06-01T12:00:00Z", "updated_at": "2024-06-01T12:00:00Z",
			"config": map[string]any{"domain_name": "example.com"},
			"tags":   []any{"prod"},
		},
	}
	c, lastReq := serve(t, payload)
	s, err := c.GetSchedule("sched-abc")
	if err != nil {
		t.Fatal(err)
	}
	if s.ID != "sched-abc" {
		t.Errorf("ID: want sched-abc, got %s", s.ID)
	}
	if s.Config.DomainName != "example.com" {
		t.Errorf("Config.DomainName: want example.com, got %s", s.Config.DomainName)
	}
	if lastReq().URL.Path != "/v1/schedules/sched-abc" {
		t.Errorf("path: want /v1/schedules/sched-abc, got %s", lastReq().URL.Path)
	}
}

func TestGetSchedule_NullableFields(t *testing.T) {
	payload := map[string]any{
		"data": map[string]any{
			"id": "s1", "user_id": "u1", "name": "n", "type": "http", "status": "active",
			"cron_expr": "@daily", "config": map[string]any{}, "tags": []any{},
			"created_at":      "2024-01-01T00:00:00Z",
			"updated_at":      "2024-01-01T00:00:00Z",
			"organization_id": nil,
			"region_id":       nil,
			"last_run_at":     nil,
			"next_run_at":     nil,
		},
	}
	c, _ := serve(t, payload)
	s, err := c.GetSchedule("s1")
	if err != nil {
		t.Fatal(err)
	}
	if s.OrganizationID != nil {
		t.Errorf("OrganizationID should be nil")
	}
	if s.LastRunAt != nil {
		t.Errorf("LastRunAt should be nil")
	}
}

// --- jobs ---

func TestListJobs(t *testing.T) {
	payload := map[string]any{
		"data": []any{
			map[string]any{
				"id": "job-1", "schedule_id": "sched-1", "user_id": "user-1",
				"type": "http", "status": "completed",
				"created_at": "2024-01-01T00:00:00Z",
				"results": map[string]any{
					"summary": map[string]any{
						"total_checks": 1, "successful_checks": 1, "failed_checks": 0, "execution_time_ms": 120,
					},
				},
			},
		},
	}
	c, lastReq := serve(t, payload)
	jobs, err := c.ListJobs("sched-1", 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}
	j := jobs[0]
	if j.Status != seabhac.JobStatusCompleted {
		t.Errorf("Status: want completed, got %s", j.Status)
	}
	if j.Results.Summary.TotalChecks != 1 {
		t.Errorf("Summary.TotalChecks: want 1, got %d", j.Results.Summary.TotalChecks)
	}
	q := lastReq().URL.Query()
	if q.Get("limit") != "10" || q.Get("offset") != "0" {
		t.Errorf("query params: want limit=10&offset=0, got %s", lastReq().URL.RawQuery)
	}
}

func TestGetJob_HTTPResults(t *testing.T) {
	payload := map[string]any{
		"data": map[string]any{
			"id": "job-2", "schedule_id": "sched-1", "user_id": "user-1",
			"type": "http", "status": "completed",
			"created_at": "2024-01-01T00:00:00Z",
			"results": map[string]any{
				"summary": map[string]any{"total_checks": 1, "successful_checks": 1, "failed_checks": 0, "execution_time_ms": 80},
				"http_results": map[string]any{
					"status_code": 200, "response_time_ms": 95, "success": true,
					"checked_at": "2024-01-01T00:00:00Z",
				},
			},
		},
	}
	c, _ := serve(t, payload)
	j, err := c.GetJob("sched-1", "job-2")
	if err != nil {
		t.Fatal(err)
	}
	if j.Results.HTTPResults == nil {
		t.Fatal("HTTPResults should not be nil")
	}
	if j.Results.HTTPResults.StatusCode != 200 {
		t.Errorf("StatusCode: want 200, got %d", j.Results.HTTPResults.StatusCode)
	}
	if j.Results.HTTPResults.ResponseTimeMs != 95 {
		t.Errorf("ResponseTimeMs: want 95, got %d", j.Results.HTTPResults.ResponseTimeMs)
	}
}

func TestGetJob_DNSBLResults(t *testing.T) {
	payload := map[string]any{
		"data": map[string]any{
			"id": "job-3", "schedule_id": "sched-1", "user_id": "user-1",
			"type": "dnsbl", "status": "completed",
			"created_at": "2024-01-01T00:00:00Z",
			"results": map[string]any{
				"summary": map[string]any{"total_checks": 2, "successful_checks": 2, "failed_checks": 0, "execution_time_ms": 50},
				"dnsbl_results": map[string]any{
					"zen.spamhaus.org": map[string]any{
						"server": "zen.spamhaus.org", "listed": false, "response": "",
						"checked_at": "2024-01-01T00:00:00Z",
					},
				},
			},
		},
	}
	c, _ := serve(t, payload)
	j, err := c.GetJob("sched-1", "job-3")
	if err != nil {
		t.Fatal(err)
	}
	r, ok := j.Results.DNSBLResults["zen.spamhaus.org"]
	if !ok {
		t.Fatal("missing zen.spamhaus.org in DNSBLResults")
	}
	if r.Listed {
		t.Error("Listed should be false")
	}
}

func TestGetJob_SSLResult(t *testing.T) {
	payload := map[string]any{
		"data": map[string]any{
			"id": "job-4", "schedule_id": "sched-1", "user_id": "user-1",
			"type": "ssl", "status": "completed",
			"created_at": "2024-01-01T00:00:00Z",
			"results": map[string]any{
				"summary": map[string]any{"total_checks": 1, "successful_checks": 1, "failed_checks": 0, "execution_time_ms": 30},
				"ssl_result": map[string]any{
					"domain": "example.com", "expiry": "2025-01-01T00:00:00Z",
					"is_valid": true, "is_self_signed": false,
					"dns_names": []any{"example.com", "www.example.com"},
					"issuer": "Let's Encrypt", "subject": "example.com",
					"cipher_suite": "TLS_AES_256_GCM_SHA384", "protocol_version": "TLSv1.3",
					"response_time": 28000000, "checked_at": "2024-01-01T00:00:00Z",
				},
			},
		},
	}
	c, _ := serve(t, payload)
	j, err := c.GetJob("sched-1", "job-4")
	if err != nil {
		t.Fatal(err)
	}
	if j.Results.SSLResult == nil {
		t.Fatal("SSLResult should not be nil")
	}
	if j.Results.SSLResult.Domain != "example.com" {
		t.Errorf("Domain: want example.com, got %s", j.Results.SSLResult.Domain)
	}
	if len(j.Results.SSLResult.DNSNames) != 2 {
		t.Errorf("DNSNames: want 2, got %d", len(j.Results.SSLResult.DNSNames))
	}
}

// --- alerts ---

func TestListAlerts(t *testing.T) {
	payload := map[string]any{
		"data": []any{
			map[string]any{
				"id": "alert-1", "user_id": "user-1", "schedule_id": "sched-1",
				"name": "High latency", "metric": "response_time_ms",
				"condition": "gt", "threshold": 500.0,
				"consecutive_count": 3, "cooldown_minutes": 30, "is_enabled": true,
				"created_at": "2024-01-01T00:00:00Z", "updated_at": "2024-01-01T00:00:00Z",
			},
		},
	}
	c, lastReq := serve(t, payload)
	alerts, err := c.ListAlerts("sched-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}
	a := alerts[0]
	if a.Metric != "response_time_ms" {
		t.Errorf("Metric: want response_time_ms, got %s", a.Metric)
	}
	if a.Threshold != 500.0 {
		t.Errorf("Threshold: want 500.0, got %f", a.Threshold)
	}
	if lastReq().URL.Path != "/v1/schedules/sched-1/alerts" {
		t.Errorf("path: want /v1/schedules/sched-1/alerts, got %s", lastReq().URL.Path)
	}
}

// --- metrics ---

func TestMetricsHTTP(t *testing.T) {
	payload := map[string]any{
		"data": []any{
			map[string]any{
				"timestamp": "2024-01-01T00:00:00Z",
				"avg_latency_ms": 120.5, "max_latency_ms": 200.0, "min_latency_ms": 80.0,
				"total_requests": 100, "successful_requests": 98, "uptime_percent": 98.0,
			},
		},
	}
	from := mustTime("2024-01-01T00:00:00Z")
	to := mustTime("2024-01-02T00:00:00Z")

	c, lastReq := serve(t, payload)
	pts, err := c.MetricsHTTP("sched-1", from, to)
	if err != nil {
		t.Fatal(err)
	}
	if len(pts) != 1 {
		t.Fatalf("expected 1 point, got %d", len(pts))
	}
	p := pts[0]
	if p.AvgLatencyMs != 120.5 {
		t.Errorf("AvgLatencyMs: want 120.5, got %f", p.AvgLatencyMs)
	}
	if p.UptimePercent != 98.0 {
		t.Errorf("UptimePercent: want 98.0, got %f", p.UptimePercent)
	}
	q := lastReq().URL.Query()
	if q.Get("from") == "" || q.Get("to") == "" {
		t.Error("from/to query params should be set")
	}
	if lastReq().URL.Path != "/v1/schedules/sched-1/metrics/http" {
		t.Errorf("path: want /v1/schedules/sched-1/metrics/http, got %s", lastReq().URL.Path)
	}
}

func TestMetricsSSL(t *testing.T) {
	payload := map[string]any{
		"data": []any{
			map[string]any{
				"timestamp": "2024-01-01T00:00:00Z",
				"avg_latency_ms": 10.0, "max_latency_ms": 15.0, "min_latency_ms": 8.0,
				"total_requests": 10, "successful_requests": 10, "uptime_percent": 100.0,
				"avg_days_until_expiry": 87.3,
			},
		},
	}
	c, _ := serve(t, payload)
	pts, err := c.MetricsSSL("sched-1", time.Time{}, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if pts[0].AvgDaysUntilExpiry != 87.3 {
		t.Errorf("AvgDaysUntilExpiry: want 87.3, got %f", pts[0].AvgDaysUntilExpiry)
	}
}

func TestMetricsEmailAuth(t *testing.T) {
	payload := map[string]any{
		"data": []any{
			map[string]any{
				"timestamp": "2024-01-01T00:00:00Z",
				"avg_latency_ms": 50.0, "max_latency_ms": 60.0, "min_latency_ms": 40.0,
				"total_requests": 5, "successful_requests": 5, "uptime_percent": 100.0,
				"spf_valid_count": 5, "dmarc_valid_count": 5, "dkim_valid_count": 4,
			},
		},
	}
	c, _ := serve(t, payload)
	pts, err := c.MetricsEmailAuth("sched-1", time.Time{}, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if pts[0].DKIMValidCount != 4 {
		t.Errorf("DKIMValidCount: want 4, got %d", pts[0].DKIMValidCount)
	}
}

func TestMetricsPageLoad(t *testing.T) {
	payload := map[string]any{
		"data": []any{
			map[string]any{
				"timestamp": "2024-01-01T00:00:00Z",
				"avg_response_time_ms": 350.0, "max_response_time_ms": 500.0, "min_response_time_ms": 200.0,
				"avg_ttfb_ms": 80.0, "avg_dom_load_ms": 200.0, "avg_page_load_ms": 340.0,
				"total_requests": 20, "successful_requests": 20, "uptime_percent": 100.0,
			},
		},
	}
	c, _ := serve(t, payload)
	pts, err := c.MetricsPageLoad("sched-1", time.Time{}, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if pts[0].AvgTTFBMs != 80.0 {
		t.Errorf("AvgTTFBMs: want 80.0, got %f", pts[0].AvgTTFBMs)
	}
}

func TestMetrics_ZeroTimesOmitsQueryParams(t *testing.T) {
	c, lastReq := serve(t, map[string]any{"data": []any{}})
	c.MetricsHTTP("sched-1", time.Time{}, time.Time{})
	q := lastReq().URL.Query()
	if q.Get("from") != "" || q.Get("to") != "" {
		t.Errorf("from/to should be absent when zero, got %s", lastReq().URL.RawQuery)
	}
}

// --- dmarc ---

func TestDMARCListReports(t *testing.T) {
	payload := map[string]any{
		"data":   []any{map[string]any{"id": "r1"}},
		"count":  42,
		"limit":  10,
		"offset": 0,
	}
	c, lastReq := serve(t, payload)
	result, err := c.DMARCListReports(10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if result.Count != 42 {
		t.Errorf("Count: want 42, got %d", result.Count)
	}
	if len(result.Data) != 1 {
		t.Errorf("Data length: want 1, got %d", len(result.Data))
	}
	q := lastReq().URL.Query()
	if q.Get("limit") != "10" || q.Get("offset") != "0" {
		t.Errorf("query params: want limit=10&offset=0, got %s", lastReq().URL.RawQuery)
	}
}

func TestDMARCGetReport(t *testing.T) {
	payload := map[string]any{
		"report":  map[string]any{"id": "r1", "org_name": "Google"},
		"records": []any{map[string]any{"source_ip": "1.2.3.4"}},
	}
	c, lastReq := serve(t, payload)
	result, err := c.DMARCGetReport("r1")
	if err != nil {
		t.Fatal(err)
	}
	if result.Report["org_name"] != "Google" {
		t.Errorf("Report.org_name: want Google, got %v", result.Report["org_name"])
	}
	if len(result.Records) != 1 {
		t.Errorf("Records length: want 1, got %d", len(result.Records))
	}
	if lastReq().URL.Path != "/v1/dmarc/reports/r1" {
		t.Errorf("path: want /v1/dmarc/reports/r1, got %s", lastReq().URL.Path)
	}
}

func TestDMARCMetrics(t *testing.T) {
	payload := map[string]any{
		"data": []any{
			map[string]any{
				"timestamp": "2024-01-01", "domain": "example.com",
				"total_messages": 1000, "dkim_pass": 980, "dkim_fail": 20,
				"spf_pass": 990, "spf_fail": 10,
				"disposition_none": 950, "disposition_quarantine": 30, "disposition_reject": 20,
				"dmarc_pass": 975,
			},
		},
	}
	c, lastReq := serve(t, payload)
	pts, err := c.DMARCMetrics("example.com")
	if err != nil {
		t.Fatal(err)
	}
	if len(pts) != 1 {
		t.Fatalf("expected 1 point, got %d", len(pts))
	}
	p := pts[0]
	if p.TotalMessages != 1000 {
		t.Errorf("TotalMessages: want 1000, got %d", p.TotalMessages)
	}
	if p.DMARCPass != 975 {
		t.Errorf("DMARCPass: want 975, got %d", p.DMARCPass)
	}
	if q := lastReq().URL.Query(); q.Get("domain") != "example.com" {
		t.Errorf("domain query param: want example.com, got %s", q.Get("domain"))
	}
}

func TestDMARCMetrics_NoDomain(t *testing.T) {
	c, lastReq := serve(t, map[string]any{"data": []any{}})
	c.DMARCMetrics("")
	if q := lastReq().URL.Query(); q.Has("domain") {
		t.Error("domain param should be absent when empty")
	}
}

func TestDMARCAnalyticsEndpoints(t *testing.T) {
	payload := map[string]any{"data": []any{map[string]any{"country": "US", "count": 100}}}

	tests := []struct {
		name string
		fn   func(*seabhac.Client) ([]map[string]any, error)
		path string
	}{
		{"geo", func(c *seabhac.Client) ([]map[string]any, error) { return c.DMARCGeo("example.com") }, "/v1/dmarc/geo"},
		{"top-ips", func(c *seabhac.Client) ([]map[string]any, error) { return c.DMARCTopIPs("example.com") }, "/v1/dmarc/top-ips"},
		{"top-failing-ips", func(c *seabhac.Client) ([]map[string]any, error) { return c.DMARCTopFailingIPs("example.com") }, "/v1/dmarc/top-failing-ips"},
		{"top-senders", func(c *seabhac.Client) ([]map[string]any, error) { return c.DMARCTopSenders("example.com") }, "/v1/dmarc/top-senders"},
		{"reporters", func(c *seabhac.Client) ([]map[string]any, error) { return c.DMARCReporters("example.com") }, "/v1/dmarc/reporters"},
		{"top-asns", func(c *seabhac.Client) ([]map[string]any, error) { return c.DMARCTopASNs("example.com") }, "/v1/dmarc/top-asns"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, lastReq := serve(t, payload)
			data, err := tt.fn(c)
			if err != nil {
				t.Fatal(err)
			}
			if len(data) != 1 {
				t.Errorf("want 1 item, got %d", len(data))
			}
			if lastReq().URL.Path != tt.path {
				t.Errorf("path: want %s, got %s", tt.path, lastReq().URL.Path)
			}
			if q := lastReq().URL.Query(); q.Get("domain") != "example.com" {
				t.Errorf("domain param: want example.com, got %s", q.Get("domain"))
			}
		})
	}
}

func TestDMARCFailReasons(t *testing.T) {
	payload := map[string]any{
		"data": map[string]any{"spf_only": 10, "dkim_only": 5, "both": 2},
	}
	c, _ := serve(t, payload)
	data, err := c.DMARCFailReasons("example.com")
	if err != nil {
		t.Fatal(err)
	}
	if data["spf_only"].(float64) != 10 {
		t.Errorf("spf_only: want 10, got %v", data["spf_only"])
	}
}

// --- error handling ---

func TestAPIError(t *testing.T) {
	c := serveError(t, 404, "schedule not found")
	_, err := c.GetSchedule("missing")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "API 404: schedule not found" {
		t.Errorf("error message: got %q", err.Error())
	}
}

func TestAPIError_Unauthorized(t *testing.T) {
	c := serveError(t, 401, "invalid api key")
	_, err := c.ListSchedules()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- WithBaseURL ---

func TestWithBaseURL(t *testing.T) {
	var called bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		json.NewEncoder(w).Encode(map[string]any{"data": []any{}})
	}))
	defer srv.Close()

	seabhac.New("k").WithBaseURL(srv.URL).ListSchedules()
	if !called {
		t.Error("custom base URL was not used")
	}
}

// --- ensure url.Values is used correctly (compile-time check via usage) ---

var _ = url.Values{}
