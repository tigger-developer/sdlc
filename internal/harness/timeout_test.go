package harness

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"
)

func TestTimeoutRetainsOnlyNativeIdentity(t *testing.T) {
	for _, test := range []struct{ name, output, want string }{
		{"codex", "{\"type\":\"thread.started\",\"thread_id\":\"native-codex\"}\n", "native-codex"},
		{"hermes", "session_id: native-hermes\n", "native-hermes"},
		{"codex", "", ""},
		{"hermes", "", ""},
	} {
		t.Run(test.name+test.want, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
			defer cancel()
			request := Request{Harness: test.name, Model: "fixture", Provider: "fixture", SessionID: "not-native", Directory: t.TempDir()}
			observed := ""
			request.OnSession = func(id string) error { observed = id; return nil }
			executor := func(_ context.Context, _ string, _ []string, _ string, _ io.Reader, stdout, stderr io.Writer) error {
				if test.name == "hermes" {
					stdout = stderr
				}
				// Split output across writes as real provider streams may do.
				for _, part := range strings.SplitAfter(test.output, ":") {
					if _, err := io.WriteString(stdout, part); err != nil {
						return err
					}
				}
				if observed != test.want {
					t.Fatalf("identity was not checkpointed before timeout: %q", observed)
				}
				<-ctx.Done()
				return context.DeadlineExceeded
			}
			_, err := Execute(ctx, request, false, nil, executor, io.Discard)
			var incident *Incident
			if !errors.As(err, &incident) || incident.Kind != "timeout" || incident.SessionID != test.want {
				t.Fatalf("native identity = %#v, want %q", incident, test.want)
			}
		})
	}
}

func TestIdentityCheckpointFailureRejectsProviderResult(t *testing.T) {
	request := Request{Harness: "codex", Model: "fixture", Directory: t.TempDir(), OnSession: func(string) error { return errors.New("record unavailable") }}
	_, err := Execute(context.Background(), request, false, nil, func(_ context.Context, _ string, _ []string, _ string, _ io.Reader, stdout, _ io.Writer) error {
		_, err := io.WriteString(stdout, "{\"type\":\"thread.started\",\"thread_id\":\"native\"}\n")
		return err
	}, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "record unavailable") {
		t.Fatalf("checkpoint error = %v", err)
	}
}
