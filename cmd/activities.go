package main

import (
	"context"

	"github.com/sirupsen/logrus"
	"go.temporal.io/sdk/activity"
)

func Activity1(ctx context.Context, input map[string]interface{}) (interface{}, error) {
	activityInfo := activity.GetInfo(ctx)
	taskToken := string(activityInfo.TaskToken)
	activityName := activityInfo.ActivityType.Name

	logger.WithFields(logrus.Fields{"input": input, "taskToken": taskToken, "activityName": activityName}).Info("activity executed")

	return input, nil
}

func Activity2(ctx context.Context, input map[string]interface{}) (interface{}, error) {
	activityInfo := activity.GetInfo(ctx)
	taskToken := string(activityInfo.TaskToken)
	activityName := activityInfo.ActivityType.Name

	logger.WithFields(logrus.Fields{"input": input, "taskToken": taskToken, "activityName": activityName}).Info("activity executed")

	return input, nil
}
