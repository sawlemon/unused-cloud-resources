package aws_unused_resources

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	cwTypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	rdsTypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
)

// UnusedRDS holds detailed info about a low-usage RDS instance.
type UnusedRDS struct {
	DBInstanceIdentifier string    // RDS instance identifier
	DBInstanceClass      string    // Instance class (e.g., db.t3.micro)
	Engine               string    // Database engine (e.g., mysql)
	InstanceCreateTime   time.Time // Creation timestamp
	DBInstanceStatus     string    // Current status (e.g., available)
	AvgCPU               float64   // Average CPU utilization (%) over the period
}

// GetUnusedRDSInstances retrieves all RDS instances in a region, computes their average CPU
// usage over the past 'days', and returns details and summary metrics for those below 'threshold'.
func GetUnusedRDSInstances(
	ctx context.Context,
	region string, // AWS region
	threshold float64, // Average CPU utilization (%)
	days int, // Number of days to consider for CPU usage
) (UnusedResourceMetrics, error) {
	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return UnusedResourceMetrics{}, err
	}
	rdsClient := rds.NewFromConfig(cfg)
	cwClient := cloudwatch.NewFromConfig(cfg)

	// List all RDS instances
	instances, err := listAllDBInstances(ctx, rdsClient)
	if err != nil {
		return UnusedResourceMetrics{}, err
	}

	metrics := UnusedResourceMetrics{
		ResourceIDs:          make([]string, 0, len(instances)),
		TotalInstancesCount:  len(instances),
		UnusedInstancesCount: 0,
	}
	var unused []UnusedRDS

	for _, db := range instances {
		avgCPU, err := getAvgCPURDS(ctx, cwClient, *db.DBInstanceIdentifier, days)
		if err != nil {
			continue
		}
		if avgCPU < threshold {
			// Safe dereference of optional fields
			id := ""
			if db.DBInstanceIdentifier != nil {
				id = *db.DBInstanceIdentifier
			}
			class := ""
			if db.DBInstanceClass != nil {
				class = *db.DBInstanceClass
			}
			engine := ""
			if db.Engine != nil {
				engine = *db.Engine
			}
			status := ""
			if db.DBInstanceStatus != nil {
				status = *db.DBInstanceStatus
			}
			createTime := time.Time{}
			if db.InstanceCreateTime != nil {
				createTime = *db.InstanceCreateTime
			}

			unused = append(unused, UnusedRDS{
				DBInstanceIdentifier: id,
				DBInstanceClass:      class,
				Engine:               engine,
				InstanceCreateTime:   createTime,
				DBInstanceStatus:     status,
				AvgCPU:               avgCPU,
			})
			metrics.ResourceIDs = append(metrics.ResourceIDs, id)
			metrics.UnusedInstancesCount++
		}
	}

	return metrics, nil
}

// listAllDBInstances retrieves all RDS DB instances in the account for the given region.
func listAllDBInstances(
	ctx context.Context,
	client *rds.Client,
) ([]rdsTypes.DBInstance, error) {
	paginator := rds.NewDescribeDBInstancesPaginator(client, &rds.DescribeDBInstancesInput{})
	var result []rdsTypes.DBInstance
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		result = append(result, page.DBInstances...)
	}
	return result, nil
}

// getAvgCPURDS retrieves the average CPU utilization for an RDS instance over 'days'.
func getAvgCPURDS(
	ctx context.Context,
	client *cloudwatch.Client,
	dbIdentifier string,
	days int,
) (float64, error) {
	end := time.Now()
	start := end.AddDate(0, 0, -days)
	input := &cloudwatch.GetMetricStatisticsInput{
		Namespace:  aws.String("AWS/RDS"),
		MetricName: aws.String("CPUUtilization"),
		Dimensions: []cwTypes.Dimension{{
			Name:  aws.String("DBInstanceIdentifier"),
			Value: aws.String(dbIdentifier),
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
