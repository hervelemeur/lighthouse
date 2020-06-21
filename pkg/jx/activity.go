package jx

import (
	"errors"

	v1 "github.com/jenkins-x/jx/v2/pkg/apis/jenkins.io/v1"
	"github.com/jenkins-x/lighthouse/pkg/apis/lighthouse/v1alpha1"
	"github.com/jenkins-x/lighthouse/pkg/record"
)

// ToPipelineState converts the PipelineActivity state to LighthouseJob state
func ToPipelineState(status v1.ActivityStatusType) v1alpha1.PipelineState {
	switch status {
	case v1.ActivityStatusTypePending, v1.ActivityStatusTypeNone:
		return v1alpha1.PendingState
	case v1.ActivityStatusTypeAborted:
		return v1alpha1.AbortedState
	case v1.ActivityStatusTypeRunning:
		return v1alpha1.RunningState
	case v1.ActivityStatusTypeSucceeded:
		return v1alpha1.SuccessState
	case v1.ActivityStatusTypeFailed, v1.ActivityStatusTypeError:
		return v1alpha1.FailureState
	default:
		return v1alpha1.FailureState
	}
}

// ConvertPipelineActivity converts a PipelineActivity from jx to an ActivityRecord
func ConvertPipelineActivity(pa *v1.PipelineActivity) (*record.ActivityRecord, error) {
	if pa == nil {
		return nil, errors.New("pipeline activity is nil")
	}

	sha := pa.Spec.LastCommitSHA
	if sha == "" && pa.Labels != nil {
		sha = pa.Labels[v1.LabelLastCommitSha]
	}

	ar := &record.ActivityRecord{
		Name:            pa.Name,
		Owner:           pa.Spec.GitOwner,
		Repo:            pa.Spec.GitRepository,
		Branch:          pa.Spec.GitBranch,
		BuildIdentifier: pa.Spec.Build,
		BaseSHA:         pa.Spec.BaseSHA,
		Context:         pa.Spec.Context,
		GitURL:          pa.Spec.GitURL,
		LogURL:          pa.Spec.BuildLogsURL,
		Status:          ToPipelineState(pa.Spec.Status),
		StartTime:       pa.Spec.StartedTimestamp,
		CompletionTime:  pa.Spec.CompletedTimestamp,
		Stages:          []*record.ActivityStageOrStep{},
	}

	for _, step := range pa.Spec.Steps {
		if step.Kind == v1.ActivityStepKindTypeStage {
			ar.Stages = append(ar.Stages, convertStage(step.Stage))
		}
	}

	return ar, nil
}

func convertStage(paStage *v1.StageActivityStep) *record.ActivityStageOrStep {
	stage := &record.ActivityStageOrStep{
		Name:           paStage.Name,
		Status:         ToPipelineState(paStage.Status),
		StartTime:      paStage.StartedTimestamp,
		CompletionTime: paStage.CompletedTimestamp,
		Steps:          []*record.ActivityStageOrStep{},
	}

	for _, child := range paStage.Steps {
		stage.Steps = append(stage.Steps, convertStep(child))
	}

	return stage
}

func convertStep(paStep v1.CoreActivityStep) *record.ActivityStageOrStep {
	return &record.ActivityStageOrStep{
		Name:           paStep.Name,
		Status:         ToPipelineState(paStep.Status),
		StartTime:      paStep.StartedTimestamp,
		CompletionTime: paStep.CompletedTimestamp,
		Steps:          []*record.ActivityStageOrStep{},
	}
}
