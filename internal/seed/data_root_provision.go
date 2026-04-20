package seed

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"stormlightlabs.org/baseball/internal/echo"
)

const (
	// DataRepoURLEnvVar overrides the snapshot repo URL used for automatic bootstrap.
	DataRepoURLEnvVar = "BASEBALL_DATA_REPO_URL"

	// DataRepoRefEnvVar optionally pins clone bootstrap to a branch/tag/SHA.
	DataRepoRefEnvVar = "BASEBALL_DATA_REPO_REF"

	// DataAutoCloneEnvVar controls automatic snapshot cloning (default: enabled).
	DataAutoCloneEnvVar = "BASEBALL_DATA_AUTO_CLONE"

	defaultDataRepoURL = "https://github.com/stormlightlabs/bigflydata.git"
)

type dataRootProvision struct {
	rootPath string
	cleanup  func()
}

func ensurePipelineDataRoot(ctx context.Context, opts PipelineOptions) (dataRootProvision, error) {
	missing, err := missingPipelineDataArtifacts(opts)
	if err != nil {
		return dataRootProvision{}, err
	}
	if len(missing) == 0 {
		return dataRootProvision{rootPath: opts.DataRoot, cleanup: func() {}}, nil
	}

	if !shouldAutoCloneDataRoot(opts.DataRoot) || !dataAutoCloneEnabled() {
		return dataRootProvision{}, fmt.Errorf(
			"data root %q is missing required files: %s",
			opts.DataRoot,
			strings.Join(missing, ", "),
		)
	}

	echo.Infof("Data root %s is missing required files: %s", opts.DataRoot, strings.Join(missing, ", "))
	echo.Infof("Bootstrapping temporary snapshot clone from %s", dataRepoURL())

	clonedRoot, cleanup, err := cloneDataSnapshotRepo(ctx, dataRepoURL(), dataRepoRef())
	if err != nil {
		return dataRootProvision{}, err
	}

	clonedOpts := opts
	clonedOpts.DataRoot = clonedRoot
	clonedOpts.LahmanCSVDir = LahmanCSVDir(clonedRoot)
	clonedOpts.RetrosheetDataDir = RetrosheetDir(clonedRoot)
	clonedOpts.FanGraphsDir = FanGraphsDir(clonedRoot)
	clonedOpts.ChadwickDataDir = ChadwickDir(clonedRoot)
	clonedOpts.SalaryDataDir = SalariesDir(clonedRoot)

	missingAfterClone, err := missingPipelineDataArtifacts(clonedOpts)
	if err != nil {
		cleanup()
		return dataRootProvision{}, err
	}
	if len(missingAfterClone) > 0 {
		cleanup()
		return dataRootProvision{}, fmt.Errorf(
			"cloned snapshot is missing required files: %s",
			strings.Join(missingAfterClone, ", "),
		)
	}

	echo.Infof("Using temporary data root %s", clonedRoot)
	return dataRootProvision{
		rootPath: clonedRoot,
		cleanup: func() {
			echo.Infof("Cleaning up temporary data root %s", clonedRoot)
			cleanup()
		},
	}, nil
}

func missingPipelineDataArtifacts(opts PipelineOptions) ([]string, error) {
	missing := make([]string, 0)

	requiredFiles := []string{
		filepath.Join(opts.LahmanCSVDir, "People.csv"),
		filepath.Join(opts.LahmanCSVDir, "Teams.csv"),
		filepath.Join(opts.FanGraphsDir, "woba.csv"),
		filepath.Join(opts.SalaryDataDir, "summary.csv"),
		filepath.Join(opts.RetrosheetDataDir, "allplayers.zip"),
		filepath.Join(opts.RetrosheetDataDir, "biodata.zip"),
	}

	for _, path := range requiredFiles {
		info, err := os.Stat(path)
		if err == nil && !info.IsDir() {
			continue
		}
		if errors.Is(err, os.ErrNotExist) || (err == nil && info.IsDir()) {
			missing = append(missing, path)
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("failed to check %s: %w", path, err)
		}
	}

	parkFactorFiles, err := filepath.Glob(filepath.Join(opts.FanGraphsDir, "pf", "*.csv"))
	if err != nil {
		return nil, fmt.Errorf("failed to inspect park factor files: %w", err)
	}
	if len(parkFactorFiles) == 0 {
		missing = append(missing, filepath.Join(opts.FanGraphsDir, "pf", "*.csv"))
	}

	slices.Sort(missing)
	return missing, nil
}

func shouldAutoCloneDataRoot(root string) bool {
	root = strings.TrimSpace(root)
	if root == "" {
		return true
	}
	cleaned := filepath.Clean(root)

	if cleaned == DefaultDataRoot {
		return true
	}

	slashed := filepath.ToSlash(cleaned)
	return strings.HasSuffix(slashed, "/tools/data")
}

func dataAutoCloneEnabled() bool {
	raw := strings.ToLower(strings.TrimSpace(os.Getenv(DataAutoCloneEnvVar)))
	switch raw {
	case "", "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return true
	}
}

func dataRepoURL() string {
	raw := strings.TrimSpace(os.Getenv(DataRepoURLEnvVar))
	if raw == "" {
		return defaultDataRepoURL
	}
	return raw
}

func dataRepoRef() string {
	return strings.TrimSpace(os.Getenv(DataRepoRefEnvVar))
}

func cloneDataSnapshotRepo(ctx context.Context, repoURL, ref string) (string, func(), error) {
	tmpDir, err := os.MkdirTemp("", "bigflydata-*")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temp directory for snapshot clone: %w", err)
	}

	cleanup := func() { _ = os.RemoveAll(tmpDir) }
	targetDir := filepath.Join(tmpDir, "repo")

	if err := cloneWithGoGit(ctx, repoURL, ref, targetDir); err != nil {
		cleanup()
		return "", nil, err
	}

	lfsCmd := exec.CommandContext(ctx, "git", "-C", targetDir, "lfs", "pull")
	lfsOutput, lfsErr := lfsCmd.CombinedOutput()
	if lfsErr != nil {
		hasPointers, probeErr := hasGitLFSPointers(targetDir)
		if probeErr != nil {
			cleanup()
			return "", nil, fmt.Errorf("failed to inspect cloned snapshot for LFS pointers: %w", probeErr)
		}
		if hasPointers {
			cleanup()
			return "", nil, fmt.Errorf(
				"git-lfs pull failed for cloned snapshot: %w\n%s",
				lfsErr,
				strings.TrimSpace(string(lfsOutput)),
			)
		}
		echo.Infof("Skipping git-lfs pull (no pointer files detected): %s", strings.TrimSpace(string(lfsOutput)))
	}

	return targetDir, cleanup, nil
}

func cloneWithGoGit(ctx context.Context, repoURL, ref, targetDir string) error {
	if ref == "" {
		_, err := git.PlainCloneContext(ctx, targetDir, false, &git.CloneOptions{
			URL:   repoURL,
			Depth: 1,
			Tags:  git.NoTags,
		})
		if err != nil {
			return fmt.Errorf("failed to clone %s: %w", repoURL, err)
		}
		return nil
	}

	if looksLikeCommitHash(ref) {
		repo, err := git.PlainCloneContext(ctx, targetDir, false, &git.CloneOptions{
			URL:        repoURL,
			NoCheckout: true,
			Tags:       git.NoTags,
		})
		if err != nil {
			return fmt.Errorf("failed to clone %s for commit checkout %q: %w", repoURL, ref, err)
		}

		worktree, err := repo.Worktree()
		if err != nil {
			return fmt.Errorf("failed to get worktree for %s: %w", repoURL, err)
		}
		if err := worktree.Checkout(&git.CheckoutOptions{
			Hash:  plumbing.NewHash(ref),
			Force: true,
		}); err != nil {
			return fmt.Errorf("failed to checkout %s at commit %q: %w", repoURL, ref, err)
		}
		return nil
	}

	referenceNames := []plumbing.ReferenceName{}
	if strings.HasPrefix(ref, "refs/") {
		referenceNames = append(referenceNames, plumbing.ReferenceName(ref))
	} else {
		referenceNames = append(referenceNames,
			plumbing.NewBranchReferenceName(ref),
			plumbing.NewTagReferenceName(ref),
		)
	}

	errs := make([]string, 0, len(referenceNames))
	for _, referenceName := range referenceNames {
		tagsMode := git.NoTags
		if strings.HasPrefix(referenceName.String(), "refs/tags/") {
			tagsMode = git.AllTags
		}

		_ = os.RemoveAll(targetDir)
		_, err := git.PlainCloneContext(ctx, targetDir, false, &git.CloneOptions{
			URL:           repoURL,
			Depth:         1,
			SingleBranch:  true,
			ReferenceName: referenceName,
			Tags:          tagsMode,
		})
		if err == nil {
			return nil
		}
		errs = append(errs, fmt.Sprintf("%s: %v", referenceName, err))
	}

	return fmt.Errorf("failed to clone %s at ref %q: %s", repoURL, ref, strings.Join(errs, "; "))
}

func looksLikeCommitHash(ref string) bool {
	if len(ref) != 40 {
		return false
	}
	for _, ch := range ref {
		switch {
		case ch >= '0' && ch <= '9':
		case ch >= 'a' && ch <= 'f':
		case ch >= 'A' && ch <= 'F':
		default:
			return false
		}
	}
	return true
}

func hasGitLFSPointers(root string) (bool, error) {
	datasetRoots := []string{
		filepath.Join(root, "lahman"),
		filepath.Join(root, "retrosheet"),
		filepath.Join(root, "fangraphs"),
		filepath.Join(root, "salaries"),
		filepath.Join(root, "chadwick"),
	}

	for _, datasetRoot := range datasetRoots {
		info, err := os.Stat(datasetRoot)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return false, err
		}
		if !info.IsDir() {
			continue
		}

		var found bool
		walkErr := filepath.WalkDir(datasetRoot, func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if d.IsDir() {
				return nil
			}
			ok, err := isGitLFSPointerFile(path)
			if err != nil {
				return err
			}
			if ok {
				found = true
				return fs.SkipAll
			}
			return nil
		})
		if walkErr != nil && !errors.Is(walkErr, fs.SkipAll) {
			return false, walkErr
		}
		if found {
			return true, nil
		}
	}

	return false, nil
}

func isGitLFSPointerFile(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()

	buf := make([]byte, 256)
	n, err := io.ReadFull(file, buf)
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
		return false, err
	}
	content := string(buf[:n])
	return strings.HasPrefix(content, "version https://git-lfs.github.com/spec/v1\n"), nil
}
