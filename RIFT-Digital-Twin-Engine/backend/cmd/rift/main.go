// Command rift is the RIFT CLI: a thin HTTP client over riftd's REST API
// implementing twin, entity, sensor, ingest, simulate, scenario, snapshot,
// replay, analyze, optimize, benchmark, export and worker subcommands.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var (
	serverURL string
	authToken string
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	serverURL = envOr("RIFT_SERVER", "http://localhost:8080")
	authToken = os.Getenv("RIFT_TOKEN")

	cmd := os.Args[1]
	args := os.Args[2:]
	var err error
	switch cmd {
	case "login":
		err = cmdLogin(args)
	case "twin":
		err = cmdTwin(args)
	case "entity":
		err = cmdEntity(args)
	case "sensor":
		err = cmdSensor(args)
	case "ingest":
		err = cmdIngest(args)
	case "simulate":
		err = cmdSimulate(args)
	case "scenario":
		err = cmdScenario(args)
	case "snapshot", "export":
		err = cmdSnapshot(args)
	case "replay":
		err = cmdReplay(args)
	case "analyze":
		err = cmdAnalyze(args)
	case "optimize":
		err = cmdOptimize(args)
	case "benchmark":
		err = cmdBenchmark(args)
	case "worker":
		err = cmdWorker(args)
	case "help", "-h", "--help":
		printUsage()
		return
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`RIFT CLI — Digital Twin Engine

Usage: rift <command> [subcommand] [flags]

Commands:
  login                    --username --password           obtain a session token (set RIFT_TOKEN to use it)
  twin list
  twin create              --name --template
  twin show                <twinId>
  entity list               <twinId>
  entity create             <twinId> --name --type
  sensor list                <twinId>
  sensor create               <twinId> --entity --type --unit --min --max
  ingest                     <sensorId> --value
  simulate clock              <twinId> [pause|resume|speed|step] [--speed] [--seconds]
  scenario create              <twinId> --name --ticks --fault type:entityId:atTick:magnitude (repeatable)
  scenario run                  <scenarioId>
  snapshot export | export       <twinId> --format json|csv --out file
  snapshot import                <twinId> --file file.json
  replay events                   <twinId> --limit 100
  analyze impact                   <entityId>
  analyze rootcause                 <entityId> --window 30
  optimize                          <twinId> --param name --min --max --fault type:entityId:0:0 --trials 60
  benchmark                          <twinId> --count 100000 --seed 42
  worker

Environment:
  RIFT_SERVER   base URL of a running riftd (default http://localhost:8080)
  RIFT_TOKEN    bearer token from 'rift login', required for admin-only endpoints`)
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

// ---------- HTTP helpers ----------

func apiRequest(method, path string, body interface{}) ([]byte, error) {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(data)
	}
	req, err := http.NewRequest(method, serverURL+path, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not reach %s (is riftd running?): %w", serverURL, err)
	}
	defer resp.Body.Close()
	out, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return out, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(out))
	}
	return out, nil
}

func printPretty(data []byte) {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		fmt.Println(string(data))
		return
	}
	pretty, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(pretty))
}

// ---------- login ----------

func cmdLogin(args []string) error {
	fs := flag.NewFlagSet("login", flag.ExitOnError)
	username := fs.String("username", "admin", "username")
	password := fs.String("password", "", "password")
	fs.Parse(args)
	data, err := apiRequest("POST", "/api/auth/login", map[string]string{"username": *username, "password": *password})
	if err != nil {
		return err
	}
	printPretty(data)
	fmt.Println("\nExport this token: export RIFT_TOKEN=<token>")
	return nil
}

// ---------- twin ----------

func cmdTwin(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: rift twin [list|create|show] ...")
	}
	switch args[0] {
	case "list":
		data, err := apiRequest("GET", "/api/twins", nil)
		if err != nil {
			return err
		}
		printPretty(data)
	case "create":
		fs := flag.NewFlagSet("twin create", flag.ExitOnError)
		name := fs.String("name", "", "twin name")
		template := fs.String("template", "", "factory|building|data_center|warehouse|smart_city")
		fs.Parse(args[1:])
		if *name == "" {
			return fmt.Errorf("--name is required")
		}
		data, err := apiRequest("POST", "/api/twins", map[string]string{"name": *name, "template": *template})
		if err != nil {
			return err
		}
		var twin struct{ ID string `json:"id"` }
		json.Unmarshal(data, &twin)
		if *template != "" && twin.ID != "" {
			genData, genErr := apiRequest("POST", "/api/twins/"+twin.ID+"/synthetic/"+*template, nil)
			if genErr == nil {
				fmt.Println("template generated:")
				printPretty(genData)
			}
		}
		printPretty(data)
	case "show":
		if len(args) < 2 {
			return fmt.Errorf("usage: rift twin show <twinId>")
		}
		data, err := apiRequest("GET", "/api/twins/"+args[1], nil)
		if err != nil {
			return err
		}
		printPretty(data)
	default:
		return fmt.Errorf("unknown twin subcommand: %s", args[0])
	}
	return nil
}

// ---------- entity ----------

func cmdEntity(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: rift entity [list|create] <twinId> ...")
	}
	sub, twinID := args[0], args[1]
	switch sub {
	case "list":
		data, err := apiRequest("GET", "/api/twins/"+twinID+"/entities", nil)
		if err != nil {
			return err
		}
		printPretty(data)
	case "create":
		fs := flag.NewFlagSet("entity create", flag.ExitOnError)
		name := fs.String("name", "", "entity name")
		typ := fs.String("type", "custom", "entity type")
		parent := fs.String("parent", "", "parent entity id")
		fs.Parse(args[2:])
		if *name == "" {
			return fmt.Errorf("--name is required")
		}
		data, err := apiRequest("POST", "/api/twins/"+twinID+"/entities", map[string]string{"name": *name, "type": *typ, "parentId": *parent})
		if err != nil {
			return err
		}
		printPretty(data)
	default:
		return fmt.Errorf("unknown entity subcommand: %s", sub)
	}
	return nil
}

// ---------- sensor ----------

func cmdSensor(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: rift sensor [list|create] <twinId> ...")
	}
	sub, twinID := args[0], args[1]
	switch sub {
	case "list":
		data, err := apiRequest("GET", "/api/twins/"+twinID+"/sensors", nil)
		if err != nil {
			return err
		}
		printPretty(data)
	case "create":
		fs := flag.NewFlagSet("sensor create", flag.ExitOnError)
		entityID := fs.String("entity", "", "owning entity id")
		typ := fs.String("type", "custom", "sensor type")
		unit := fs.String("unit", "", "unit")
		min := fs.Float64("min", 0, "range min")
		max := fs.Float64("max", 100, "range max")
		fs.Parse(args[2:])
		if *entityID == "" {
			return fmt.Errorf("--entity is required")
		}
		data, err := apiRequest("POST", "/api/twins/"+twinID+"/sensors", map[string]interface{}{
			"entityId": *entityID, "type": *typ, "unit": *unit, "rangeMin": *min, "rangeMax": *max, "source": "http",
		})
		if err != nil {
			return err
		}
		printPretty(data)
	default:
		return fmt.Errorf("unknown sensor subcommand: %s", sub)
	}
	return nil
}

// ---------- ingest ----------

func cmdIngest(args []string) error {
	fs := flag.NewFlagSet("ingest", flag.ExitOnError)
	value := fs.Float64("value", 0, "reading value")
	if len(args) < 1 {
		return fmt.Errorf("usage: rift ingest <sensorId> --value=X")
	}
	sensorID := args[0]
	fs.Parse(args[1:])
	data, err := apiRequest("POST", "/api/telemetry", map[string]interface{}{"sensorId": sensorID, "value": *value})
	if err != nil {
		return err
	}
	printPretty(data)
	return nil
}

// ---------- simulate ----------

func cmdSimulate(args []string) error {
	if len(args) < 2 || args[0] != "clock" {
		return fmt.Errorf("usage: rift simulate clock <twinId> [pause|resume|speed|step] [flags]")
	}
	twinID := args[1]
	if len(args) == 2 {
		data, err := apiRequest("GET", "/api/twins/"+twinID+"/clock", nil)
		if err != nil {
			return err
		}
		printPretty(data)
		return nil
	}
	action := args[2]
	fs := flag.NewFlagSet("clock", flag.ExitOnError)
	speed := fs.Float64("speed", 1, "speed multiplier")
	seconds := fs.Int("seconds", 60, "seconds to step")
	fs.Parse(args[3:])
	data, err := apiRequest("POST", "/api/twins/"+twinID+"/clock/"+action, map[string]interface{}{"speed": *speed, "seconds": *seconds})
	if err != nil {
		return err
	}
	printPretty(data)
	return nil
}

// ---------- scenario ----------

func parseFault(spec string) (map[string]interface{}, error) {
	parts := strings.Split(spec, ":")
	if len(parts) != 4 {
		return nil, fmt.Errorf("fault spec must be type:entityId:atTick:magnitude, got %q", spec)
	}
	atTick, _ := strconv.Atoi(parts[2])
	magnitude, _ := strconv.ParseFloat(parts[3], 64)
	return map[string]interface{}{"type": parts[0], "entityId": parts[1], "atTick": atTick, "magnitude": magnitude}, nil
}

type stringList []string

func (s *stringList) String() string { return strings.Join(*s, ",") }
func (s *stringList) Set(v string) error {
	*s = append(*s, v)
	return nil
}

func cmdScenario(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: rift scenario [create <twinId>|run <scenarioId>] ...")
	}
	switch args[0] {
	case "create":
		twinID := args[1]
		fs := flag.NewFlagSet("scenario create", flag.ExitOnError)
		name := fs.String("name", "What-If Scenario", "scenario name")
		ticks := fs.Int("ticks", 50, "duration in ticks")
		var faults stringList
		fs.Var(&faults, "fault", "type:entityId:atTick:magnitude (repeatable)")
		fs.Parse(args[2:])
		var faultSpecs []map[string]interface{}
		for _, f := range faults {
			spec, err := parseFault(f)
			if err != nil {
				return err
			}
			faultSpecs = append(faultSpecs, spec)
		}
		data, err := apiRequest("POST", "/api/twins/"+twinID+"/scenarios", map[string]interface{}{
			"name": *name, "durationTicks": *ticks, "faults": faultSpecs,
		})
		if err != nil {
			return err
		}
		printPretty(data)
	case "run":
		scenarioID := args[1]
		data, err := apiRequest("POST", "/api/scenarios/"+scenarioID+"/run", nil)
		if err != nil {
			return err
		}
		printPretty(data)
	default:
		return fmt.Errorf("unknown scenario subcommand: %s", args[0])
	}
	return nil
}

// ---------- snapshot / export ----------

func cmdSnapshot(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: rift snapshot [export|import] <twinId> ...")
	}
	sub, twinID := args[0], args[1]
	switch sub {
	case "export":
		fs := flag.NewFlagSet("export", flag.ExitOnError)
		format := fs.String("format", "json", "json|csv")
		out := fs.String("out", "", "output file (default stdout)")
		fs.Parse(args[2:])
		path := "/api/twins/" + twinID + "/export." + *format
		data, err := apiRequest("GET", path, nil)
		if err != nil {
			return err
		}
		if *out == "" {
			fmt.Println(string(data))
			return nil
		}
		return os.WriteFile(*out, data, 0o644)
	case "import":
		fs := flag.NewFlagSet("import", flag.ExitOnError)
		file := fs.String("file", "", "JSON file with {\"rows\": [...]}")
		fs.Parse(args[2:])
		if *file == "" {
			return fmt.Errorf("--file is required")
		}
		content, err := os.ReadFile(*file)
		if err != nil {
			return err
		}
		var body map[string]interface{}
		if err := json.Unmarshal(content, &body); err != nil {
			return err
		}
		data, err := apiRequest("POST", "/api/twins/"+twinID+"/import", body)
		if err != nil {
			return err
		}
		printPretty(data)
	default:
		return fmt.Errorf("unknown snapshot subcommand: %s", sub)
	}
	return nil
}

// ---------- replay ----------

func cmdReplay(args []string) error {
	if len(args) < 2 || args[0] != "events" {
		return fmt.Errorf("usage: rift replay events <twinId> [--limit N]")
	}
	twinID := args[1]
	fs := flag.NewFlagSet("replay", flag.ExitOnError)
	limit := fs.Int("limit", 100, "max events")
	fs.Parse(args[2:])
	data, err := apiRequest("GET", fmt.Sprintf("/api/twins/%s/events?limit=%d", twinID, *limit), nil)
	if err != nil {
		return err
	}
	printPretty(data)
	return nil
}

// ---------- analyze ----------

func cmdAnalyze(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: rift analyze [impact|rootcause] <entityId> ...")
	}
	sub, entityID := args[0], args[1]
	switch sub {
	case "impact":
		data, err := apiRequest("GET", "/api/entities/"+entityID+"/impact", nil)
		if err != nil {
			return err
		}
		printPretty(data)
	case "rootcause":
		fs := flag.NewFlagSet("rootcause", flag.ExitOnError)
		window := fs.Int("window", 30, "window in minutes")
		fs.Parse(args[2:])
		data, err := apiRequest("GET", fmt.Sprintf("/api/entities/%s/rootcause?windowMinutes=%d", entityID, *window), nil)
		if err != nil {
			return err
		}
		printPretty(data)
	default:
		return fmt.Errorf("unknown analyze subcommand: %s", sub)
	}
	return nil
}

// ---------- optimize ----------

func cmdOptimize(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: rift optimize <twinId> --param name --min 0 --max 1 --fault type:entityId:0:0")
	}
	twinID := args[0]
	fs := flag.NewFlagSet("optimize", flag.ExitOnError)
	param := fs.String("param", "allocation", "parameter name")
	min := fs.Float64("min", 0, "lower bound")
	max := fs.Float64("max", 1, "upper bound")
	trials := fs.Int("trials", 60, "search trials")
	var faults stringList
	fs.Var(&faults, "fault", "type:entityId:atTick:magnitude (repeatable)")
	fs.Parse(args[1:])
	var faultSpecs []map[string]interface{}
	for _, f := range faults {
		spec, err := parseFault(f)
		if err != nil {
			return err
		}
		faultSpecs = append(faultSpecs, spec)
	}
	data, err := apiRequest("POST", "/api/twins/"+twinID+"/optimize", map[string]interface{}{
		"parameterName": *param, "min": *min, "max": *max, "trials": *trials, "faults": faultSpecs,
	})
	if err != nil {
		return err
	}
	printPretty(data)
	return nil
}

// ---------- benchmark ----------

func cmdBenchmark(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: rift benchmark <twinId> --count 100000 --seed 42")
	}
	twinID := args[0]
	fs := flag.NewFlagSet("benchmark", flag.ExitOnError)
	count := fs.Int("count", 10000, "entities to generate")
	seed := fs.Int64("seed", 42, "RNG seed")
	fs.Parse(args[1:])
	data, err := apiRequest("POST", fmt.Sprintf("/api/twins/%s/synthetic/benchmark?count=%d&seed=%d", twinID, *count, *seed), nil)
	if err != nil {
		return err
	}
	printPretty(data)
	return nil
}

// ---------- worker ----------

func cmdWorker(args []string) error {
	fmt.Println(`This build of RIFT runs telemetry ingestion, rules, anomaly detection and
simulation inside the single riftd process using a Go worker-pool per
subsystem (see internal/telemetry.Pipeline and internal/simulation.Engine).

To scale horizontally, run multiple riftd instances behind a shared
PostgreSQL/TimescaleDB + Kafka/NATS backend (swap internal/registry and
internal/telemetry's storage for the distributed backends described in the
README's Connector Framework section) and partition twins across instances;
each instance is already a self-contained "worker" in that topology.`)
	return nil
}
