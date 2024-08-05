package main

import (
	"context"
	"fmt"
	"log"
	// "encoding/json"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/sawlemon/unused-cloud-resources/aws_unused_resources"
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

	unused_ebs_data := aws_unused_resources.get_unused_ebs_volumes(cfg)
	fmt.Printf("Unused EBS Volume IDs %s\nTotal Volume Count %d\nUnused Count: %d", 
		unused_ebs_data.ResourceIDs, 
		unused_ebs_data.TotalInstancesCount,
		unused_ebs_data.UnusedInstancesCount,
	)
}