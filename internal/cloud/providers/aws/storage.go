package aws

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/spotinst/spotctl/internal/cloud"
	"github.com/spotinst/spotctl/internal/log"
)

func (c *Cloud) CreateBucket(ctx context.Context, bucket string) (*cloud.Bucket, error) {
	log.Debugf("Attempting to create bucket %q", bucket)
	bucket = strings.TrimPrefix(bucket, "s3://")

	b, err := c.GetBucket(ctx, bucket)
	if err == nil {
		return b, nil
	}

	return c.createBucket(ctx, bucket)
}

func (c *Cloud) GetBucket(ctx context.Context, bucket string) (*cloud.Bucket, error) {
	log.Debugf("Attempting to find bucket %q", bucket)
	retval := &cloud.Bucket{Name: bucket}

	// Create a new S3 service client.
	svc := s3.New(c.session)

	// Attempt one GetBucketLocation call the "normal" way (i.e. as the bucket owner).
	response, err := svc.GetBucketLocation(&s3.GetBucketLocationInput{
		Bucket: aws.String(bucket),
	})
	if err != nil { // and fallback to brute-forcing if it fails.
		log.Debugf("Unable to get bucket location from region %q; scanning all regions: %v", c.options.Region, err)
		if response, err = c.findBucketLocation(ctx, bucket); err != nil {
			return nil, err
		}
	}
	if response.LocationConstraint == nil { // US classic does not return a region
		retval.Region = "us-east-1"
	} else {
		retval.Region = *response.LocationConstraint

		if retval.Region == "EU" { // another special case: `EU` means `eu-west-1`
			retval.Region = "eu-west-1"
		}
	}

	log.Debugf("Found bucket in region %q", retval.Region)
	return retval, nil
}

func (c *Cloud) findBucketLocation(ctx context.Context, bucket string) (*s3.GetBucketLocationOutput, error) {
	regions, err := c.DescribeRegions(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to list regions: %v", err)
	}

	log.Debugf("Querying for bucket location for %q", bucket)
	out := make(chan *s3.GetBucketLocationOutput, len(regions))
	req := &s3.GetBucketLocationInput{
		Bucket: aws.String(bucket),
	}

	for _, region := range regions {
		go func(regionName string) {
			log.Debugf("Executing GetBucketLocation in %q", regionName)
			svc := s3.New(c.session, &aws.Config{Region: aws.String(regionName)})
			result, bucketError := svc.GetBucketLocationWithContext(ctx, req)
			if bucketError == nil {
				log.Debugf("GetBucketLocation succeeded in %q", regionName)
				out <- result
			}
		}(region.Name)
	}

	select {
	case bucketLocation := <-out:
		return bucketLocation, nil
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("could not retrieve location for bucket %q", bucket)
	}
}

func (c *Cloud) createBucket(ctx context.Context, bucket string) (*cloud.Bucket, error) {
	// Create a new S3 service client.
	svc := s3.New(c.session)

	// Create the bucket.
	_, err := svc.CreateBucketWithContext(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucket),
		CreateBucketConfiguration: &s3.CreateBucketConfiguration{
			LocationConstraint: aws.String(c.options.Region),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create bucket %q: %v", bucket, err)
	}

	// Wait until bucket is created before finishing.
	log.Debugf("Waiting for bucket to be created...")
	err = svc.WaitUntilBucketExistsWithContext(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		return nil, fmt.Errorf("error occurred while waiting for bucket %q "+
			"to be created: %v", bucket, err)
	}

	// Enable versioning.
	log.Debugf("Updating versioning configuration...")
	_, err = svc.PutBucketVersioningWithContext(ctx, &s3.PutBucketVersioningInput{
		Bucket: aws.String(bucket),
		VersioningConfiguration: &s3.VersioningConfiguration{
			Status: aws.String("Enabled"),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("unable to update versioning configuration "+
			"for bucket %q: %v", bucket, err)
	}

	log.Debugf("Bucket %q successfully created", bucket)
	return &cloud.Bucket{
		Name:   bucket,
		Region: c.options.Region,
	}, nil
}
