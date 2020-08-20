package builders

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/filecoin-project/test-vectors/schema"
)

// Generator is a batch generator and organizer of test vectors.
//
// Test vector scripts are simple programs (main function). Test vector scripts
// can delegate to the Generator to handle the execution, reporting and capture
// of emitted test vectors into files.
//
// Generator supports the following CLI flags:
//
//  -o <directory>
//		directory where test vector JSON files will be saved; if omitted,
//		vectors will be written to stdout.
//
//  -u
//		update any existing test vector files in the output directory IF their
//		content has changed. Note `_meta` is ignored when checking equality.
//
//  -f
//		force regeneration and overwrite any existing vectors in the output
//		directory.
//
//  -i <include regex>
//		regex inclusion filter to select a subset of vectors to execute; matched
//		against the vector's ID.
//
// Scripts can bundle test vectors into "groups". The generator will execute
// each group in parallel, and will write each vector in a file:
// <output_dir>/<group>--<vector_id>.json
type Generator struct {
	OutputPath    string
	Mode          OverwriteMode
	IncludeFilter *regexp.Regexp

	wg sync.WaitGroup
}

// OverwriteMode is the mode used when overwriting existing test vector files.
type OverwriteMode int

const (
	// OverwriteNone will not overwrite existing test vector files.
	OverwriteNone OverwriteMode = iota
	// OverwriteUpdate will update test vector files if they're different.
	OverwriteUpdate
	// OverwriteForce will force overwrite the vector files.
	OverwriteForce
)

var GenscriptCommit = "dirty"

// genData is the generation data to stamp into vectors.
var genData = []schema.GenerationData{
	{
		Source:  "genscript",
		Version: GenscriptCommit,
	},
}

func init() {
	genData = append(genData, getBuildInfo()...)
}

func getBuildInfo() []schema.GenerationData {
	deps := []string{"github.com/filecoin-project/lotus", "github.com/filecoin-project/specs-actors"}

	bi, ok := debug.ReadBuildInfo()
	if !ok {
		panic("cant read build info")
	}

	var result []schema.GenerationData

	for _, v := range bi.Deps {
		for _, dep := range deps {
			if strings.HasPrefix(v.Path, dep) {
				result = append(result, schema.GenerationData{Source: v.Path, Version: v.Version})
			}
		}
	}

	return result
}

type MessageVectorGenItem struct {
	Metadata *schema.Metadata
	Selector schema.Selector
	Func     func(*Builder)
}

func NewGenerator() *Generator {
	// Consume CLI parameters.
	var outputDir string
	const outputDirUsage = "directory where test vector JSON files will be saved; if omitted, vectors will be written to stdout."
	flag.StringVar(&outputDir, "o", "", outputDirUsage)
	flag.StringVar(&outputDir, "out", "", outputDirUsage)

	var update bool
	const updateUsage = "update any existing test vector files in the output directory IF their content has changed. Note `_meta` is ignored when checking equality."
	flag.BoolVar(&update, "u", false, updateUsage)
	flag.BoolVar(&update, "update", false, updateUsage)

	var force bool
	const forceUsage = "force regeneration and overwrite any existing vectors in the output directory."
	flag.BoolVar(&force, "f", false, forceUsage)
	flag.BoolVar(&force, "force", false, forceUsage)

	var includeFilter string
	const includeFilterUsage = "regex inclusion filter to select a subset of vectors to execute; matched against the vector's ID."
	flag.StringVar(&includeFilter, "i", "", includeFilterUsage)
	flag.StringVar(&includeFilter, "include", "", includeFilterUsage)

	flag.Parse()

	mode := OverwriteNone
	if force {
		mode = OverwriteForce
	} else if update {
		mode = OverwriteUpdate
	}
	ret := Generator{Mode: mode}

	// If output directory is provided, we ensure it exists, or create it.
	// Else, we'll output to stdout.
	if outputDir != "" {
		err := ensureDirectory(outputDir)
		if err != nil {
			log.Fatal(err)
		}
		ret.OutputPath = outputDir
	}

	// If a filter has been provided, compile it into a regex.
	if includeFilter != "" {
		exp, err := regexp.Compile(includeFilter)
		if err != nil {
			log.Fatalf("supplied inclusion filter regex %s is invalid: %s", includeFilter, err)
		}
		ret.IncludeFilter = exp
	}

	return &ret
}

func (g *Generator) Wait() {
	g.wg.Wait()
}

func (g *Generator) MessageVectorGroup(group string, vectors ...*MessageVectorGenItem) {
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()

		var tmpOutDir string
		if g.OutputPath != "" {
<<<<<<< HEAD
			dir, err := ioutil.TempDir(os.TempDir(), group)
=======
			p, err := ioutil.TempDir("", group)
>>>>>>> e85364b85b161aee0ff9e2f3d06e00a344b32203
			if err != nil {
				log.Printf("failed to create temp output directory: %s", err)
				return
			}
			defer func() {
				if err := os.RemoveAll(dir); err != nil {
					log.Printf("failed to remove temp output directory: %s", err)
				}
			}()
			tmpOutDir = dir
		}

		var wg sync.WaitGroup
		for _, item := range vectors {
			if id := item.Metadata.ID; g.IncludeFilter != nil && !g.IncludeFilter.MatchString(id) {
				log.Printf("skipping %s: does not match inclusion filter", id)
				continue
			}

			filename := fmt.Sprintf("%s--%s.json", group, item.Metadata.ID)
			tmpFilePath := filepath.Join(tmpOutDir, filename)
			var w io.Writer
			if g.OutputPath == "" {
				w = os.Stdout
			} else {
				out, err := os.OpenFile(tmpFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
				if err != nil {
					log.Printf("failed to open file %s: %s", tmpFilePath, err)
					return
				}
				w = out
			}

			wg.Add(1)
			go func(item *MessageVectorGenItem) {
				defer wg.Done()
				g.generateOne(w, item, w != os.Stdout)

				if g.OutputPath != "" {
					outFilePath := filepath.Join(g.OutputPath, filename)
					_, err := os.Stat(outFilePath)
					exists := !os.IsNotExist(err)

					// if file (probably) exists and we're not force overwriting it, check equality
					if exists && g.Mode != OverwriteForce {
						eql, err := g.vectorsEqual(tmpFilePath, outFilePath)
						if err != nil {
							log.Printf("failed to check new vs existing vector equality: %s", err)
							return
						}
						if eql {
							log.Printf("not writing %s: no changes", item.Metadata.ID)
							return
						}
						if g.Mode == OverwriteNone {
							log.Printf("⚠️ WARNING: not writing %s: vector changed, use -u or -f to overwrite", item.Metadata.ID)
							return
						}
					}
					// Move vector from tmp dir to final location
					if err := os.Rename(tmpFilePath, outFilePath); err != nil {
						log.Printf("failed to move generated test vector: %s", err)
					}
					log.Printf("wrote test vector: %s", outFilePath)
				}
			}(item)
		}

		wg.Wait()
	}()
}

// parseVectorFile unnmarshals a JSON serialized test vector stored at the
// given file path and returns it.
func (g *Generator) parseVectorFile(p string) (*schema.TestVector, error) {
	raw, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("reading test vector file: %w", err)
	}
	var vector schema.TestVector
	err = json.Unmarshal(raw, &vector)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling test vector: %w", err)
	}
	return &vector, nil
}

// vectorBytesNoMeta parses the vector at the given file path and returns the
// serialized bytes for the vector after stripping the metadata.
func (g *Generator) vectorBytesNoMeta(p string) ([]byte, error) {
	v, err := g.parseVectorFile(p)
	if err != nil {
		return nil, err
	}
	v.Meta = nil
	return json.Marshal(v)
}

// vectorsEqual determines if two vectors are "equal". They are considered
// equal if they serialize to the same bytes without a `_meta` property.
func (g *Generator) vectorsEqual(apath, bpath string) (bool, error) {
	abytes, err := g.vectorBytesNoMeta(apath)
	if err != nil {
		return false, err
	}
	bbytes, err := g.vectorBytesNoMeta(bpath)
	if err != nil {
		return false, err
	}
	return bytes.Equal(abytes, bbytes), nil
}

func (g *Generator) generateOne(w io.Writer, b *MessageVectorGenItem, indent bool) {
	log.Printf("generating test vector: %s", b.Metadata.ID)

	// stamp with our generation data.
	b.Metadata.Gen = genData

	vector := MessageVector(b.Metadata, b.Selector)

	// TODO: currently if an assertion fails, we call os.Exit(1), which
	//  aborts all ongoing vector generations. The Asserter should
	//  call runtime.Goexit() instead so only that goroutine is
	//  cancelled. The assertion error must bubble up somehow.
	b.Func(vector)

	buf := new(bytes.Buffer)
	vector.Finish(buf)

	final := buf
	if indent {
		// reparse and reindent.
		final = new(bytes.Buffer)
		if err := json.Indent(final, buf.Bytes(), "", "\t"); err != nil {
			log.Printf("failed to indent json: %s", err)
		}
	}

	n, err := w.Write(final.Bytes())
	if err != nil {
		log.Printf("failed to write to output: %s", err)
		return
	}

	log.Printf("generated test vector: %s (size: %d bytes)", b.Metadata.ID, n)
}

// ensureDirectory checks if the provided path is a directory. If yes, it
// returns nil. If the path doesn't exist, it creates the directory and
// returns nil. If the path is not a directory, or another error occurs, an
// error is returned.
func ensureDirectory(path string) error {
	switch stat, err := os.Stat(path); {
	case os.IsNotExist(err):
		// create directory.
		log.Printf("creating directory %s", path)
		err := os.MkdirAll(path, 0700)
		if err != nil {
			return fmt.Errorf("failed to create directory %s: %s", path, err)
		}

	case err == nil && !stat.IsDir():
		return fmt.Errorf("path %s exists, but it's not a directory", path)

	case err != nil:
		return fmt.Errorf("failed to stat directory %s: %w", path, err)
	}
	return nil
}
