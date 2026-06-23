package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/kunchenguid/no-mistakes/internal/db"
	"github.com/kunchenguid/no-mistakes/internal/gate"
	"github.com/kunchenguid/no-mistakes/internal/paths"
	"github.com/kunchenguid/no-mistakes/internal/types"
)

func TestRunsCommandShowsAttachableRunIDAndHeadSHA(t *testing.T) {
	setupTestRepo(t)
	p, err := paths.New()
	if err != nil {
		t.Fatal(err)
	}
	d, err := db.Open(p.DB())
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()

	repo, _, err := gate.Init(context.Background(), d, p, ".")
	if err != nil {
		t.Fatal(err)
	}
	run, err := d.InsertRun(repo.ID, "feature/run-id", "1234567890abcdef", "base-sha")
	if err != nil {
		t.Fatal(err)
	}
	if err := d.UpdateRunStatus(run.ID, types.RunRunning); err != nil {
		t.Fatal(err)
	}

	out, err := executeCmd("runs", "--limit", "1")
	if err != nil {
		t.Fatalf("runs command failed: %v", err)
	}
	if !strings.Contains(out, run.ID) {
		t.Fatalf("runs output should include attachable run ID %q, got:\n%s", run.ID, out)
	}
	if !strings.Contains(out, "12345678") {
		t.Fatalf("runs output should include short head SHA, got:\n%s", out)
	}
}

func TestNoActiveRunRecentRunsShowsAttachableRunIDAndHeadSHA(t *testing.T) {
	setupTestRepo(t)
	p, err := paths.New()
	if err != nil {
		t.Fatal(err)
	}
	d, err := db.Open(p.DB())
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()

	repo, _, err := gate.Init(context.Background(), d, p, ".")
	if err != nil {
		t.Fatal(err)
	}
	run, err := d.InsertRun(repo.ID, "feature/recent", "abcdef1234567890", "base-sha")
	if err != nil {
		t.Fatal(err)
	}
	if err := d.UpdateRunStatus(run.ID, types.RunCompleted); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	printNoActiveRun(&out, d, repo.ID)
	got := out.String()
	if !strings.Contains(got, run.ID) {
		t.Fatalf("recent runs output should include attachable run ID %q, got:\n%s", run.ID, got)
	}
	if !strings.Contains(got, "abcdef12") {
		t.Fatalf("recent runs output should include short head SHA, got:\n%s", got)
	}
}
