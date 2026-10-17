package export

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pahlevikun/applydiff/internal/paths"
)

// Engine compiles LaTeX to PDF.
type Engine interface {
	Name() string
	Compile(texPath, workDir, pdfPath string) error
}

var defaultEngine Engine

// SetDefaultEngine configures the process-wide PDF engine (called from runtime.Open).
func SetDefaultEngine(e Engine) {
	defaultEngine = e
}

// ResolveEngine picks configured > xelatex on PATH > tectonic in Paths.Bin or PATH.
func ResolveEngine(p paths.Paths) (Engine, error) {
	if v := os.Getenv("APPLYDIFF_PDF_ENGINE"); v != "" {
		switch strings.ToLower(v) {
		case "xelatex":
			if _, err := exec.LookPath("xelatex"); err != nil {
				return nil, fmt.Errorf("APPLYDIFF_PDF_ENGINE=xelatex but xelatex is not on PATH")
			}
			return xelatexEngine{}, nil
		case "tectonic":
			return tectonicEngineAt(p.Bin)
		default:
			return nil, fmt.Errorf("unknown APPLYDIFF_PDF_ENGINE %q (use xelatex or tectonic)", v)
		}
	}
	if _, err := exec.LookPath("xelatex"); err == nil {
		return xelatexEngine{}, nil
	}
	eng, err := tectonicEngineAt(p.Bin)
	if err != nil {
		return nil, fmt.Errorf("no PDF engine: install xelatex or tectonic")
	}
	return eng, nil
}

// DetectPDFEngine reports the active or would-be engine name.
func DetectPDFEngine() string {
	if defaultEngine != nil {
		return defaultEngine.Name()
	}
	p, err := paths.Resolve()
	if err != nil {
		p = paths.Paths{}
	}
	eng, err := ResolveEngine(p)
	if err != nil {
		return "none"
	}
	return eng.Name()
}

// CompilePDF uses the process default engine when set, otherwise ResolveEngine.
func CompilePDF(texPath, workDir, pdfPath string) error {
	eng := defaultEngine
	if eng == nil {
		p, err := paths.Resolve()
		if err != nil {
			p = paths.Paths{}
		}
		eng, err = ResolveEngine(p)
		if err != nil {
			return err
		}
	}
	return eng.Compile(texPath, workDir, pdfPath)
}

type xelatexEngine struct{}

func (xelatexEngine) Name() string { return "xelatex" }

func (xelatexEngine) Compile(texPath, workDir, pdfPath string) error {
	return runXeLaTeXLocal(texPath, workDir, pdfPath)
}

type tectonicEngine struct {
	bin string
}

func (e tectonicEngine) Name() string { return "tectonic" }

func (e tectonicEngine) Compile(texPath, workDir, pdfPath string) error {
	if err := os.MkdirAll(filepath.Dir(pdfPath), 0o755); err != nil {
		return err
	}
	outDir := filepath.Dir(pdfPath)
	args := []string{
		"-X", "compile",
		"--keep-logs",
		"--outdir", outDir,
		texPath,
	}
	cmd := exec.Command(e.bin, args...)
	cmd.Dir = workDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("tectonic: %w: %s", err, truncate(string(out), 400))
	}
	base := strings.TrimSuffix(filepath.Base(texPath), filepath.Ext(texPath)) + ".pdf"
	got := filepath.Join(outDir, base)
	if got != pdfPath && fileExists(got) {
		return os.Rename(got, pdfPath)
	}
	return nil
}

func tectonicEngineAt(binDir string) (Engine, error) {
	name := "tectonic"
	if runtime.GOOS == "windows" {
		name = "tectonic.exe"
	}
	local := filepath.Join(binDir, name)
	if fileExists(local) {
		return tectonicEngine{bin: local}, nil
	}
	if p, err := exec.LookPath("tectonic"); err == nil {
		return tectonicEngine{bin: p}, nil
	}
	return nil, fmt.Errorf("tectonic not found in %s or PATH", binDir)
}
