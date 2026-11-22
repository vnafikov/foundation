package appinfo

import (
	"context"

	"github.com/vnafikov/common/text"

	"github.com/vnafikov/foundation/logger"
)

const (
	templateName = "Application info"
	tpl          = `{{if .IsDev.IsValid}}The application is running in {{if .IsDev.Bool}}development{{else}}production{{end}}.
{{end}}{{if .HasExternalLog.IsValid}}{{if .HasExternalLog.Bool}}External logging is enabled{{else}}Logging to standard output only{{end}}` +
		`{{else}}Logging{{end}} at the {{.LogLevel}} level.
{{if .DisableScheduler.Bool}}The scheduler is disabled.
{{end}}Version: {{if .Version.IsValid}}{{.Version.String}}{{else}}0{{end}}-{{.GoVersion}}
{{if .HTTPPort.IsValid}}HTTP server port: {{.HTTPPort.String}}
{{end}}{{if .GRPCPort.IsValid}}gRPC server port: {{.GRPCPort.String}}
{{end}}{{if .WorkerID.IsValid}}Worker ID: {{.WorkerID.Int}}
{{end}}{{if and (not .DisableScheduler.Bool) .ScheduledTasks.IsValid}}Scheduled tasks: {{.ScheduledTasks.Int}}
{{end}}{{if .BaseDomain.IsValid}}Base domain: {{.BaseDomain.String}}
{{end}}Time zone: {{.TimeZoneName}} {{.TimeZoneOffset}}
Logical CPUs: {{.CPUs}}
Process ID: {{.PID}}
Parent process ID: {{.PPID}}
{{if .Extras}}{{range $key, $value := .Extras}}{{$key}}{{if $value}}: {{$value}}{{end}}
{{end}}{{end}}`
)

func Print(ctx context.Context, options Options) error {
	data := newTemplateData(options)
	msg, err := text.Template(templateName, tpl, data)
	if err != nil {
		logger.Error(ctx, err, "text.Template")

		return err
	}

	logger.Important(ctx, msg)

	return nil
}
