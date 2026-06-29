package clank

import (
	"context"
	"fmt"
	"io"

	"sigs.k8s.io/yaml"
)

type ProposalSink interface {
	Deliver(ctx context.Context, ps ProposalSet) error
}

type MarkdownSink struct {
	W io.Writer
}

func (s *MarkdownSink) Deliver(_ context.Context, ps ProposalSet) error {
	if _, err := fmt.Fprintf(s.W, "## ProposalSet: %s (%d considered)\n", ps.FailureClass, len(ps.Proposals)); err != nil {
		return err
	}
	for _, c := range ps.Proposals {
		if c.ID == ps.Recommended {
			_, err := fmt.Fprintf(s.W, "**Recommended:** %s — %s\n", c.ID, c.ContractRef)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

type YAMLSink struct {
	W io.Writer
}

func (s *YAMLSink) Deliver(_ context.Context, ps ProposalSet) error {
	out, err := yaml.Marshal(ps)
	if err != nil {
		return fmt.Errorf("yaml sink: marshal proposal set: %w", err)
	}
	if _, err := s.W.Write(out); err != nil {
		return fmt.Errorf("yaml sink: write: %w", err)
	}
	return nil
}
