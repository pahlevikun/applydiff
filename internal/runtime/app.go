package runtime

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pahlevikun/applydiff/internal/cv"
	"github.com/pahlevikun/applydiff/internal/jobconfig"
	"github.com/pahlevikun/applydiff/internal/llm"
	"github.com/pahlevikun/applydiff/internal/scan"
	"github.com/pahlevikun/applydiff/internal/source"
	"github.com/pahlevikun/applydiff/internal/source/greenhouse"
	"github.com/pahlevikun/applydiff/internal/store"
)

// App wires core dependencies for CLI, TUI, and HTTP server.
type App struct {
	Store  *store.Store
	Config *jobconfig.Config
	Runner *scan.Runner
	LLM    llm.Runner
	Facts  *cv.Facts

	DBPath   string
	CfgPath  string
	FactsPath string
}

// Options configures bootstrap paths.
type Options struct {
	DBPath    string
	ConfigPath string
	FactsPath string
}

// Open creates a migrated store and loads config from the environment / defaults.
func Open(ctx context.Context, opt Options) (*App, error) {
	dbPath := opt.DBPath
	if dbPath == "" {
		dbPath = os.Getenv("DATABASE_PATH")
		if dbPath == "" {
			dbPath = "applydiff.db"
		}
	}
	cfgPath := opt.ConfigPath
	if cfgPath == "" {
		cfgPath = os.Getenv("JOBS_CONFIG")
		if cfgPath == "" {
			cfgPath = filepath.Join("jobs", "config.yaml")
		}
	}
	factsPath := opt.FactsPath
	if factsPath == "" {
		factsPath = os.Getenv("FACTS_PATH")
		if factsPath == "" {
			factsPath = "facts.json"
		}
	}

	cfg, err := jobconfig.Load(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("load jobs config: %w", err)
	}
	st, err := store.New(dbPath)
	if err != nil {
		return nil, err
	}
	if err := st.Migrate(ctx); err != nil {
		st.Close()
		return nil, err
	}
	registerSources(cfg)

	var facts *cv.Facts
	if data, err := os.ReadFile(factsPath); err == nil {
		facts, _ = cv.LoadFactsJSON(data)
	}

	app := &App{
		Store:     st,
		Config:    cfg,
		DBPath:    dbPath,
		CfgPath:   cfgPath,
		FactsPath: factsPath,
		Runner: &scan.Runner{
			Store:  st,
			Config: cfg,
		},
		LLM:   llm.NewRunnerFromEnv(),
		Facts: facts,
	}
	return app, nil
}

func registerSources(cfg *jobconfig.Config) {
	if cfg.Sources.Greenhouse.Enabled {
		source.Register(greenhouse.NewFromConfig(cfg.Sources.Greenhouse))
	}
}

// FactsOrFile returns facts from store or the bootstrap file.
func (a *App) FactsOrFile(ctx context.Context) *cv.Facts {
	if a == nil {
		return nil
	}
	facts, err := a.Store.LoadFacts(ctx)
	if err == nil {
		return facts
	}
	return a.Facts
}

// Close releases the database.
func (a *App) Close() error {
	if a == nil || a.Store == nil {
		return nil
	}
	return a.Store.Close()
}
