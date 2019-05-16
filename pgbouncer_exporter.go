package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
)

const (
	namespace = "pgbouncer"
	indexHTML = `
	<html>
		<head>
			<title>PgBouncer Exporter</title>
		</head>
		<body>
			<h1>PgBouncer Exporter</h1>
			<p>
			<a href='%s'>Metrics</a>
			</p>
		</body>
	</html>`
)

func main() {
	var (
		showVersion             = flag.Bool("version", false, "Print version information.")
		listenAddress           = flag.String("web.listen-address", ":9127", "Address on which to expose metrics and web interface.")
		connectionStringPointer = flag.String("pgBouncer.connectionString", "postgres://postgres:@localhost:6543/pgbouncer?sslmode=disable", "Connection string for accessing pgBouncer.")
		metricsPath             = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	)

	flag.Parse()

	if *showVersion {
		fmt.Fprintln(os.Stdout, version.Print("pgbouncer_exporter"))
		os.Exit(0)
	}

	userCfg, err := pgx.ParseConnectionString(*connectionStringPointer)
	if err != nil {
		panic(err)
	}

	envCfg, err := pgx.ParseEnvLibpq()
	if err != nil {
		panic(err)
	}

	cfg := pgx.ConnConfig{
		RuntimeParams: map[string]string{"client_encoding": "UTF8"},
		// We need to use SimpleProtocol in order to communicate with PgBouncer
		PreferSimpleProtocol: true,
		CustomConnInfo: func(_ *pgx.Conn) (*pgtype.ConnInfo, error) {
			connInfo := pgtype.NewConnInfo()
			connInfo.InitializeDataTypes(map[string]pgtype.OID{
				"int4":    pgtype.Int4OID,
				"name":    pgtype.NameOID,
				"oid":     pgtype.OIDOID,
				"text":    pgtype.TextOID,
				"varchar": pgtype.VarcharOID,
			})

			return connInfo, nil
		},
	}

	cfg = cfg.Merge(envCfg)
	cfg = cfg.Merge(userCfg)

	exporter := NewExporter(cfg, namespace)
	prometheus.MustRegister(exporter)

	log.Infoln("Starting pgbouncer exporter version: ", version.Info())

	http.Handle(*metricsPath, prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(fmt.Sprintf(indexHTML, *metricsPath)))
	})

	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
