import {
  BlockPublicAccess,
  Bucket,
  BucketEncryption,
  BucketProps,
  LifecycleRule,
  StorageClass,
} from "aws-cdk-lib/aws-s3";
import { Duration } from "aws-cdk-lib/core";
import { Construct } from "constructs";

const DEFAULT_ENCRYPTION = BucketEncryption.S3_MANAGED;
const DEFAULT_VERSIONED = true;
const DEFAULT_BLOCK_PUBLIC_ACCESS: BlockPublicAccess = {
  blockPublicAcls: true,
  blockPublicPolicy: true,
  ignorePublicAcls: true,
  restrictPublicBuckets: true,
};
const DEFAULT_LIFECYCLE_RULES: LifecycleRule[] = [
  {
    id: "ManageNonCurrentVersions",
    enabled: true,
    noncurrentVersionExpiration: Duration.days(90),
    abortIncompleteMultipartUploadAfter: Duration.days(7),
    noncurrentVersionTransitions: [
      {
        storageClass: StorageClass.INFREQUENT_ACCESS,
        transitionAfter: Duration.days(30),
      },
    ],
    expiredObjectDeleteMarker: true,
  },
];

export interface S3BucketProps extends BucketProps {}

export class S3Bucket extends Bucket {
  constructor(scope: Construct, id: string, props: S3BucketProps) {
    const {
      encryption = DEFAULT_ENCRYPTION,
      versioned = DEFAULT_VERSIONED,
      blockPublicAccess = DEFAULT_BLOCK_PUBLIC_ACCESS,
      lifecycleRules,
      ...otherProps
    } = props;

    const updatedLifecycleRules = [
      ...DEFAULT_LIFECYCLE_RULES,
      ...(lifecycleRules || []),
    ];

    const updatedProps: S3BucketProps = {
      encryption,
      versioned,
      blockPublicAccess,
      lifecycleRules: updatedLifecycleRules,
      ...otherProps,
    };

    super(scope, id, updatedProps);
  }
}
