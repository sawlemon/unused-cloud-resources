package aws_unused_resources

import (
	"context"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	cwTypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	elbv2Types "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
)

// UnusedLoadBalancer holds detailed info about an underutilized load balancer.
type UnusedLoadBalancer struct {
	LoadBalancerArn  string  // ARN of the load balancer
	LoadBalancerName string  // name (e.g., app/my-lb/abc123)
	Type             string  // APPLICATION | NETWORK
	Scheme           string  // internet-facing or internal
	AvgRequestCount  float64 // average daily RequestCount over the period
}

// GetUnusedLoadBalancers lists all ALBs/NLBs, evaluates their average daily RequestCount
// over the past 'days', and returns those below 'threshold' requests per day.
func GetUnusedLoadBalancers(
	ctx context.Context,
	region string,
	threshold float64, // Number of requests per day
	days int, // Number of days to look back
) (UnusedResourceMetrics, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return UnusedResourceMetrics{}, err
	}
	elbv2Client := elasticloadbalancingv2.NewFromConfig(cfg)
	cwClient := cloudwatch.NewFromConfig(cfg)

	lbs, err := listAllLoadBalancers(ctx, elbv2Client)
	if err != nil {
		return UnusedResourceMetrics{}, err
	}

	metrics := UnusedResourceMetrics{
		ResourceIDs:          make([]string, 0, len(lbs)),
		TotalInstancesCount:  len(lbs),
		UnusedInstancesCount: 0,
	}
	var unused []UnusedLoadBalancer

	for _, lb := range lbs {
		arn := aws.ToString(lb.LoadBalancerArn)
		// extract the identifier for CloudWatch dimension
		parts := strings.SplitN(arn, ":loadbalancer/", 2)
		dimVal := ""
		if len(parts) == 2 {
			dimVal = parts[1]
		}
		// choose namespace based on LB type
		namespace := "AWS/ApplicationELB"
		if lb.Type == elbv2Types.LoadBalancerTypeEnumNetwork {
			namespace = "AWS/NetworkELB"
		}
		avgReq, err := getAvgRequestCount(ctx, cwClient, dimVal, namespace, days)
		if err != nil {
			continue
		}
		if avgReq < threshold {
			unused = append(unused, UnusedLoadBalancer{
				LoadBalancerArn:  arn,
				LoadBalancerName: aws.ToString(lb.LoadBalancerName),
				Type:             string(lb.Type),
				Scheme:           string(lb.Scheme),
				AvgRequestCount:  avgReq,
			})
			metrics.ResourceIDs = append(metrics.ResourceIDs, arn)
			metrics.UnusedInstancesCount++
		}
	}

	return metrics, nil
}

// listAllLoadBalancers retrieves all ALBs and NLBs in the account for the given region.
func listAllLoadBalancers(
	ctx context.Context,
	client *elasticloadbalancingv2.Client,
) ([]elbv2Types.LoadBalancer, error) {
	paginator := elasticloadbalancingv2.NewDescribeLoadBalancersPaginator(client, &elasticloadbalancingv2.DescribeLoadBalancersInput{})
	var result []elbv2Types.LoadBalancer
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		result = append(result, page.LoadBalancers...)
	}
	return result, nil
}

// getAvgRequestCount fetches the average daily RequestCount for a load balancer over 'days'.
func getAvgRequestCount(
	ctx context.Context,
	client *cloudwatch.Client,
	lbName string,
	namespace string,
	days int,
) (float64, error) {
	end := time.Now()
	start := end.AddDate(0, 0, -days)
	input := &cloudwatch.GetMetricStatisticsInput{
		Namespace:  aws.String(namespace),
		MetricName: aws.String("RequestCount"),
		Dimensions: []cwTypes.Dimension{{
			Name:  aws.String("LoadBalancer"),
			Value: aws.String(lbName),
		}},
		StartTime: aws.Time(start),
		EndTime:   aws.Time(end),
		Period:    aws.Int32(int32(60 * 60 * 24)), // one datapoint per day
		Statistics: []cwTypes.Statistic{
			cwTypes.StatisticSum,
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
		if dp.Sum != nil {
			sum += *dp.Sum
		}
	}
	return sum / float64(len(resp.Datapoints)), nil
}
