# seabhac Go client

A Go client for the [Seabhac](https://seabhac.io) monitoring API.

## Installation

```bash
go get github.com/seabhac-io/go-client
```

## Usage

```go
import seabhac "github.com/seabhac-io/go-client"

client := seabhac.New("your-api-key")
```

To target a self-hosted or staging instance:

```go
client := seabhac.New("your-api-key").WithBaseURL("https://api.example.com")
```

## Schedules

```go
// List all schedules
schedules, err := client.ListSchedules()

// Get a single schedule
schedule, err := client.GetSchedule("schedule-id")
```

## Jobs

```go
// List jobs for a schedule (paginated)
jobs, err := client.ListJobs("schedule-id", 20, 0)

// Get a single job with full results
job, err := client.GetJob("schedule-id", "job-id")

// Access type-specific results
if job.Results.HTTPResults != nil {
    fmt.Println(job.Results.HTTPResults.StatusCode)
}
if job.Results.SSLResult != nil {
    fmt.Println(job.Results.SSLResult.Expiry)
}
```

## Alerts

```go
alerts, err := client.ListAlerts("schedule-id")
```

## Metrics

Each method accepts an optional time range (`from`, `to`). Pass zero `time.Time` values to use the API default.

```go
now := time.Now()
week := now.Add(-7 * 24 * time.Hour)

points, err := client.MetricsHTTP("schedule-id", week, now)
points, err := client.MetricsDNS("schedule-id", week, now)
points, err := client.MetricsSSL("schedule-id", week, now)      // []SSLMetricPoint
points, err := client.MetricsEmailAuth("schedule-id", week, now) // []EmailAuthMetricPoint
points, err := client.MetricsPageLoad("schedule-id", week, now)  // []PageLoadMetricPoint
points, err := client.MetricsSSH("schedule-id", week, now)
```

## DMARC

```go
// List DMARC aggregate reports (paginated)
list, err := client.DMARCListReports(20, 0)

// Get a single report with records
report, err := client.DMARCGetReport("report-id")

// Metrics (optionally filter by domain)
metrics, err := client.DMARCMetrics("example.com")

// Analytics
geo, err      := client.DMARCGeo("example.com")
topIPs, err   := client.DMARCTopIPs("example.com")
failing, err  := client.DMARCTopFailingIPs("example.com")
senders, err  := client.DMARCTopSenders("example.com")
reasons, err  := client.DMARCFailReasons("example.com")
reporters,err := client.DMARCReporters("example.com")
asns, err     := client.DMARCTopASNs("example.com")
```

## Schedule types

| Constant | Value |
|---|---|
| `ScheduleTypeDNSBL` | `dnsbl` |
| `ScheduleTypeEHLO` | `ehlo` |
| `ScheduleTypeDNS` | `dns` |
| `ScheduleTypeHTTP` | `http` |
| `ScheduleTypeEmailAuth` | `email_auth` |
| `ScheduleTypeTCP` | `tcp` |
| `ScheduleTypeUDP` | `udp` |
| `ScheduleTypePageLoad` | `pageload` |
| `ScheduleTypeSSL` | `ssl` |
| `ScheduleTypeSSH` | `ssh` |
| `ScheduleTypeBrokenLinks` | `broken_links` |
| `ScheduleTypeDomainExpiry` | `domain_expiry` |
| `ScheduleTypeDMARC` | `dmarc` |

## License

See [LICENSE](LICENSE).
