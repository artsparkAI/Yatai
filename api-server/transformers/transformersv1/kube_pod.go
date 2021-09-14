package transformersv1

import (
	"context"
	"sort"
	"strings"

	apiv1 "k8s.io/api/core/v1"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/common/consts"
	"github.com/bentoml/yatai/schemas/modelschemas"
	"github.com/bentoml/yatai/schemas/schemasv1"
)

func ToPodSchema(ctx context.Context, pod *models.KubePodWithStatus) (v *schemasv1.KubePodSchema, err error) {
	vs, err := ToPodSchemas(ctx, []*models.KubePodWithStatus{pod})
	if err != nil {
		return nil, err
	}
	return vs[0], nil
}

func ToPodSchemas(ctx context.Context, pods []*models.KubePodWithStatus) (vs []*schemasv1.KubePodSchema, err error) {
	sort.SliceStable(pods, func(i, j int) bool {
		iName := pods[i].Pod.Name
		jName := pods[j].Pod.Name

		return strings.Compare(iName, jName) >= 0
	})

	sort.SliceStable(pods, func(i, j int) bool {
		it := pods[i].Pod.Status.StartTime
		jt := pods[j].Pod.Status.StartTime

		if it == nil {
			return false
		}

		if jt == nil {
			return true
		}

		return it.Before(jt)
	})

	sort.SliceStable(pods, func(i, j int) bool {
		return pods[i].Pod.Labels[consts.KubeLabelYataiDeploymentSnapshotType] == string(modelschemas.DeploymentSnapshotTypeStable)
	})

	for _, p := range pods {
		var statusReady bool
		for _, c := range p.Pod.Status.Conditions {
			if c.Type == apiv1.PodReady {
				statusReady = c.Status == apiv1.ConditionTrue
			}
		}
		isCanary := p.Pod.Labels[consts.KubeLabelYataiDeploymentSnapshotType] == string(modelschemas.DeploymentSnapshotTypeCanary)
		status := schemasv1.KubePodStatusSchema{
			Phase:     p.Pod.Status.Phase,
			Ready:     statusReady,
			StartTime: p.Pod.Status.StartTime,
			IsOld:     false,
			IsCanary:  isCanary,
			HostIp:    p.Pod.Status.HostIP,
		}
		vs = append(vs, &schemasv1.KubePodSchema{
			Name:      p.Pod.Name,
			NodeName:  p.Pod.Spec.NodeName,
			Status:    status,
			RawStatus: p.Pod.Status,
			PodStatus: p.Status,
			Warnings:  p.Warnings,
		})
	}
	return
}
