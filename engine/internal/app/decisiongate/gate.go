package decisiongate

import (
	"errors"
	"strings"
	"time"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/domain"
)

type UserSteering string

const (
	UserSteeringNone  UserSteering = ""
	UserSteeringPivot UserSteering = "pivot"
	UserSteeringStop  UserSteering = "stop"
)

type Input struct {
	DecisionID        string
	MissionID         string
	Stage             domain.StageName
	Evidence          []domain.Evidence
	UserSteering      UserSteering
	NewHypothesisSeed string
	Now               time.Time
}

func Evaluate(input Input) (domain.Decision, error) {
	if strings.TrimSpace(input.DecisionID) == "" {
		return domain.Decision{}, errors.New("decision id is required")
	}
	if strings.TrimSpace(input.MissionID) == "" {
		return domain.Decision{}, errors.New("mission id is required")
	}
	if !input.Stage.IsValid() {
		return domain.Decision{}, errors.New("stage is invalid")
	}

	evidenceRefs := evidenceIDs(input.Evidence)
	switch {
	case input.UserSteering == UserSteeringPivot:
		return decision(input, domain.DecisionPivot, confidence(input.Evidence), evidenceRefs, []string{
			"user_requested_pivot",
		}, "用户请求调整方向。", defaultPivotAction(input.NewHypothesisSeed)), nil
	case input.UserSteering == UserSteeringStop:
		return decision(input, domain.DecisionKill, confidence(input.Evidence), evidenceRefs, []string{
			"user_requested_stop",
		}, "用户请求停止当前 Mission。", "停止后保留事件记录，等待用户创建新的 Mission。"), nil
	case len(input.Evidence) == 0:
		return decision(input, domain.DecisionAskUser, 0, nil, []string{
			"missing_evidence",
		}, "当前阶段缺少证据，不能静默继续。", "补充至少一条可信证据或让用户确认下一步。"), nil
	case hasConflictingEvidence(input.Evidence):
		return decision(input, domain.DecisionAskUser, confidence(input.Evidence), evidenceRefs, []string{
			"conflicting_evidence",
		}, "当前阶段证据存在冲突，需要用户判断。", "对比冲突证据后选择继续验证、调整假设或暂停。"), nil
	case hasHighRiskWithLowConfidence(input.Evidence):
		return decision(input, domain.DecisionPause, confidence(input.Evidence), evidenceRefs, []string{
			"low_confidence",
			"high_risk",
		}, "当前阶段置信度不足且风险较高。", "暂停推进，先补充更高质量证据。"), nil
	case confidence(input.Evidence) >= 0.7:
		return decision(input, domain.DecisionContinue, confidence(input.Evidence), evidenceRefs, nil, "当前阶段证据达到继续推进阈值。", "进入下一阶段并保留当前证据链。"), nil
	default:
		return decision(input, domain.DecisionAskUser, confidence(input.Evidence), evidenceRefs, []string{
			"insufficient_confidence",
		}, "当前阶段证据不足以自动继续。", "补充证据或由用户确认是否继续。"), nil
	}
}

func decision(
	input Input,
	decisionType domain.DecisionType,
	confidence float64,
	evidenceRefs []string,
	risks []string,
	reason string,
	nextBestAction string,
) domain.Decision {
	return domain.Decision{
		ID:             input.DecisionID,
		MissionID:      input.MissionID,
		Stage:          input.Stage,
		Type:           decisionType,
		Confidence:     confidence,
		Reason:         reason,
		EvidenceRefs:   evidenceRefs,
		Risks:          risks,
		NextBestAction: nextBestAction,
		CreatedAt:      input.Now,
	}
}

func confidence(evidence []domain.Evidence) float64 {
	if len(evidence) == 0 {
		return 0
	}
	var sum float64
	for _, item := range evidence {
		sum += item.Confidence
	}
	return sum / float64(len(evidence))
}

func evidenceIDs(evidence []domain.Evidence) []string {
	ids := make([]string, 0, len(evidence))
	for _, item := range evidence {
		ids = append(ids, item.ID)
	}
	return ids
}

func hasHighRiskWithLowConfidence(evidence []domain.Evidence) bool {
	return confidence(evidence) < 0.7 && highestRisk(evidence) >= riskWeight(domain.RiskHigh)
}

func hasConflictingEvidence(evidence []domain.Evidence) bool {
	hasStrongSupport := false
	hasHighRisk := false
	for _, item := range evidence {
		if item.Confidence >= 0.7 && riskWeight(item.Risk) <= riskWeight(domain.RiskMedium) {
			hasStrongSupport = true
		}
		if item.Confidence >= 0.7 && riskWeight(item.Risk) >= riskWeight(domain.RiskHigh) {
			hasHighRisk = true
		}
	}
	return hasStrongSupport && hasHighRisk
}

func highestRisk(evidence []domain.Evidence) int {
	highest := 0
	for _, item := range evidence {
		if weight := riskWeight(item.Risk); weight > highest {
			highest = weight
		}
	}
	return highest
}

func riskWeight(risk domain.RiskLevel) int {
	switch risk {
	case domain.RiskLow:
		return 1
	case domain.RiskMedium:
		return 2
	case domain.RiskHigh:
		return 3
	case domain.RiskCritical:
		return 4
	default:
		return 0
	}
}

func defaultPivotAction(seed string) string {
	if strings.TrimSpace(seed) == "" {
		return "基于现有证据生成新的核心假设。"
	}
	return "围绕新的假设种子继续验证：" + seed
}
