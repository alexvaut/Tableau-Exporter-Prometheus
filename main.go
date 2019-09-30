package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	c "tableau/config"

	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
)

var (
	hitTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tableau_server_hits_total",
		Help: "The total number of hits on tableau server.",
	})

	sessionTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tableau_server_sessions_total",
		Help: "The total number of sessions on tableau server.",
	})

	distinctUserCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "tableau_server_users_count",
		Help: "The number of distinct users on the tableau server on a period of time.",
	}, []string{"period"})

	sessionDurationSeconds = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "tableau_session_duration_seconds",
		Help:    "The time to answer user request per project/workbook/view.",
		Buckets: []float64{10, 60, 300, 600, 900, 1800, 3600, 7200, 10800, 14400, 28800},
	})

	workbookHitTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tableau_hits_total",
		Help: "The total number of hits per project/workbook/view.",
	}, []string{"project", "workbook", "view"})

	workbookDistinctUserCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "tableau_users_count",
		Help: "The number of distinct users per project/workbook/view on a specific period of time.",
	}, []string{"period", "project", "workbook", "view"})

	viewLoadingTimeSeconds = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "tableau_response_time_seconds",
		Help:    "The time to answer user request per project/workbook/view.",
		Buckets: []float64{0.5, 1, 2, 3, 4, 5, 10, 15, 20, 30, 40, 50, 60},
	}, []string{"project", "workbook", "view"})

	m = map[string]float64{}
)

func WaitForCtrlC() {
	var end_waiter sync.WaitGroup
	end_waiter.Add(1)
	var signal_channel chan os.Signal
	signal_channel = make(chan os.Signal, 1)
	signal.Notify(signal_channel, os.Interrupt)
	go func() {
		<-signal_channel
		end_waiter.Done()
	}()
	end_waiter.Wait()
}

func startHttpServer(port int) *http.Server {
	srv := &http.Server{Addr: fmt.Sprintf(":%d", port)}

	go func() {
		// returns ErrServerClosed on graceful close
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			// NOTE: there is a chance that next line won't have time to run,
			// as main() doesn't wait for this goroutine to stop. don't use
			// code with race conditions like these for production. see post
			// comments below on more discussion on how to handle this.
			log.Fatalf("ListenAndServe(): %s", err)
		}
	}()

	// returning reference so caller can call Shutdown()
	return srv
}

func QueryOutInt(db *sql.DB, query string) int64 {
	var ret int64
	err := db.QueryRow(query).Scan(&ret)
	if err != nil {
		panic(err)
	}
	return ret
}

func QueryCountRow(db *sql.DB, query string) int64 {
	return QueryOutInt(db, `select count(*) FROM(`+query+`) as xdsdssff`)
}

func TotalVecQueryOffset(db *sql.DB, prometheusObject PromObj, query string) {
	keyOffset := query + "_offset"
	offset := int64(m[keyOffset])
	if offset == 0 {
		offset = QueryCountRow(db, query)
	}
	//only for debugging and get something:
	//offset = 0
	TotalVecQuery(db, prometheusObject, query+fmt.Sprintf(" OFFSET %d", offset))
}

func TotalVecQuery(db *sql.DB, prometheusObject PromObj, query string) {
	rows, err := db.Query(query)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("Zero rows found")
		} else {
			panic(err)
		}
	}

	columns, _ := rows.Columns()
	count := len(columns)
	values := make([]interface{}, count)
	valuePtrs := make([]interface{}, count)

	for rows.Next() {
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		rows.Scan(valuePtrs...)
		labels := make([]string, count-1)
		key := query

		for i := 0; i < len(columns)-1; i++ {
			val := values[i]
			b, _ := val.(string)
			//fmt.Printf("%d-%t-%s\n", i, ok, b)
			labels[i] = b
			key = key + labels[i] + "/"
		}

		value, ok := values[len(columns)-1].(int64)
		if ok {
			prometheusObject.Set(key, labels, float64(value))
		} else {
			value, ok := values[len(columns)-1].(float64)
			if ok {
				prometheusObject.Set(key, labels, value)
			}
		}
	}
}

func main() {

	fmt.Println("Starting...")
	var config c.Configurations = GetConfig()
	fmt.Printf("Configuration read. Scrap time = %d seconds.\n", config.ScrapeIntervalSeconds)

	http.Handle("/metrics", promhttp.Handler())
	srv := startHttpServer(config.Port)
	fmt.Printf("Metrics server started on 'http://localhost:%d/metrics'\n", config.Port)

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", config.Database.Host, config.Database.Port, config.Database.User, config.Database.Password, config.Database.Name)
	fmt.Println("Tableau DB connection: " + fmt.Sprintf("host=%s port=%d user=%s password=*** dbname=%s sslmode=disable", config.Database.Host, config.Database.Port, config.Database.User, config.Database.Name))
	var firstConnection bool = true

	go func() {
		for {

			db, err := sql.Open("postgres", psqlInfo)
			if err != nil {
				panic(err)
			}
			defer db.Close()

			err = db.Ping()
			if err != nil {
				panic(err)
			}
			defer db.Close()

			if firstConnection {
				fmt.Print("Tableau DB connected.")
				firstConnection = false
			}

			var periods = [...]string{"month", "week", "day", "hour"}
			for _, period := range periods {
				TotalVecQuery(db, GaugeObj{distinctUserCount.WithLabelValues(period)}, fmt.Sprintf("select count(distinct hist_actor_user_id) from historical_events where created_at >= date_trunc('%s', current_timestamp)", period))

				TotalVecQuery(db, GaugeVecObj{workbookDistinctUserCount.MustCurryWith(prometheus.Labels{"period": period})},
					fmt.Sprintf(`select hist_projects.name as projectName, hist_workbooks.name as workbookName, hist_views.name as viewName, count(distinct hist_actor_user_id) as count from historical_events
						LEFT JOIN hist_projects ON historical_events.hist_project_id = hist_projects.id
						LEFT JOIN hist_workbooks ON historical_events.hist_workbook_id = hist_workbooks.id
						LEFT JOIN hist_views ON historical_events.hist_view_id = hist_views.id				
						where created_at >= date_trunc('%s', current_timestamp)
						group by hist_workbooks.name, hist_views.name, hist_projects.name`, period))
			}

			TotalVecQuery(db, CounterObj{hitTotal}, `select count(*) from historical_events`)
			TotalVecQuery(db, CounterObj{sessionTotal}, `select count(http_requests.vizql_session) FROM http_requests where http_requests.vizql_session notnull`)

			TotalVecQuery(db, CounterVecObj{workbookHitTotal},
				`select hist_projects.name as projectName, hist_workbooks.name as workbookName, hist_views.name as viewName, count(*) as count from historical_events
				LEFT JOIN hist_projects ON historical_events.hist_project_id = hist_projects.id
				LEFT JOIN hist_workbooks ON historical_events.hist_workbook_id = hist_workbooks.id
				LEFT JOIN hist_views ON historical_events.hist_view_id = hist_views.id				
				group by hist_workbooks.name, hist_views.name, hist_projects.name`)

			TotalVecQueryOffset(db, HistogramVecObj{viewLoadingTimeSeconds}, ` select
			dataset._workbooks_project_name as project,
			dataset._workbooks_name as workbook,
			dataset._views_name as view,
			EXTRACT(EPOCH from (dataset._http_requests_completed_at - _http_requests_created_at)) as durationSeconds`+
				httpQuery+
				`where (dataset._http_requests_action = 'bootstrapSession' or dataset._http_requests_action = 'show') 
			and dataset._workbooks_project_name notnull
			order by _http_requests_created_at `)

			TotalVecQueryOffset(db, HistogramObj{sessionDurationSeconds}, ` select
			table4.durationSeconds
			from
			(			
			SELECT 
				max(http_requests.completed_at) AS end_at,
				EXTRACT(EPOCH from ( max(http_requests.completed_at) - min(http_requests.created_at))) as durationSeconds,
				(SELECT max(http_requests.completed_at) from http_requests) as maxTime
			FROM http_requests
			where http_requests.vizql_session notnull
			GROUP BY  http_requests.vizql_session
			) as table4
			where (maxTime - end_at) >= INTERVAL '3600' second `)

			db.Close()
			time.Sleep(time.Duration(config.ScrapeIntervalSeconds) * time.Second)
		}
	}()

	WaitForCtrlC()
	fmt.Println("Exiting...")
	if err := srv.Shutdown(context.TODO()); err != nil {
		panic(err) // failure/timeout shutting down the server gracefully
	}

	fmt.Println("Exit.")
}

func GetConfig() c.Configurations {
	// Set the file name of the configurations file
	viper.SetConfigName("config")

	// Set the path to look for the configurations file
	viper.AddConfigPath(".")

	// Enable VIPER to read Environment Variables
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetConfigType("yml")

	var configuration c.Configurations
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Error reading config file, %s", err)
	}

	err := viper.Unmarshal(&configuration)
	if err != nil {
		fmt.Printf("Unable to decode into struct, %v", err)
	}

	return configuration
}
