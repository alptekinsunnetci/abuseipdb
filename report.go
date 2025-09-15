package main

import (
	"html/template"
	"os"
	"time"
)

type ReportRow struct {
	IPAddress            string
	CountryName          string
	NumReports           int
	AbuseConfidenceScore int
	LastReportedAt       string
}

type ReportData struct {
	Rows      []ReportRow
	Timestamp string
}

const reportTemplate = `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>AbuseIPDB Report</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.3/dist/css/bootstrap.min.css" rel="stylesheet">
  </head>
  <body>
    <div class="container py-4">
      <h1 class="mb-4">AbuseIPDB Son 7 Gün Raporu</h1>
      <p class="text-muted mb-3">Rapor Oluşturulma Tarihi: {{.Timestamp}}</p>
      <div class="table-responsive">
        <table class="table table-striped table-hover align-middle">
          <thead class="table-dark">
            <tr>
              <th>IP Address</th>
              <th>Country Code</th>
              <th>Num Reports</th>
              <th>Abuse Confidence Score</th>
              <th>Last Reported At</th>
            </tr>
          </thead>
          <tbody>
          {{range .Rows}}
            <tr>
              <td><code>{{.IPAddress}}</code></td>
              <td>{{.CountryName}}</td>
              <td>{{.NumReports}}</td>
              <td><span class="badge bg-{{scoreClass .AbuseConfidenceScore}}">{{.AbuseConfidenceScore}}</span></td>
              <td>{{.LastReportedAt}}</td>
            </tr>
          {{else}}
            <tr><td colspan="5" class="text-center text-muted">Son 7 günde raporlanan IP bulunamadı.</td></tr>
          {{end}}
          </tbody>
        </table>
      </div>
    </div>
  </body>
</html>`

func scoreClass(score int) string {
	switch {
	case score >= 75:
		return "danger"
	case score >= 50:
		return "warning"
	case score > 0:
		return "info"
	default:
		return "secondary"
	}
}

func RenderReport(rows []ReportRow, outPath string) error {
	funcMap := template.FuncMap{
		"scoreClass": scoreClass,
	}
	tpl, err := template.New("report").Funcs(funcMap).Parse(reportTemplate)
	if err != nil {
		return err
	}
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()

	reportData := ReportData{
		Rows:      rows,
		Timestamp: time.Now().Format("02.01.2006 15:04:05"),
	}

	return tpl.Execute(f, reportData)
}
