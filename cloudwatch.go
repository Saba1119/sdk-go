package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

func main() {
	// Create a new session to interact with AWS
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"), // Replace with your desired region
	})
	if err != nil {
		fmt.Println("Error creating session: ", err)
		return
	}

	// Create a CloudWatch Logs client
	logsSvc := cloudwatchlogs.New(sess)

	// Create a new log group
	groupName := "my-log-group"
	_, err = logsSvc.CreateLogGroup(&cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: aws.String(groupName),
	})
	if err != nil {
		fmt.Println("Error creating log group: ", err)
		return
	}

	// Define a filter to monitor logs for the word "error" or "exception"
	filterName := "my-log-filter"
	filterPattern := "error"
	_, err = logsSvc.PutMetricFilter(&cloudwatchlogs.PutMetricFilterInput{
		LogGroupName:  aws.String(groupName),
		FilterName:    aws.String(filterName),
		FilterPattern: aws.String(filterPattern),
		MetricTransformations: []*cloudwatchlogs.MetricTransformation{
			{
				MetricName:      aws.String("ErrorCount"),
				MetricNamespace: aws.String("MyApplication"),
				MetricValue:     aws.String("1"),
			},
		},
	})
	if err != nil {
		fmt.Println("Error creating log filter: ", err)
		return
	}

	// Create a CloudWatch client
	cwSvc := cloudwatch.New(sess)

	// Define an alarm to trigger when CPU usage reaches 5% on an EC2 instance
	metricName := "CPUUtilization"
	namespace := "AWS/EC2"
	alarmName := "MyAlarm"
	instanceId := "i-0ee9c00ee93381f3a"
	_, err = cwSvc.PutMetricAlarm(&cloudwatch.PutMetricAlarmInput{
		AlarmName:          aws.String(alarmName),
		ComparisonOperator: aws.String("GreaterThanThreshold"),
		EvaluationPeriods:  aws.Int64(1),
		MetricName:         aws.String(metricName),
		Namespace:          aws.String(namespace),
		Period:             aws.Int64(60),
		Statistic:          aws.String("Average"),
		Threshold:          aws.Float64(5.0),
		ActionsEnabled:     aws.Bool(true),
		AlarmActions: []*string{
			aws.String(fmt.Sprintf("arn:aws:sns:%s:%s:pipeline", aws.StringValue(sess.Config.Region), "554248189203")),
		},
		OKActions: []*string{
			aws.String(fmt.Sprintf("arn:aws:sns:%s:%s:pipeline", aws.StringValue(sess.Config.Region), "554248189203")),
		},
		Dimensions: []*cloudwatch.Dimension{
			{
				Name:  aws.String("InstanceId"),
				Value: aws.String(instanceId),
			},
		},
	})
	if err != nil {
		fmt.Println("Error creating alarm:", err)
		return
	}

	fmt.Println("Successfully created CloudWatch log group, filter, and alarm!")
}
	
