package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/rdruid-talent-analyzer/go-backend/internal/analysis"
	"github.com/rdruid-talent-analyzer/go-backend/internal/models"
	"github.com/rdruid-talent-analyzer/go-backend/internal/output"
	"github.com/rdruid-talent-analyzer/go-backend/internal/wcl"
	"github.com/rdruid-talent-analyzer/go-backend/internal/web"
)

func main() {
	_ = godotenv.Load()
	_ = godotenv.Load("../.env")

	rootCmd := &cobra.Command{
		Use:   "flourish",
		Short: "Resto Druid talent analyzer",
	}

	// --- analyze command ---
	var fightID int
	var playerName string

	analyzeCmd := &cobra.Command{
		Use:   "analyze <report_code>",
		Short: "Analyze a WarcraftLogs report",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			runAnalyze(args[0], fightID, playerName)
		},
	}
	analyzeCmd.Flags().IntVar(&fightID, "fight", 0, "Fight ID (interactive if omitted)")
	analyzeCmd.Flags().StringVar(&playerName, "player", "", "Player name (interactive if omitted)")
	rootCmd.AddCommand(analyzeCmd)

	// Allow `flourish <code>` as shorthand for `flourish analyze <code>`
	rootCmd.Args = cobra.ArbitraryArgs
	rootCmd.Run = func(cmd *cobra.Command, args []string) {
		if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
			runAnalyze(args[0], fightID, playerName)
			return
		}
		cmd.Help()
	}
	rootCmd.Flags().IntVar(&fightID, "fight", 0, "Fight ID")
	rootCmd.Flags().StringVar(&playerName, "player", "", "Player name")

	// --- serve command ---
	var port string

	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the web server",
		Run: func(cmd *cobra.Command, args []string) {
			runServe(port)
		},
	}
	serveCmd.Flags().StringVar(&port, "port", "", "Port (default: $PORT or 8000)")
	rootCmd.AddCommand(serveCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func newClient() wcl.Querier {
	clientID := os.Getenv("WCL_CLIENT_ID")
	clientSecret := os.Getenv("WCL_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		fmt.Println("Error: WCL_CLIENT_ID and WCL_CLIENT_SECRET must be set")
		os.Exit(1)
	}
	inner := wcl.NewClient(clientID, clientSecret)
	return wcl.NewCachedClient(inner, "data/cache")
}

func runAnalyze(reportCode string, fightID int, playerName string) {
	client := newClient()

	report, err := client.GetReport(reportCode)
	if err != nil {
		fmt.Printf("Error fetching report: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Report: %s\n", report["title"])

	rawFights, _ := report["fights"].([]any)
	masterData, _ := report["masterData"].(map[string]any)
	rawActors, _ := masterData["actors"].([]any)

	type fightInfo struct {
		id        int
		name      string
		kill      bool
		startTime float64
		endTime   float64
	}
	var fights []fightInfo
	for _, f := range rawFights {
		fm := f.(map[string]any)
		if toInt(fm["encounterID"]) > 0 {
			fights = append(fights, fightInfo{
				id:        toInt(fm["id"]),
				name:      fmt.Sprint(fm["name"]),
				kill:      toBool(fm["kill"]),
				startTime: toFloat(fm["startTime"]),
				endTime:   toFloat(fm["endTime"]),
			})
		}
	}

	if fightID == 0 {
		fmt.Println("\nFights:")
		for _, f := range fights {
			status := "Kill"
			if !f.kill {
				status = "Wipe"
			}
			fmt.Printf("  %3d: %s (%s, %.0fs)\n", f.id, f.name, status, (f.endTime-f.startTime)/1000)
		}
		fmt.Print("Select fight ID: ")
		reader := bufio.NewReader(os.Stdin)
		line, _ := reader.ReadString('\n')
		fightID, _ = strconv.Atoi(strings.TrimSpace(line))
	}

	var selectedFight *fightInfo
	for i := range fights {
		if fights[i].id == fightID {
			selectedFight = &fights[i]
			break
		}
	}
	if selectedFight == nil {
		fmt.Println("Fight not found")
		os.Exit(1)
	}

	type actorInfo struct {
		id, petOwner         int
		name, subType, server string
	}
	var allActors []actorInfo
	for _, a := range rawActors {
		am := a.(map[string]any)
		allActors = append(allActors, actorInfo{
			id:       toInt(am["id"]),
			name:     fmt.Sprint(am["name"]),
			subType:  fmt.Sprint(am["subType"]),
			server:   fmt.Sprint(am["server"]),
			petOwner: toInt(am["petOwner"]),
		})
	}

	var druids []actorInfo
	for _, a := range allActors {
		if a.subType == "Druid" {
			druids = append(druids, a)
		}
	}

	var selectedPlayer *actorInfo
	if playerName == "" {
		if len(druids) == 1 {
			selectedPlayer = &druids[0]
			fmt.Printf("Auto-selected: %s\n", selectedPlayer.name)
		} else {
			fmt.Println("\nResto Druids:")
			for _, d := range druids {
				fmt.Printf("  %3d: %s (%s)\n", d.id, d.name, d.server)
			}
			fmt.Print("Select player ID: ")
			reader := bufio.NewReader(os.Stdin)
			line, _ := reader.ReadString('\n')
			pid, _ := strconv.Atoi(strings.TrimSpace(line))
			for i := range druids {
				if druids[i].id == pid {
					selectedPlayer = &druids[i]
					break
				}
			}
		}
	} else {
		for i := range druids {
			if strings.EqualFold(druids[i].name, playerName) {
				selectedPlayer = &druids[i]
				break
			}
		}
	}
	if selectedPlayer == nil {
		fmt.Println("Player not found")
		os.Exit(1)
	}

	fmt.Printf("\nFetching events for %s in %s...\n", selectedPlayer.name, selectedFight.name)
	events, err := client.GetEvents(reportCode, selectedFight.id, selectedPlayer.id, selectedFight.startTime, selectedFight.endTime)
	if err != nil {
		fmt.Printf("Error fetching events: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Fetched %d events\n", len(events))

	regrowthFilter := `IN RANGE FROM (type = "applybuff" OR type = "refreshbuff") AND ability.id = 8936 TO type = "removebuff" AND ability.id = 8936 GROUP BY target ON target END`
	damageTaken, _ := client.GetDamageTaken(reportCode, selectedFight.id, selectedPlayer.id, selectedFight.startTime, selectedFight.endTime, regrowthFilter)

	configPath := "config/talents.yaml"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		configPath = "../config/talents.yaml"
	}
	config, err := models.LoadConfig(configPath)
	if err != nil {
		config = &models.Config{
			Mastery: models.MasteryConfig{BaseStacks: 2, DRTable: []float64{1.0, 1.7, 2.3, 2.8, 3.2}},
			Talents: map[string]models.TalentConfig{},
		}
	}

	attributors := web.BuildAttributors(config, damageTaken)

	petIDs := make(map[int]bool)
	playerPetIDs := make(map[int]bool)
	for _, a := range allActors {
		if a.petOwner > 0 {
			petIDs[a.id] = true
			if a.petOwner == selectedPlayer.id {
				playerPetIDs[a.id] = true
			}
		}
	}

	pipeline := analysis.NewPipeline(attributors, petIDs, playerPetIDs)
	results := pipeline.Run(events)

	fmt.Println()
	fmt.Print(output.RenderResults(results, selectedFight.name, selectedPlayer.name))
}

func runServe(port string) {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	if port == "" {
		port = os.Getenv("PORT")
	}
	if port == "" {
		port = "8000"
	}

	client := newClient()
	apiRouter := web.NewRouter(client, "data/results")

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Mount("/", apiRouter)

	frontendDir := findFrontendDir()
	if frontendDir != "" {
		assetsDir := filepath.Join(frontendDir, "assets")
		if info, err := os.Stat(assetsDir); err == nil && info.IsDir() {
			r.Handle("/assets/*", http.StripPrefix("/assets/", http.FileServer(http.Dir(assetsDir))))
		}
		indexPath := filepath.Join(frontendDir, "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			r.NotFound(func(w http.ResponseWriter, r *http.Request) {
				if strings.HasPrefix(r.URL.Path, "/api/") {
					http.Error(w, `{"detail":"Not found"}`, 404)
					return
				}
				http.ServeFile(w, r, indexPath)
			})
		}
	}

	log.Info().Str("port", port).Msg("Starting server")
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal().Err(err).Msg("Server failed")
	}
}

func findFrontendDir() string {
	for _, dir := range []string{"../frontend/dist", "frontend/dist"} {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			if _, err := fs.Stat(os.DirFS(dir), "index.html"); err == nil {
				return dir
			}
		}
	}
	return ""
}

func toInt(v any) int {
	switch n := v.(type) {
	case int:
		return n
	case float64:
		return int(n)
	default:
		return 0
	}
}

func toFloat(v any) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	default:
		return 0
	}
}

func toBool(v any) bool {
	b, _ := v.(bool)
	return b
}
