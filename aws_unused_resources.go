package main

import (
	"context"
	"fmt"
	"log"
	// "encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	// "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func main() {
	// Load the shared AWS configuration.
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"),
		// replace the "limited-admin" with the profile of your choice or completely remove that when running in a Lambda Environment
		config.WithSharedConfigProfile("limited-admin"),
	)
	if err != nil {
		log.Fatalf("Unable to load AWS SDK config: %v", err)
	}

	unused_ebs_data := get_unused_ebs_volumes(cfg)
	fmt.Printf("Unused EBS Volume IDs %s\nTotal Volume Count %d\nUnused Count: %d", 
		unused_ebs_data.ResourceIDs, 
		unused_ebs_data.TotalInstancesCount,
		unused_ebs_data.UnusedInstancesCount,
	)
}

type UnusedResourceMetrics struct {
	ResourceIDs []string
	TotalInstancesCount int
	UnusedInstancesCount int
}

func get_unused_ebs_volumes(cfg aws.Config) UnusedResourceMetrics {
	// Create an EC2 service client.
	svc := ec2.NewFromConfig(cfg)

	// Create a paginator for the DescribeVolumes API call.
	paginator := ec2.NewDescribeVolumesPaginator(svc, &ec2.DescribeVolumesInput{})

	unused_ebs_volumes := UnusedResourceMetrics{}
	totalEBScount := 0
	unusedEBScount := 0

	// Iterate through the pages of results.
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			log.Fatalf("Failed to retrieve page: %v", err)
		}

		for _, volume := range page.Volumes {
			totalEBScount += 1
			// volumeJSON, err := json.Marshal(volume)
			// if err != nil {
			// 	log.Fatalf("Failed to marshal volume: %v", err)
			// }

			if len(volume.Attachments) == 0 {
				unusedEBScount += 1
				volumeID := aws.ToString(volume.VolumeId)
				unused_ebs_volumes.ResourceIDs = append(unused_ebs_volumes.ResourceIDs, volumeID)
			}
		}
	}
	
	unused_ebs_volumes.TotalInstancesCount = totalEBScount
	unused_ebs_volumes.UnusedInstancesCount = unusedEBScount
	
	return unused_ebs_volumes
}
