// Package s3unused provides functions to detect underutilized S3 buckets.
package aws_unused_resources

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	cwTypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// UnusedBucket holds detailed info about a low-usage S3 bucket.
type UnusedBucket struct {
	BucketName     string    // S3 bucket name
	CreationDate   time.Time // when the bucket was created
	AvgObjectCount float64   // average number of objects over the period
}

// GetUnusedS3Buckets lists all S3 buckets, computes their average object count over 'days',
// and returns detailed and summary metrics for those below 'threshold'.
func GetUnusedS3Buckets(
	ctx context.Context,
	region string,
	threshold float64,
	days int,
) (UnusedResourceMetrics, error) {
	// Load AWS config
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return UnusedResourceMetrics{}, err
	}
	s3Client := s3.NewFromConfig(cfg)
	cwClient := cloudwatch.NewFromConfig(cfg)

	// List all buckets
	buckets, err := listAllBuckets(ctx, s3Client)
	if err != nil {
		return UnusedResourceMetrics{}, err
	}

	metrics := UnusedResourceMetrics{
		ResourceIDs:          make([]string, 0, len(buckets)),
		TotalInstancesCount:  len(buckets),
		UnusedInstancesCount: 0,
	}
	var unused []UnusedBucket

	for _, b := range buckets {
		avgCount, err := getAvgObjectCount(ctx, cwClient, *b.Name, days)
		if err != nil {
			// skip on error
			continue
		}
		if avgCount < threshold {
			unused = append(unused, UnusedBucket{
				BucketName:     *b.Name,
				CreationDate:   *b.CreationDate,
				AvgObjectCount: avgCount,
			})
			metrics.ResourceIDs = append(metrics.ResourceIDs, *b.Name)
			metrics.UnusedInstancesCount++
		}
	}

	return metrics, nil
}

// listAllBuckets retrieves all S3 buckets in the account.
func listAllBuckets(
	ctx context.Context,
	client *s3.Client,
) ([]s3Types.Bucket, error) {
	resp, err := client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, err
	}
	return resp.Buckets, nil
}

// getAvgObjectCount fetches the average NumberOfObjects metric for a bucket over 'days'.
func getAvgObjectCount(
	ctx context.Context,
	client *cloudwatch.Client,
	bucketName string,
	days int,
) (float64, error) {
	end := time.Now()
	start := end.AddDate(0, 0, -days)
	input := &cloudwatch.GetMetricStatisticsInput{
		Namespace:  aws.String("AWS/S3"),
		MetricName: aws.String("NumberOfObjects"),
		Dimensions: []cwTypes.Dimension{
			{Name: aws.String("BucketName"), Value: aws.String(bucketName)},
			{Name: aws.String("StorageType"), Value: aws.String("AllStorageTypes")},
		},
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
