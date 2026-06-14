package scoring

import "github.com/shikherjha/relay-engine/internal/models"

// Scorer is the disposition policy interface (engine-rl-hook). The rule engine
// is the default implementation; an RL policy can be dropped in later without
// touching the HTTP layer — rules remain the safe fallback.
type Scorer interface {
	Score(req models.DispositionRequest) models.DispositionResponse
}

// RuleScorer is the authoritative, explainable rule engine (default).
type RuleScorer struct{}

func (RuleScorer) Score(req models.DispositionRequest) models.DispositionResponse {
	return Score(req)
}

// RLScorer is a placeholder for a learned policy. Until a model is trained it
// delegates to the rule engine and annotates the decision, so behaviour is
// always safe and the interface is demo-ready.
type RLScorer struct {
	Fallback Scorer
}

func (s RLScorer) Score(req models.DispositionRequest) models.DispositionResponse {
	resp := s.Fallback.Score(req)
	resp.Reasons = append(resp.Reasons, "rl-hook: policy not trained → rule fallback")
	return resp
}

// NewScorer selects a policy by name (config DispositionScorer).
func NewScorer(mode string) Scorer {
	switch mode {
	case "rl":
		return RLScorer{Fallback: RuleScorer{}}
	default:
		return RuleScorer{}
	}
}
