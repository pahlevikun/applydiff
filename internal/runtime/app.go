package runtime

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pahlevikun/applydiff/internal/cv"
	"github.com/pahlevikun/applydiff/internal/jobconfig"
	"github.com/pahlevikun/applydiff/internal/export"
	"github.com/pahlevikun/applydiff/internal/llm"
	"github.com/pahlevikun/applydiff/internal/paths"
	"github.com/pahlevikun/applydiff/internal/runs"
	"github.com/pahlevikun/applydiff/internal/scan"
	"github.com/pahlevikun/applydiff/internal/source"
	"github.com/pahlevikun/applydiff/internal/source/fourdayweek"
	"github.com/pahlevikun/applydiff/internal/source/greenhouse"
	"github.com/pahlevikun/applydiff/internal/source/jobicy"
	"github.com/pahlevikun/applydiff/internal/source/remotive"
	"github.com/pahlevikun/applydiff/internal/source/remoteok"
	"github.com/pahlevikun/applydiff/internal/source/weworkremotely"
	"github.com/pahlevikun/applydiff/internal/source/workingnomads"
	"github.com/pahlevikun/applydiff/internal/store"
)

// App wires core dependencies for CLI, TUI, and HTTP server.
type App struct {
	Store  *store.Store
	Config *jobconfig.Config
	Runner *scan.Runner
	Runs   *runs.Runner
	LLM    llm.Runner
	Facts  *cv.Facts

	Paths     paths.Paths
	DBPath    string
	CfgPath   string
	FactsPath string
}

// Options configures bootstrap paths.
type Options struct {
	DBPath     string
	ConfigPath string
	FactsPath  string
}

// Open creates a migrated store and loads config from the environment / defaults.
func Open(ctx context.Context, opt Options) (*App, error) {
	_ = EnsureLoginShellPATH()

	p, err := paths.Resolve()
	if err != nil {
		return nil, err
	}
	if opt.DBPath != "" {
		p.DB = opt.DBPath
	}
	if opt.ConfigPath != "" {
		p.Config = opt.ConfigPath
	}
	if opt.FactsPath != "" {
		p.Facts = opt.FactsPath
	}
	if err := paths.EnsureDefaults(p); err != nil {
		return nil, fmt.Errorf("prepare data dir: %w", err)
	}

	cfg, err := jobconfig.Load(p.Config)
	if err != nil {
		return nil, fmt.Errorf("load jobs config: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(p.DB), 0o755); err != nil {
		return nil, err
	}
	st, err := store.New(p.DB)
	if err != nil {
		return nil, err
	}
	if err := st.Migrate(ctx); err != nil {
		st.Close()
		return nil, err
	}
	runner := runs.New(st)
	_ = runner.Recover(ctx)
	if eng, err := export.ResolveEngine(p); err == nil {
		export.SetDefaultEngine(eng)
	}
	registerSources(cfg)

	var facts *cv.Facts
	if data, err := os.ReadFile(p.Facts); err == nil {
		facts, _ = cv.LoadFactsJSON(data)
	}

	app := &App{
		Store:     st,
		Config:    cfg,
		Paths:     p,
		DBPath:    p.DB,
		CfgPath:   p.Config,
		FactsPath: p.Facts,
		Runner: &scan.Runner{
			Store:  st,
			Config: cfg,
		},
		Runs: runner,
		LLM:   llm.NewRunnerFromEnv(),
		Facts: facts,
	}
	return app, nil
}

func registerSources(cfg *jobconfig.Config) {
	source.Reset()
	if cfg.Sources.Greenhouse.Enabled {
		source.Register(greenhouse.NewFromConfig(cfg.Sources.Greenhouse))
	}
	if cfg.Sources.Remotive.Enabled {
		source.Register(remotive.NewFromConfig(cfg.Sources.Remotive))
	}
	if cfg.Sources.RemoteOK.Enabled {
		source.Register(remoteok.NewFromConfig(cfg.Sources.RemoteOK))
	}
	if cfg.Sources.Jobicy.Enabled {
		source.Register(jobicy.NewFromConfig(cfg.Sources.Jobicy))
	}
	if cfg.Sources.WeWorkRemotely.Enabled {
		source.Register(weworkremotely.NewFromConfig(cfg.Sources.WeWorkRemotely))
	}
	if cfg.Sources.WorkingNomads.Enabled {
		source.Register(workingnomads.NewFromConfig(cfg.Sources.WorkingNomads))
	}
	if cfg.Sources.FourDayWeek.Enabled {
		source.Register(fourdayweek.NewFromConfig(cfg.Sources.FourDayWeek))
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
