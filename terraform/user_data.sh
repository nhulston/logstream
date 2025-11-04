#!/bin/bash
set -e

# Log all output to a file for debugging
exec > >(tee /var/log/user-data.log)
exec 2>&1

echo "Starting LogStream EC2 bootstrap..."

# Update system
echo "Updating system packages..."
dnf update -y

# Install Docker
echo "Installing Docker..."
dnf install -y docker

# Start and enable Docker service
systemctl start docker
systemctl enable docker

# Add ec2-user to docker group
usermod -aG docker ec2-user

# Install Docker Compose
echo "Installing Docker Compose..."
DOCKER_COMPOSE_VERSION="2.24.5"
curl -L "https://github.com/docker/compose/releases/download/v$${DOCKER_COMPOSE_VERSION}/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose
ln -sf /usr/local/bin/docker-compose /usr/bin/docker-compose

# Install Git
echo "Installing Git..."
dnf install -y git

# Install helpful utilities
echo "Installing utilities..."
dnf install -y htop vim wget jq

# Clone repository
echo "Cloning repository..."
sudo -u ec2-user git clone ${git_repo_url} /home/ec2-user/logstream

# Ensure ownership is correct
chown -R ec2-user:ec2-user /home/ec2-user/logstream

# Wait for Docker to be fully ready
echo "Waiting for Docker to be ready..."
sleep 15

# Start services
echo "Starting LogStream services..."
cd /home/ec2-user/logstream

if [ ! -f docker-compose.yml ]; then
    echo "ERROR: docker-compose.yml not found!"
    exit 1
fi

sudo -u ec2-user docker-compose up -d
echo "Services started. Check status with: docker-compose ps"

# Create welcome message
cat > /etc/motd <<'MOTD'
╔══════════════════════════════════════════════════════════╗
║                                                          ║
║   Welcome to LogStream - Distributed Log Ingestion      ║
║                                                          ║
║   Quick Start:                                           ║
║   - cd ~/logstream                                       ║
║   - ./manage.sh start    # Start all services            ║
║   - ./manage.sh status   # Check service status          ║
║   - ./manage.sh logs     # View logs                     ║
║                                                          ║
║   Services:                                              ║
║   - gRPC Ingestion: port 50051                           ║
║   - Query API: http://localhost:8080                     ║
║   - Kafka: localhost:9092                                ║
║   - PostgreSQL: localhost:5432                           ║
║                                                          ║
╚══════════════════════════════════════════════════════════╝
MOTD

echo "Bootstrap complete! LogStream EC2 instance is ready."
