// Package seabhac provides a client for the Seabhac API.
package seabhac

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const defaultBase = "https://api.seabhac.io"

// --- enums ---

type ScheduleType string

const (
	ScheduleTypeDNSBL        ScheduleType = "dnsbl"
	ScheduleTypeEHLO         ScheduleType = "ehlo"
	ScheduleTypeDNS          ScheduleType = "dns"
	ScheduleTypeHTTP         ScheduleType = "http"
	ScheduleTypeEmailAuth    ScheduleType = "email_auth"
	ScheduleTypeTCP          ScheduleType = "tcp"
	ScheduleTypeUDP          ScheduleType = "udp"
	ScheduleTypePageLoad     ScheduleType = "pageload"
	ScheduleTypeSSL          ScheduleType = "ssl"
	ScheduleTypeSSH          ScheduleType = "ssh"
	ScheduleTypeBrokenLinks  ScheduleType = "broken_links"
	ScheduleTypeDomainExpiry ScheduleType = "domain_expiry"
	ScheduleTypeDMARC        ScheduleType = "dmarc"
)

type ScheduleStatus string

const (
	ScheduleStatusActive   ScheduleStatus = "active"
	ScheduleStatusInactive ScheduleStatus = "inactive"
	ScheduleStatusPaused   ScheduleStatus = "paused"
)

type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusCancelled JobStatus = "cancelled"
	JobStatusRetrying  JobStatus = "retrying"
	JobStatusPaused    JobStatus = "paused"
)

// --- domain types ---

type ScheduleConfig struct {
	Hostname        string         `json:"hostname,omitempty"`
	DNSBLServers    []string       `json:"dnsbl_servers,omitempty"`
	DomainName      string         `json:"domain_name,omitempty"`
	RecordTypes     []string       `json:"record_types,omitempty"`
	DNSServers      []string       `json:"dns_servers,omitempty"`
	ServerHost      string         `json:"server_host,omitempty"`
	ServerPort      int            `json:"server_port,omitempty"`
	Host            string         `json:"host,omitempty"`
	Port            int            `json:"port,omitempty"`
	Proto           string         `json:"proto,omitempty"`
	ExpectedBanner  string         `json:"expected_banner,omitempty"`
	Timeout         int            `json:"timeout,omitempty"`
	RetryCount      int            `json:"retry_count,omitempty"`
	NotifyOnFailure bool           `json:"notify_on_failure,omitempty"`
	BaselineEnabled bool           `json:"baseline_enabled,omitempty"`
	RunAllRegions   bool           `json:"run_all_regions,omitempty"`
	CustomFields    map[string]any `json:"custom_fields,omitempty"`
}

type Schedule struct {
	ID               string         `json:"id"`
	UserID           string         `json:"user_id"`
	OrganizationID   *string        `json:"organization_id"`
	Name             string         `json:"name"`
	Description      string         `json:"description"`
	Type             ScheduleType   `json:"type"`
	Status           ScheduleStatus `json:"status"`
	RegionID         *string        `json:"region_id"`
	CronExpr         string         `json:"cron_expr"`
	Config           ScheduleConfig `json:"config"`
	BaselineEnabled  bool           `json:"baseline_enabled"`
	ShowOnStatusPage bool           `json:"show_on_status_page"`
	Tags             []string       `json:"tags"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	LastRunAt        *time.Time     `json:"last_run_at"`
	NextRunAt        *time.Time     `json:"next_run_at"`
}

type JobSummary struct {
	TotalChecks      int   `json:"total_checks"`
	SuccessfulChecks int   `json:"successful_checks"`
	FailedChecks     int   `json:"failed_checks"`
	ExecutionTimeMs  int64 `json:"execution_time_ms"`
}

type DNSBLResult struct {
	Server    string    `json:"server"`
	Listed    bool      `json:"listed"`
	Response  string    `json:"response"`
	Error     string    `json:"error"`
	CheckedAt time.Time `json:"checked_at"`
}

type DNSRecord struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value"`
	TTL   int    `json:"ttl"`
}

type EHLOResult struct {
	Connected    bool      `json:"connected"`
	EHLOResponse string    `json:"ehlo_response"`
	Extensions   []string  `json:"extensions"`
	Error        string    `json:"error"`
	CheckedAt    time.Time `json:"checked_at"`
}

type HTTPResult struct {
	StatusCode     int       `json:"status_code"`
	ResponseTimeMs int64     `json:"response_time_ms"`
	Success        bool      `json:"success"`
	Error          string    `json:"error"`
	CheckedAt      time.Time `json:"checked_at"`
}

type EmailAuthResult struct {
	SPFRecord      string    `json:"spf_record"`
	SPFValid       bool      `json:"spf_valid"`
	SPFError       string    `json:"spf_error"`
	DMARCRecord    string    `json:"dmarc_record"`
	DMARCValid     bool      `json:"dmarc_valid"`
	DMARCError     string    `json:"dmarc_error"`
	DKIMValid      bool      `json:"dkim_valid"`
	DKIMError      string    `json:"dkim_error"`
	ResponseTimeMs int64     `json:"response_time_ms"`
	Success        bool      `json:"success"`
	CheckedAt      time.Time `json:"checked_at"`
}

type TCPResult struct {
	Connected      bool      `json:"connected"`
	ResponseTimeMs int64     `json:"response_time_ms"`
	Error          *string   `json:"error"`
	CheckedAt      time.Time `json:"checked_at"`
}

type UDPResult struct {
	Reachable      bool      `json:"reachable"`
	ResponseTimeMs int64     `json:"response_time_ms"`
	Error          *string   `json:"error"`
	CheckedAt      time.Time `json:"checked_at"`
}

type SSLResult struct {
	Domain          string    `json:"domain"`
	Expiry          time.Time `json:"expiry"`
	IsValid         bool      `json:"is_valid"`
	IsSelfSigned    bool      `json:"is_self_signed"`
	DNSNames        []string  `json:"dns_names"`
	Issuer          string    `json:"issuer"`
	Subject         string    `json:"subject"`
	CipherSuite     string    `json:"cipher_suite"`
	ProtocolVersion string    `json:"protocol_version"`
	ResponseTime    int64     `json:"response_time"`
	Error           *string   `json:"error"`
	CheckedAt       time.Time `json:"checked_at"`
}

type SSHResult struct {
	Host         string    `json:"host"`
	Port         int       `json:"port"`
	Connected    bool      `json:"connected"`
	Banner       string    `json:"banner"`
	Version      string    `json:"version"`
	BannerMatch  bool      `json:"banner_match"`
	ResponseTime int64     `json:"response_time"`
	Success      bool      `json:"success"`
	Error        *string   `json:"error"`
	CheckedAt    time.Time `json:"checked_at"`
}

type PageLoadResult struct {
	URL            string    `json:"url"`
	ResponseTime   int64     `json:"response_time"`
	TTFB           int64     `json:"ttfb"`
	DOMContentLoad int64     `json:"dom_content_load"`
	PageLoad       int64     `json:"page_load"`
	FCP            int64     `json:"fcp"`
	LCP            int64     `json:"lcp"`
	CLS            float64   `json:"cls"`
	ResourceCount  int       `json:"resource_count"`
	TransferSize   int64     `json:"transfer_size"`
	ConsoleErrors  []string  `json:"console_errors"`
	HTTPStatusCode int       `json:"http_status_code"`
	Success        bool      `json:"success"`
	Error          *string   `json:"error"`
	CheckedAt      time.Time `json:"checked_at"`
}

type BrokenLink struct {
	URL        string `json:"url"`
	StatusCode int    `json:"status_code"`
	Error      string `json:"error"`
}

type BrokenLinksResult struct {
	TargetURL   string       `json:"target_url"`
	TotalLinks  int          `json:"total_links"`
	BrokenCount int          `json:"broken_count"`
	BrokenLinks []BrokenLink `json:"broken_links"`
	Success     bool         `json:"success"`
	Error       *string      `json:"error"`
	CheckedAt   time.Time    `json:"checked_at"`
}

type DomainExpiryResult struct {
	Domain          string    `json:"domain"`
	ExpiryDate      time.Time `json:"expiry_date"`
	DaysUntilExpiry int       `json:"days_until_expiry"`
	IsExpired       bool      `json:"is_expired"`
	Registrar       string    `json:"registrar"`
	ResponseTime    int64     `json:"response_time"`
	Success         bool      `json:"success"`
	Error           *string   `json:"error"`
	CheckedAt       time.Time `json:"checked_at"`
}

type JobResults struct {
	Summary            JobSummary                `json:"summary"`
	DNSBLResults       map[string]DNSBLResult    `json:"dnsbl_results,omitempty"`
	DNSResults         map[string][]DNSRecord    `json:"dns_results,omitempty"`
	EHLOResults        *EHLOResult               `json:"ehlo_results,omitempty"`
	HTTPResults        *HTTPResult               `json:"http_results,omitempty"`
	EmailAuthResults   *EmailAuthResult          `json:"email_auth_results,omitempty"`
	TCPResult          *TCPResult                `json:"tcp_result,omitempty"`
	UDPResult          *UDPResult                `json:"udp_result,omitempty"`
	SSLResult          *SSLResult                `json:"ssl_result,omitempty"`
	SSHResult          *SSHResult                `json:"ssh_result,omitempty"`
	PageLoadResult     *PageLoadResult           `json:"page_load_result,omitempty"`
	BrokenLinksResult  *BrokenLinksResult        `json:"broken_links_result,omitempty"`
	DomainExpiryResult *DomainExpiryResult       `json:"domain_expiry_result,omitempty"`
}

type Job struct {
	ID           string       `json:"id"`
	ScheduleID   string       `json:"schedule_id"`
	UserID       string       `json:"user_id"`
	Type         ScheduleType `json:"type"`
	Status       JobStatus    `json:"status"`
	StartedAt    *time.Time   `json:"started_at"`
	CompletedAt  *time.Time   `json:"completed_at"`
	Duration     *int64       `json:"duration"`
	ErrorMessage *string      `json:"error_message"`
	RegionID     *string      `json:"region_id"`
	CreatedAt    time.Time    `json:"created_at"`
	Results      JobResults   `json:"results"`
}

type AlertRule struct {
	ID               string    `json:"id"`
	UserID           string    `json:"user_id"`
	ScheduleID       string    `json:"schedule_id"`
	Name             string    `json:"name"`
	Metric           string    `json:"metric"`
	Condition        string    `json:"condition"`
	Threshold        float64   `json:"threshold"`
	ConsecutiveCount int       `json:"consecutive_count"`
	CooldownMinutes  int       `json:"cooldown_minutes"`
	IsEnabled        bool      `json:"is_enabled"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type MetricPoint struct {
	Timestamp          time.Time `json:"timestamp"`
	AvgLatencyMs       float64   `json:"avg_latency_ms"`
	MaxLatencyMs       float64   `json:"max_latency_ms"`
	MinLatencyMs       float64   `json:"min_latency_ms"`
	TotalRequests      int64     `json:"total_requests"`
	SuccessfulRequests int64     `json:"successful_requests"`
	UptimePercent      float64   `json:"uptime_percent"`
}

type SSLMetricPoint struct {
	MetricPoint
	AvgDaysUntilExpiry float64 `json:"avg_days_until_expiry"`
}

type EmailAuthMetricPoint struct {
	MetricPoint
	SPFValidCount   int64 `json:"spf_valid_count"`
	DMARCValidCount int64 `json:"dmarc_valid_count"`
	DKIMValidCount  int64 `json:"dkim_valid_count"`
}

type PageLoadMetricPoint struct {
	Timestamp          time.Time `json:"timestamp"`
	AvgResponseTimeMs  float64   `json:"avg_response_time_ms"`
	MaxResponseTimeMs  float64   `json:"max_response_time_ms"`
	MinResponseTimeMs  float64   `json:"min_response_time_ms"`
	AvgTTFBMs          float64   `json:"avg_ttfb_ms"`
	AvgDOMLoadMs       float64   `json:"avg_dom_load_ms"`
	AvgPageLoadMs      float64   `json:"avg_page_load_ms"`
	TotalRequests      int64     `json:"total_requests"`
	SuccessfulRequests int64     `json:"successful_requests"`
	UptimePercent      float64   `json:"uptime_percent"`
}

type DMARCMetricPoint struct {
	Timestamp             string `json:"timestamp"` // date string (YYYY-MM-DD)
	Domain                string `json:"domain"`
	TotalMessages         int64  `json:"total_messages"`
	DKIMPass              int64  `json:"dkim_pass"`
	DKIMFail              int64  `json:"dkim_fail"`
	SPFPass               int64  `json:"spf_pass"`
	SPFFail               int64  `json:"spf_fail"`
	DispositionNone       int64  `json:"disposition_none"`
	DispositionQuarantine int64  `json:"disposition_quarantine"`
	DispositionReject     int64  `json:"disposition_reject"`
	DMARCPass             int64  `json:"dmarc_pass"`
}

// DMARCReportList is returned by DMARCListReports.
type DMARCReportList struct {
	Data   []map[string]any `json:"data"`
	Count  int              `json:"count"`
	Limit  int              `json:"limit"`
	Offset int              `json:"offset"`
}

// DMARCReport is returned by DMARCGetReport.
type DMARCReport struct {
	Report  map[string]any   `json:"report"`
	Records []map[string]any `json:"records"`
}

// --- client ---

type Client struct {
	apiKey string
	base   string
	http   *http.Client
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

func getMetrics[T any](c *Client, scheduleID, kind string, from, to time.Time) ([]T, error) {
	var r struct {
		Data []T `json:"data"`
	}
	p := url.Values{}
	if !from.IsZero() {
		p.Set("from", from.UTC().Format(time.RFC3339))
	}
	if !to.IsZero() {
		p.Set("to", to.UTC().Format(time.RFC3339))
	}
	return r.Data, c.get("/v1/schedules/"+scheduleID+"/metrics/"+kind, p, &r)
}

// --- schedules ---

func (c *Client) ListSchedules() ([]Schedule, error) {
	var r struct{ Data []Schedule `json:"data"` }
	return r.Data, c.get("/v1/schedules", nil, &r)
}

func (c *Client) GetSchedule(id string) (Schedule, error) {
	var r struct{ Data Schedule `json:"data"` }
	return r.Data, c.get("/v1/schedules/"+id, nil, &r)
}

// --- jobs ---

func (c *Client) ListJobs(scheduleID string, limit, offset int) ([]Job, error) {
	var r struct{ Data []Job `json:"data"` }
	p := url.Values{"limit": {fmt.Sprint(limit)}, "offset": {fmt.Sprint(offset)}}
	return r.Data, c.get("/v1/schedules/"+scheduleID+"/jobs", p, &r)
}

func (c *Client) GetJob(scheduleID, jobID string) (Job, error) {
	var r struct{ Data Job `json:"data"` }
	return r.Data, c.get("/v1/schedules/"+scheduleID+"/jobs/"+jobID, nil, &r)
}

// --- alerts ---

func (c *Client) ListAlerts(scheduleID string) ([]AlertRule, error) {
	var r struct{ Data []AlertRule `json:"data"` }
	return r.Data, c.get("/v1/schedules/"+scheduleID+"/alerts", nil, &r)
}

// --- metrics ---

func (c *Client) MetricsHTTP(id string, from, to time.Time) ([]MetricPoint, error) {
	return getMetrics[MetricPoint](c, id, "http", from, to)
}
func (c *Client) MetricsDNS(id string, from, to time.Time) ([]MetricPoint, error) {
	return getMetrics[MetricPoint](c, id, "dns", from, to)
}
func (c *Client) MetricsSSL(id string, from, to time.Time) ([]SSLMetricPoint, error) {
	return getMetrics[SSLMetricPoint](c, id, "ssl", from, to)
}
func (c *Client) MetricsEmailAuth(id string, from, to time.Time) ([]EmailAuthMetricPoint, error) {
	return getMetrics[EmailAuthMetricPoint](c, id, "email_auth", from, to)
}
func (c *Client) MetricsPageLoad(id string, from, to time.Time) ([]PageLoadMetricPoint, error) {
	return getMetrics[PageLoadMetricPoint](c, id, "pageload", from, to)
}
func (c *Client) MetricsSSH(id string, from, to time.Time) ([]MetricPoint, error) {
	return getMetrics[MetricPoint](c, id, "ssh", from, to)
}

// --- dmarc ---

func (c *Client) DMARCListReports(limit, offset int) (DMARCReportList, error) {
	var r DMARCReportList
	p := url.Values{"limit": {fmt.Sprint(limit)}, "offset": {fmt.Sprint(offset)}}
	return r, c.get("/v1/dmarc/reports", p, &r)
}

func (c *Client) DMARCGetReport(id string) (DMARCReport, error) {
	var r DMARCReport
	return r, c.get("/v1/dmarc/reports/"+id, nil, &r)
}

func (c *Client) DMARCMetrics(domain string) ([]DMARCMetricPoint, error) {
	var r struct{ Data []DMARCMetricPoint `json:"data"` }
	p := url.Values{}
	if domain != "" {
		p.Set("domain", domain)
	}
	return r.Data, c.get("/v1/dmarc/metrics", p, &r)
}

func (c *Client) dmarcQuery(endpoint, domain string) ([]map[string]any, error) {
	var r struct{ Data []map[string]any `json:"data"` }
	p := url.Values{}
	if domain != "" {
		p.Set("domain", domain)
	}
	return r.Data, c.get("/v1/dmarc/"+endpoint, p, &r)
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
