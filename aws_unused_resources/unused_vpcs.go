package aws_unused_resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// UnusedVpc holds detailed info about a low-usage VPC.
type UnusedVpc struct {
	VpcID         string // VPC ID
	CidrBlock     string // Primary CIDR block
	IsDefault     bool   // Whether this is the default VPC
	InstanceCount int    // Number of running EC2 instances in this VPC
}

// GetUnusedVPCs lists all VPCs in the specified region, counts running EC2 instances
// in each, and returns those with a count <= threshold along with summary metrics.
func GetUnusedVPCs(
	ctx context.Context,
	region string,
	threshold int,
) (UnusedResourceMetrics, error) {
	// Load AWS config
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return UnusedResourceMetrics{}, err
	}
	ec2Client := ec2.NewFromConfig(cfg)

	// Retrieve all VPCs
	vpcs, err := listAllVPCs(ctx, ec2Client)
	if err != nil {
		return UnusedResourceMetrics{}, err
	}

	// Prepare metrics
	metrics := UnusedResourceMetrics{
		ResourceIDs:          make([]string, 0, len(vpcs)),
		TotalInstancesCount:  len(vpcs),
		UnusedInstancesCount: 0,
	}

	var unused []UnusedVpc
	for _, v := range vpcs {
		count, err := countInstancesInVPC(ctx, ec2Client, *v.VpcId)
		if err != nil {
			// Skip this VPC on error
			continue
		}

		if count <= threshold {
			// Determine default status safely
			isDefault := false
			if v.IsDefault != nil && *v.IsDefault {
				isDefault = true
			}
			// Grab primary CIDR block safely
			cidr := ""
			if len(v.CidrBlockAssociationSet) > 0 && v.CidrBlockAssociationSet[0].CidrBlock != nil {
				cidr = *v.CidrBlockAssociationSet[0].CidrBlock
			}
			unused = append(unused, UnusedVpc{
				VpcID:         *v.VpcId,
				CidrBlock:     cidr,
				IsDefault:     isDefault,
				InstanceCount: count,
			})
			metrics.ResourceIDs = append(metrics.ResourceIDs, *v.VpcId)
			metrics.UnusedInstancesCount++
		}
	}

	return metrics, nil
}

// listAllVPCs returns all VPCs in the AWS account for the given region.
func listAllVPCs(
	ctx context.Context,
	client *ec2.Client,
) ([]ec2Types.Vpc, error) {
	resp, err := client.DescribeVpcs(ctx, &ec2.DescribeVpcsInput{})
	if err != nil {
		return nil, err
	}
	return resp.Vpcs, nil
}

// countInstancesInVPC returns the number of running EC2 instances in the specified VPC.
func countInstancesInVPC(
	ctx context.Context,
	client *ec2.Client,
	vpcId string,
) (int, error) {
	input := &ec2.DescribeInstancesInput{
		Filters: []ec2Types.Filter{
			{Name: aws.String("instance-state-name"), Values: []string{"running"}},
			{Name: aws.String("vpc-id"), Values: []string{vpcId}},
		},
	}
	count := 0
	paginator := ec2.NewDescribeInstancesPaginator(client, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return 0, err
		}
		for _, res := range page.Reservations {
			count += len(res.Instances)
		}
	}
	return count, nil
}
