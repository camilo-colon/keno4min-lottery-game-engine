#!/bin/bash

# Script to create MongoDB secrets for different environments
# Usage: ./scripts/setup-secrets.sh <environment> <mongodb-uri>
# Example: ./scripts/setup-secrets.sh dev "mongodb+srv://user:pass@cluster.mongodb.net"

set -e

ENVIRONMENT=${1:-dev}
MONGODB_URI=$2

# Validate environment
if [[ ! "$ENVIRONMENT" =~ ^(dev|staging|prod)$ ]]; then
  echo "❌ Error: Environment must be 'dev', 'staging', or 'prod'"
  exit 1
fi

# Validate MongoDB URI
if [ -z "$MONGODB_URI" ]; then
  echo "❌ Error: MongoDB URI is required"
  echo "Usage: $0 <environment> <mongodb-uri>"
  exit 1
fi

SECRET_NAME="/keno4min/${ENVIRONMENT}/mongodb"
SECRET_VALUE=$(cat <<EOF
{
  "uri": "${MONGODB_URI}",
  "database": "keno4min"
}
EOF
)

echo "🔐 Creating secret: ${SECRET_NAME}"

# Check if secret already exists
if aws secretsmanager describe-secret --secret-id "${SECRET_NAME}" 2>/dev/null; then
  echo "⚠️  Secret already exists. Updating..."
  aws secretsmanager update-secret \
    --secret-id "${SECRET_NAME}" \
    --secret-string "${SECRET_VALUE}"
  echo "✅ Secret updated successfully"
else
  echo "📝 Creating new secret..."
  aws secretsmanager create-secret \
    --name "${SECRET_NAME}" \
    --description "MongoDB credentials for Keno4min lottery - ${ENVIRONMENT}" \
    --secret-string "${SECRET_VALUE}" \
    --tags Key=Environment,Value="${ENVIRONMENT}" Key=Project,Value=Keno4min
  echo "✅ Secret created successfully"
fi

echo ""
echo "Secret ARN:"
aws secretsmanager describe-secret --secret-id "${SECRET_NAME}" --query 'ARN' --output text
