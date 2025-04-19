package aws_unused_resources

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	cwTypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// UnusedInstance holds detailed info about a low-usage EC2 instance.
type UnusedInstance struct {
	InstanceID   string    // EC2 Instance ID
	InstanceType string    // Instance type (e.g., t3.micro)
	LaunchTime   time.Time // when the instance was started
	State        string    // current state (e.g., running)
	AvgCPU       float64   // average CPU utilization (%) over the period
}

// GetUnusedEC2Instances retrieves all running EC2 instances in a region,
// computes their average CPU use over the past 'days', and returns both
// details and summary metrics for those below 'threshold'.
func GetUnusedEC2Instances(
	ctx context.Context,
	region string,
	threshold float64,
	days int,
) UnusedResourceMetrics {
	// Load AWS config for specified region
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return UnusedResourceMetrics{}
	}
	ec2Client := ec2.NewFromConfig(cfg)
	cwClient := cloudwatch.NewFromConfig(cfg)

	// Fetch all running instances
	instances, err := listRunningInstances(ctx, ec2Client)
	if err != nil {
		return UnusedResourceMetrics{}
	}

	// Initialize metrics
	metrics := UnusedResourceMetrics{
		ResourceIDs:          make([]string, 0, len(instances)),
		TotalInstancesCount:  len(instances),
		UnusedInstancesCount: 0,
	}

	var unused []UnusedInstance
	for _, inst := range instances {
		avgCPU, err := getAvgCPU(ctx, cwClient, *inst.InstanceId, days)
		if err != nil {
			// skip on error
			continue
		}
		if avgCPU < threshold {
			// record detailed instance info
			unused = append(unused, UnusedInstance{
				InstanceID:   *inst.InstanceId,
				InstanceType: string(inst.InstanceType),
				LaunchTime:   *inst.LaunchTime,
				State:        string(inst.State.Name),
				AvgCPU:       avgCPU,
			})
			// update summary metrics
			metrics.ResourceIDs = append(metrics.ResourceIDs, *inst.InstanceId)
			metrics.UnusedInstancesCount++
		}
	}

	return metrics
}

// listRunningInstances returns all EC2 instances currently in the "running" state.
func listRunningInstances(
	ctx context.Context,
	client *ec2.Client,
) ([]ec2Types.Instance, error) {
	input := &ec2.DescribeInstancesInput{
		Filters: []ec2Types.Filter{{
			Name:   aws.String("instance-state-name"),
			Values: []string{"running"},
		}},
	}
	var result []ec2Types.Instance
	paginator := ec2.NewDescribeInstancesPaginator(client, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, res := range page.Reservations {
			result = append(result, res.Instances...)
		}
	}
	return result, nil
}

// getAvgCPU retrieves the average CPU utilization metric for an EC2 instance over 'days'.
func getAvgCPU(
	ctx context.Context,
	client *cloudwatch.Client,
	instanceID string,
	days int,
) (float64, error) {
	end := time.Now()
	start := end.AddDate(0, 0, -days)
	input := &cloudwatch.GetMetricStatisticsInput{
		Namespace:  aws.String("AWS/EC2"),
		MetricName: aws.String("CPUUtilization"),
		Dimensions: []cwTypes.Dimension{{
			Name:  aws.String("InstanceId"),
			Value: aws.String(instanceID),
		}},
		StartTime: aws.Time(start),
		EndTime:   aws.Time(end),
		Period:    aws.Int32(int32(60 * 60 * 24)), // one datapoint per day
		Statistics: []cwTypes.Statistic{
			cwTypes.StatisticAverage,
		},
	}
	resp, err := client.GetMetricStatistics(ctx, input)
	if err != nil {
		return 0, err
	}
	if len(resp.Datapoints) == 0 {
		return 0, nil
	}
	var sum float64
	for _, dp := range resp.Datapoints {
		sum += *dp.Average
	}
	return sum / float64(len(resp.Datapoints)), nil
}
