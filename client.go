// Package seabhac provides a minimal client for the Seabhac API.
package seabhac

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const defaultBase = "https://api.seabhac.io"

type Client struct {
	apiKey  string
	base    string
	http    *http.Client
}

func New(apiKey string) *Client {
	return &Client{apiKey: apiKey, base: defaultBase, http: &http.Client{Timeout: 30 * time.Second}}
}

func (c *Client) WithBaseURL(u string) *Client { c.base = u; return c }

func (c *Client) get(path string, params url.Values, out any) error {
	u := c.base + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Api-Key", c.apiKey)
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		var e struct{ Error string `json:"error"` }
		json.NewDecoder(resp.Body).Decode(&e)
		return fmt.Errorf("API %d: %s", resp.StatusCode, e.Error)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// --- schedules ---

func (c *Client) ListSchedules() ([]map[string]any, error) {
	var r struct{ Data []map[string]any `json:"data"` }
	return r.Data, c.get("/v1/schedules", nil, &r)
}

func (c *Client) GetSchedule(id string) (map[string]any, error) {
	var r struct{ Data map[string]any `json:"data"` }
	return r.Data, c.get("/v1/schedules/"+id, nil, &r)
}

// --- jobs ---

func (c *Client) ListJobs(scheduleID string, limit, offset int) ([]map[string]any, error) {
	var r struct{ Data []map[string]any `json:"data"` }
	p := url.Values{"limit": {fmt.Sprint(limit)}, "offset": {fmt.Sprint(offset)}}
	return r.Data, c.get("/v1/schedules/"+scheduleID+"/jobs", p, &r)
}

func (c *Client) GetJob(scheduleID, jobID string) (map[string]any, error) {
	var r struct{ Data map[string]any `json:"data"` }
	return r.Data, c.get("/v1/schedules/"+scheduleID+"/jobs/"+jobID, nil, &r)
}

// --- alerts ---

func (c *Client) ListAlerts(scheduleID string) ([]map[string]any, error) {
	var r struct{ Data []map[string]any `json:"data"` }
	return r.Data, c.get("/v1/schedules/"+scheduleID+"/alerts", nil, &r)
}

// --- metrics ---

func (c *Client) metrics(scheduleID, kind string, from, to time.Time) ([]map[string]any, error) {
	var r struct{ Data []map[string]any `json:"data"` }
	p := url.Values{}
	if !from.IsZero() {
		p.Set("from", from.UTC().Format(time.RFC3339))
	}
	if !to.IsZero() {
		p.Set("to", to.UTC().Format(time.RFC3339))
	}
	return r.Data, c.get("/v1/schedules/"+scheduleID+"/metrics/"+kind, p, &r)
}

func (c *Client) MetricsHTTP(id string, from, to time.Time) ([]map[string]any, error) {
	return c.metrics(id, "http", from, to)
}
func (c *Client) MetricsDNS(id string, from, to time.Time) ([]map[string]any, error) {
	return c.metrics(id, "dns", from, to)
}
func (c *Client) MetricsSSL(id string, from, to time.Time) ([]map[string]any, error) {
	return c.metrics(id, "ssl", from, to)
}
func (c *Client) MetricsEmailAuth(id string, from, to time.Time) ([]map[string]any, error) {
	return c.metrics(id, "email_auth", from, to)
}
func (c *Client) MetricsPageLoad(id string, from, to time.Time) ([]map[string]any, error) {
	return c.metrics(id, "pageload", from, to)
}
func (c *Client) MetricsSSH(id string, from, to time.Time) ([]map[string]any, error) {
	return c.metrics(id, "ssh", from, to)
}

// --- dmarc ---

func (c *Client) DMARCListReports(limit, offset int) (map[string]any, error) {
	var r map[string]any
	p := url.Values{"limit": {fmt.Sprint(limit)}, "offset": {fmt.Sprint(offset)}}
	return r, c.get("/v1/dmarc/reports", p, &r)
}

func (c *Client) DMARCGetReport(id string) (map[string]any, error) {
	var r map[string]any
	return r, c.get("/v1/dmarc/reports/"+id, nil, &r)
}

func (c *Client) dmarcQuery(endpoint, domain string) ([]map[string]any, error) {
	var r struct{ Data []map[string]any `json:"data"` }
	p := url.Values{}
	if domain != "" {
		p.Set("domain", domain)
	}
	return r.Data, c.get("/v1/dmarc/"+endpoint, p, &r)
}

func (c *Client) DMARCMetrics(domain string) ([]map[string]any, error) {
	return c.dmarcQuery("metrics", domain)
}
func (c *Client) DMARCGeo(domain string) ([]map[string]any, error) {
	return c.dmarcQuery("geo", domain)
}
func (c *Client) DMARCTopIPs(domain string) ([]map[string]any, error) {
	return c.dmarcQuery("top-ips", domain)
}
func (c *Client) DMARCTopFailingIPs(domain string) ([]map[string]any, error) {
	return c.dmarcQuery("top-failing-ips", domain)
}
func (c *Client) DMARCTopSenders(domain string) ([]map[string]any, error) {
	return c.dmarcQuery("top-senders", domain)
}
func (c *Client) DMARCFailReasons(domain string) (map[string]any, error) {
	var r struct{ Data map[string]any `json:"data"` }
	p := url.Values{}
	if domain != "" {
		p.Set("domain", domain)
	}
	return r.Data, c.get("/v1/dmarc/fail-reasons", p, &r)
}
func (c *Client) DMARCReporters(domain string) ([]map[string]any, error) {
	return c.dmarcQuery("reporters", domain)
}
func (c *Client) DMARCTopASNs(domain string) ([]map[string]any, error) {
	return c.dmarcQuery("top-asns", domain)
}
