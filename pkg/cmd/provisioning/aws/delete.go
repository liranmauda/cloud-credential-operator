package aws

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/openshift/cloud-credential-operator/pkg/aws"
)

var (
	// DeleteOpts captures the options that affect deletion
	// of the generated objects.
	DeleteOpts = options{}
)

// deleteOIDCObjectsFromBucket deletes the OIDC objects from the S3 bucket
func deleteOIDCObjectsFromBucket(client aws.Client, bucketName, namePrefix string) error {
	paginator := s3.NewListObjectsV2Paginator(client, &s3.ListObjectsV2Input{
		Bucket: awssdk.String(bucketName),
	})
	for paginator.HasMorePages() {
		objectsMetadata, err := paginator.NextPage(context.Background())
		if err != nil {
			return errors.Wrapf(err, "failed to fetch list of Identity Provider objects in the bucket %s", bucketName)
		}

		for _, objectMetadata := range objectsMetadata.Contents {
			objectTags, err := client.GetObjectTagging(context.Background(), &s3.GetObjectTaggingInput{
				Key:    objectMetadata.Key,
				Bucket: awssdk.String(bucketName),
			})
			if err != nil {
				return errors.Wrapf(err, "failed to fetch tags of Identity Provider object %s in the bucket %s", *objectMetadata.Key, bucketName)
			}

			for _, tag := range objectTags.TagSet {
				if *tag.Key == fmt.Sprintf("%s/%s", ccoctlAWSResourceTagKeyPrefix, namePrefix) {
					_, err := client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
						Key:    objectMetadata.Key,
						Bucket: awssdk.String(bucketName),
					})
					if err != nil {
						return errors.Wrapf(err, "failed to delete Identity Provider object %s in the bucket %s", *objectMetadata.Key, bucketName)
					}
					log.Printf("Identity Provider object %s deleted from the bucket %s", *objectMetadata.Key, bucketName)
					break
				}
			}
		}
	}

	return nil
}

// deleteOIDCBucket deletes the OIDC S3 bucket
func deleteOIDCBucket(client aws.Client, bucketName, namePrefix string) error {
	bucketTags, err := client.GetBucketTagging(context.Background(), &s3.GetBucketTaggingInput{
		Bucket: awssdk.String(bucketName),
	})
	if err != nil {
		return errors.Wrapf(err, "failed to fetch tags of the bucket %s", bucketName)
	}

	for _, tag := range bucketTags.TagSet {
		if *tag.Key == fmt.Sprintf("%s/%s", ccoctlAWSResourceTagKeyPrefix, namePrefix) {
			_, err := client.DeleteBucket(context.Background(), &s3.DeleteBucketInput{
				Bucket: awssdk.String(bucketName),
			})
			if err != nil {
				return errors.Wrapf(err, "failed to delete the Identity Provider bucket %s", bucketName)
			}
			log.Printf("Identity Provider bucket %s deleted", bucketName)
			break
		}
	}
	return nil
}

// deleteCloudFrontOriginAccessIdentity deletes the CloudFront origin access identities if created
func deleteCloudFrontOriginAccessIdentity(client aws.Client, namePrefix string) error {
	paginator := cloudfront.NewListCloudFrontOriginAccessIdentitiesPaginator(client, &cloudfront.ListCloudFrontOriginAccessIdentitiesInput{})
	for paginator.HasMorePages() {
		listCloudFrontOriginAccessIdentitiesOutput, err := paginator.NextPage(context.Background())
		if err != nil {
			return errors.Wrapf(err, "failed to fetch a list of CloudFront origin access identities")
		}
		for _, originAccessIdentity := range listCloudFrontOriginAccessIdentitiesOutput.CloudFrontOriginAccessIdentityList.Items {
			if *originAccessIdentity.Comment == fmt.Sprintf("%s/%s", ccoctlAWSResourceTagKeyPrefix, namePrefix) {
				getCloudFrontOriginAccessIdentityOutput, err := client.GetCloudFrontOriginAccessIdentity(context.Background(), &cloudfront.GetCloudFrontOriginAccessIdentityInput{
					Id: originAccessIdentity.Id,
				})
				if err != nil {
					return errors.Wrapf(err, "failed to get the CloudFront origin access identity with ID %s", *originAccessIdentity.Id)
				}

				_, err = client.DeleteCloudFrontOriginAccessIdentity(context.Background(), &cloudfront.DeleteCloudFrontOriginAccessIdentityInput{
					Id:      originAccessIdentity.Id,
					IfMatch: getCloudFrontOriginAccessIdentityOutput.ETag,
				})
				if err != nil {
					return errors.Wrapf(err, "failed to delete the CloudFront origin access identity with ID %s", *originAccessIdentity.Id)
				}
				log.Printf("CloudFront origin access identity with ID %s deleted", *originAccessIdentity.Id)
			}
		}
	}
	return nil
}

// deleteCloudFrontDistribution deletes the CloudFront distribution if created
func deleteCloudFrontDistribution(client aws.Client, namePrefix string) error {
	paginator := cloudfront.NewListDistributionsPaginator(client, &cloudfront.ListDistributionsInput{})
	for paginator.HasMorePages() {
		ListCloudFrontDistributionsOutput, err := paginator.NextPage(context.Background())
		if err != nil {
			return errors.Wrapf(err, "failed to fetch a list of CloudFront distributions")
		}
		for _, distribution := range ListCloudFrontDistributionsOutput.DistributionList.Items {
			listTagsForCloudFrontResourceOutput, err := client.ListTagsForResource(context.Background(), &cloudfront.ListTagsForResourceInput{
				Resource: distribution.ARN,
			})
			if err != nil {
				return errors.Wrapf(err, "failed to fetch tags for CloudFront distribution with ID %s", *distribution.Id)
			}

			for _, tag := range listTagsForCloudFrontResourceOutput.Tags.Items {
				if *tag.Key == fmt.Sprintf("%s/%s", ccoctlAWSResourceTagKeyPrefix, namePrefix) {
					getCloudFrontDistributionOutput, err := client.GetDistribution(context.Background(), &cloudfront.GetDistributionInput{
						Id: distribution.Id,
					})
					if err != nil {
						return errors.Wrapf(err, "failed to get CloudFront Distribution with ID %s", *distribution.Id)
					}

					getCloudFrontDistributionOutput.Distribution.DistributionConfig.Enabled = awssdk.Bool(false)

					updateCloudFrontDistributionOutput, err := client.UpdateDistribution(context.Background(), &cloudfront.UpdateDistributionInput{
						Id:                 distribution.Id,
						IfMatch:            getCloudFrontDistributionOutput.ETag,
						DistributionConfig: getCloudFrontDistributionOutput.Distribution.DistributionConfig,
					})
					if err != nil {
						return errors.Wrapf(err, "failed to disable CloudFront Distribution with ID %s", *distribution.Id)
					}

					for {
						getCloudFrontDistributionOutput, err := client.GetDistribution(context.Background(), &cloudfront.GetDistributionInput{
							Id: distribution.Id,
						})
						if err != nil {
							return errors.Wrapf(err, "failed to get CloudFront Distribution with ID %s", *distribution.Id)
						}

						if *getCloudFrontDistributionOutput.Distribution.Status == cloudFrontDistributionDeployedStatus {
							log.Printf("CloudFront distribution with ID %s is successfully disabled", *distribution.Id)
							break
						}
						log.Printf("Waiting %s for CloudFront distribution with ID %s to be disabled...", cloudFrontDistributionStatusCheckDelay, *distribution.Id)
						time.Sleep(cloudFrontDistributionStatusCheckDelay)
					}

					_, err = client.DeleteDistribution(context.Background(), &cloudfront.DeleteDistributionInput{
						Id:      distribution.Id,
						IfMatch: updateCloudFrontDistributionOutput.ETag,
					})
					if err != nil {
						return errors.Wrapf(err, "failed to delete CloudFront distribution with ID %s", *distribution.Id)
					}
					log.Printf("CloudFront distribution with ID %s deleted", *distribution.Id)
					break
				}
			}

		}
	}
	return nil
}

// deleteIAMRoles deletes the IAM Roles created by ccoctl
func deleteIAMRoles(client aws.Client, namePrefix string) error {
	paginator := iam.NewListRolesPaginator(client, &iam.ListRolesInput{})
	for paginator.HasMorePages() {
		roleList, err := paginator.NextPage(context.Background())
		if err != nil {
			return errors.Wrapf(err, "failed to fetch a list of IAM roles")
		}

		for _, roleMetadata := range roleList.Roles {
			roleOutput, err := client.GetRole(context.Background(), &iam.GetRoleInput{
				RoleName: roleMetadata.RoleName,
			})
			if err != nil {
				return errors.Wrapf(err, "failed to fetch IAM role %s", *roleMetadata.RoleName)
			}

			for _, tag := range roleOutput.Role.Tags {
				if *tag.Key == fmt.Sprintf("%s/%s", ccoctlAWSResourceTagKeyPrefix, namePrefix) {
					if err := deleteRolePolicies(client, *roleOutput.Role.RoleName); err != nil {
						return errors.Wrapf(err, "failed to delete policies associated with IAM Role %s", *roleOutput.Role.RoleName)
					}

					_, err := client.DeleteRole(context.Background(), &iam.DeleteRoleInput{
						RoleName: roleOutput.Role.RoleName,
					})
					if err != nil {
						return errors.Wrapf(err, "failed to delete IAM Role %s", *roleOutput.Role.RoleName)
					}
					log.Printf("IAM Role %s deleted", *roleOutput.Role.RoleName)
					break
				}
			}
		}
	}

	return nil
}

// deleteRolePolicies deletes the Polices associated with IAM Role created by ccoctl
func deleteRolePolicies(client aws.Client, roleName string) error {
	paginator := iam.NewListRolePoliciesPaginator(client, &iam.ListRolePoliciesInput{
		RoleName: awssdk.String(roleName),
	})
	for paginator.HasMorePages() {
		policies, err := paginator.NextPage(context.Background())
		if err != nil {
			return errors.Wrapf(err, "failed to fetch a list of policies associated with IAM role %s", roleName)
		}

		for _, policyName := range policies.PolicyNames {
			_, err := client.DeleteRolePolicy(context.Background(), &iam.DeleteRolePolicyInput{
				RoleName:   awssdk.String(roleName),
				PolicyName: &policyName,
			})
			if err != nil {
				return errors.Wrapf(err, "failed to delete policy %s associated with IAM Role %s", policyName, roleName)
			}
			log.Printf("Policy %s associated with IAM Role %s deleted", policyName, roleName)
		}
	}

	return nil
}

// deleteIAMIdentityProvider deletes the IAM Identity Provider
func deleteIAMIdentityProvider(client aws.Client, namePrefix string) error {
	oidcProviderList, err := client.ListOpenIDConnectProviders(context.Background(), &iam.ListOpenIDConnectProvidersInput{})
	if err != nil {
		return errors.Wrap(err, "failed to fetch list of Identity Providers")
	}

	for _, provider := range oidcProviderList.OpenIDConnectProviderList {
		ok, err := isExistingIdentifyProvider(client, *provider.Arn, namePrefix)
		if err != nil {
			return errors.Wrapf(err, "failed to check for existing Identity Provider")
		}

		if ok {
			_, err := client.DeleteOpenIDConnectProvider(context.Background(), &iam.DeleteOpenIDConnectProviderInput{
				OpenIDConnectProviderArn: awssdk.String(*provider.Arn),
			})
			if err != nil {
				return errors.Wrapf(err, "failed to delete Identity Provider with ARN %s", *provider.Arn)
			}
			log.Printf("Identity Provider with ARN %s deleted", *provider.Arn)
			break
		}
	}

	return nil
}

func deleteCmd(cmd *cobra.Command, args []string) {
	awsClient, err := newAWSClient(DeleteOpts.Region)
	if err != nil {
		log.Fatal(err)
	}
	bucketName := fmt.Sprintf("%s-oidc", DeleteOpts.Name)

	if err := deleteOIDCObjectsFromBucket(awsClient, bucketName, DeleteOpts.Name); err != nil {
		log.Print(err)
	}

	if err := deleteOIDCBucket(awsClient, bucketName, DeleteOpts.Name); err != nil {
		log.Print(err)
	}

	if err := deleteCloudFrontDistribution(awsClient, DeleteOpts.Name); err != nil {
		log.Print(err)
	}

	if err := deleteCloudFrontOriginAccessIdentity(awsClient, DeleteOpts.Name); err != nil {
		log.Print(err)
	}

	if err := deleteIAMRoles(awsClient, DeleteOpts.Name); err != nil {
		log.Print(err)
	}

	if err := deleteIAMIdentityProvider(awsClient, DeleteOpts.Name); err != nil {
		log.Print(err)
	}
}

// NewDeleteCmd implements the "delete" command for the credentials provisioning
func NewDeleteCmd() *cobra.Command {
	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete credentials objects",
		Long:  "Deleting objects related to cloud credentials",
		Run:   deleteCmd,
	}

	deleteCmd.PersistentFlags().StringVar(&DeleteOpts.Name, "name", "", "User-defined name for all created AWS resources (can be separate from the cluster's infra-id)")
	deleteCmd.MarkPersistentFlagRequired("name")
	deleteCmd.PersistentFlags().StringVar(&DeleteOpts.Region, "region", "", "AWS region where the resources were created")
	deleteCmd.MarkPersistentFlagRequired("region")

	return deleteCmd
}
