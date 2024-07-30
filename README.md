# dmarc-monitor

A serverless-first application designed to help organizations monitor their DMARC records and reports.

**Note:** This project serves as a playground for experimenting with various developer frameworks, AWS services, and serverless architectures. It is not intended for production use at this stage.

## Architecture

The application is currently in early development, and the architecture is subject to change. Below is the proposed architecture:

### Primary Services

- **Ingest Service**: Utilizes a series of AWS resources and Lambda functions to receive DMARC reports from email providers, parse them, and store the data in a DynamoDB table.
- **Report Service**: Queries the DynamoDB table to generate overviews for the UI or digestible reports for the IT admin.

### Supporting Services

These services operate behind the scenes and are invisible to end users:

- **Tenant Service**: Manages tenant configuration and provides necessary information to other services.
- **Token Service**: Distributes AWS IAM credentials to tenant-facing services to prevent data leakage between tenants.
- **Notification Service**: Sends email or Slack notifications to the IT admin when specific events occur, such as an increase in DMARC failures.

## Tech Stack

- **Frontend**: Next.js, TailwindCSS
- **Backend**: Go, AWS Lambda, DynamoDB, S3, API Gateway, Cognito
- **Infrastructure as Code**: AWS CDK
- **CI/CD**: GitHub Actions
- **Monitoring & Logging**: AWS CloudWatch, X-Ray

## Design Decisions

- **Serverless-first**: The goal is to experiment with serverless architectures to minimize operational costs.
- **Go**: Chosen for its speed and efficiency, complementing the serverless architecture. JavaScript or Python have longer cold start times and higher memory usage, potentially increasing costs.
- **AWS**: Preferred for its robust services and personal preference.
- **AWS CDK**: Selected for managing application infrastructure due to its suitability over Terraform and its ease of use compared to CloudFormation.
