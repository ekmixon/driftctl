package repository

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/applicationautoscaling"
	"github.com/aws/aws-sdk-go/service/applicationautoscaling/applicationautoscalingiface"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
)

type AppAutoScalingRepository interface {
	ServiceNamespaceValues() []string
	DescribeScalableTargets(string) ([]*applicationautoscaling.ScalableTarget, error)
	DescribeScheduledActions(string) ([]*applicationautoscaling.ScheduledAction, error)
}

type appAutoScalingRepository struct {
	client applicationautoscalingiface.ApplicationAutoScalingAPI
	cache  cache.Cache
}

func NewAppAutoScalingRepository(session *session.Session, c cache.Cache) *appAutoScalingRepository {
	return &appAutoScalingRepository{
		applicationautoscaling.New(session),
		c,
	}
}

func (r *appAutoScalingRepository) ServiceNamespaceValues() []string {
	return applicationautoscaling.ServiceNamespace_Values()
}

func (r *appAutoScalingRepository) DescribeScalableTargets(namespace string) ([]*applicationautoscaling.ScalableTarget, error) {
	cacheKey := fmt.Sprintf("appAutoScalingDescribeScalableTargets_%s", namespace)
	if v := r.cache.Get(cacheKey); v != nil {
		return v.([]*applicationautoscaling.ScalableTarget), nil
	}

	input := &applicationautoscaling.DescribeScalableTargetsInput{
		ServiceNamespace: &namespace,
	}
	result, err := r.client.DescribeScalableTargets(input)
	if err != nil {
		return nil, err
	}

	r.cache.Put(cacheKey, result.ScalableTargets)
	return result.ScalableTargets, nil
}

func (r *appAutoScalingRepository) DescribeScheduledActions(namespace string) ([]*applicationautoscaling.ScheduledAction, error) {
	cacheKey := fmt.Sprintf("appAutoScalingDescribeScheduledActions_%s", namespace)
	if v := r.cache.Get(cacheKey); v != nil {
		return v.([]*applicationautoscaling.ScheduledAction), nil
	}

	input := &applicationautoscaling.DescribeScheduledActionsInput{
		ServiceNamespace: &namespace,
	}
	result, err := r.client.DescribeScheduledActions(input)
	if err != nil {
		return nil, err
	}

	r.cache.Put(cacheKey, result.ScheduledActions)
	return result.ScheduledActions, nil
}
