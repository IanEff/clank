package clank

import (
	"context"
	"encoding/json"
)

type Tool interface {
	Spec() ToolSpec
	Run(ctx context.Context, args json.RawMessage) (EvidenceRef, error)
}

// ProposeToolSpec is the model's terminal `propose` tool: the leading
// FailureClass, the competing hypotheses, and the candidate actions (each drawn
// from the catalog). Its input schema is generated from proposeInput, so the
// shape the model is held to is the shape the engine decodes.
func ProposeToolSpec() ToolSpec {
	return ToolSpec{
		Name:        "propose",
		Description: "Emit your final answer: the leading failure class, the competing hypotheses, and the candidate actions, each drawn from the action catalog.",
		InputSchema: SchemaOf[proposeInput](),
	}
}
