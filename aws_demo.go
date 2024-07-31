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
		config.WithSharedConfigProfile("limited-admin"),
	)
	if err != nil {
		log.Fatalf("Unable to load AWS SDK config: %v", err)
	}

	fmt.Printf("unused_ebs_volumes %v\n", get_unused_ebs_volumes(cfg))
}

func get_unused_ebs_volumes(cfg aws.Config) []string {
	// Create an EC2 service client.
	svc := ec2.NewFromConfig(cfg)

	// Create a paginator for the DescribeVolumes API call.
	paginator := ec2.NewDescribeVolumesPaginator(svc, &ec2.DescribeVolumesInput{})

	unattached_volumes := []string{}

	// Iterate through the pages of results.
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			log.Fatalf("Failed to retrieve page: %v", err)
		}

		for _, volume := range page.Volumes {

			// volumeJSON, err := json.Marshal(volume)
			// if err != nil {
			// 	log.Fatalf("Failed to marshal volume: %v", err)
			// }

			if len(volume.Attachments) == 0 {
				volumeID := aws.ToString(volume.VolumeId)
				unattached_volumes = append(unattached_volumes, volumeID)
			}
		}
	}

	return unattached_volumes
}
