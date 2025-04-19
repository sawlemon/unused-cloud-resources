package main

import (
	"context"
	"fmt"
	"log"

	aws_unused "github.com/sawlemon/unused-cloud-resources/aws_unused_resources"
)

func main() {

	// unused_ebs_data := aws_unused.Get_unused_ebs_volumes("us-east-1")
	// fmt.Printf("\nUnused EBS Volume IDs %s\nTotal Volume Count %d\nUnused Count: %d",
	// 	unused_ebs_data.ResourceIDs,
	// 	unused_ebs_data.TotalInstancesCount,
	// 	unused_ebs_data.UnusedInstancesCount,
	// )

	// unused_ec2_data := aws_unused.GetUnusedEC2Instances(context.Background(), "us-east-1", 5.0, 7)
	// fmt.Printf("\nUnused EC2 Instances IDs %s\nTotal Volume Count %d\nUnused Count: %d",
	// 	unused_ec2_data.ResourceIDs,
	// 	unused_ec2_data.TotalInstancesCount,
	// 	unused_ec2_data.UnusedInstancesCount,
	// )

	// unused_s3_data, err := aws_unused.GetUnusedS3Buckets(context.Background(), "us-east-1", 1, 7)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Printf("\nUnused S3 Buckets %s\nTotal Volume Count %d\nUnused Count: %d",
	// 	unused_s3_data.ResourceIDs,
	// 	unused_s3_data.TotalInstancesCount,
	// 	unused_s3_data.UnusedInstancesCount,
	// )

	unused_vpcs_data, err := aws_unused.GetUnusedVPCs(context.Background(), "us-east-1", 0)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\nUnused S3 Buckets %s\nTotal Volume Count %d\nUnused Count: %d",
		unused_vpcs_data.ResourceIDs,
		unused_vpcs_data.TotalInstancesCount,
		unused_vpcs_data.UnusedInstancesCount,
	)
}
