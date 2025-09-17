#!/bin/bash

set -e

# Colors for output
GREEN='\033[0;32m'
AQUA='\033[0;36m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Variables to track the installation progress
CREATE_DOCKER_COMPOSE=false
CLONE_NEKO_ROOMS=false
CREATE_NGINX_CONFIG=false
CREATE_HTPASSWD=false
OBTAIN_CERTIFICATE=false
CREATE_CRON_JOB=false
CREATE_MANAGE_SCRIPT=false

# Function to roll back the installation on error
rollback() {
  echo -e "${RED}\nAn error occurred. Rolling back partial installation...${NC}"

  # 1. Remove Docker containers/images if the compose file was created
  if [ "$CREATE_DOCKER_COMPOSE" == true ]; then
    echo -e "${YELLOW}Removing Docker containers and images...${NC}"
    cd /opt/neko-rooms || true
    docker-compose down || true

    docker rmi m1k1o/neko-rooms:latest -f || true
    rm -f docker-compose.yml
    cd ~ || true
  fi

  # 2. Remove cloned repository
  if [ "$CLONE_NEKO_ROOMS" == true ]; then
    echo -e "${YELLOW}Removing cloned neko-rooms repo...${NC}"
    rm -rf "/opt/neko-rooms" || true
  fi

  # 3. Remove NGINX configuration
  if [ "$CREATE_NGINX_CONFIG" == true ]; then
    echo -e "${YELLOW}Removing NGINX configuration...${NC}"
    rm -f "/etc/nginx/sites-available/${DOMAIN}-neko-rooms.conf"
    rm -f "/etc/nginx/sites-enabled/${DOMAIN}-neko-rooms.conf"
    systemctl reload nginx || true
  fi

  # 4. Remove .htpasswd file
  if [ "$CREATE_HTPASSWD" == true ]; then
    echo -e "${YELLOW}Removing .htpasswd file...${NC}"
    rm -f "/etc/nginx/.htpasswd"
  fi

  # 5. Remove Certbot certificates and configuration
  if [ "$OBTAIN_CERTIFICATE" == true ]; then
    echo -e "${YELLOW}Removing generated certificates...${NC}"
    certbot delete --cert-name "$DOMAIN" -n || true

    rm -f "/etc/letsencrypt/renewal/${DOMAIN}.conf" || true
  fi

  # 6. Remove the cron job
  if [ "$CREATE_CRON_JOB" == true ]; then
    echo -e "${YELLOW}Removing Certbot cron job...${NC}"
    rm -f "/etc/cron.d/certbot-renew"
    systemctl restart crond || true
  fi

  # 7. Remove the manage_htpasswd script
  if [ "$CREATE_MANAGE_SCRIPT" == true ]; then
    echo -e "${YELLOW}Removing manage_htpasswd.sh...${NC}"
    rm -f "/usr/local/bin/manage_htpasswd.sh"
  fi

  echo -e "${GREEN}================================================${NC}"
  echo -e "${RED}Rollback complete. Exiting.${NC}"
  echo -e "${GREEN}================================================${NC}"
  exit 1
}

# Function to get all server IPs
get_server_ips() {
  local ip_list
  local fallback_ip_service="checkip.amazonaws.com"

  ip_list=$(hostname -I | awk '{for (i=1; i<=NF; i++) if ($i !~ /^172\./ && $i !~ /^10\./ && $i !~ /^192\.168\./) print $i}' | tr '\n' ', ' | sed 's/, $//')

  # If the hostname command fails, use the following methods
  if [ -z "$ip_list" ]; then
    ip_list=$(dig +short myip.opendns.com @resolver1.opendns.com 2>/dev/null)
    if [ $? -ne 0 ]; then
      echo "dig command failed, falling back to $fallback_ip_service" >&2
    fi
  fi

  if [ -z "$ip_list" ]; then
    ip_list=$(curl -s $fallback_ip_service)
  fi

  echo "$ip_list"
}

# Get all server IPs
SERVER_IPS=$(get_server_ips)



# Introduction message
echo -e "${GREEN}==================================================${NC}"
echo -e "${GREEN}Neko Rooms Installer${NC} ${AQUA}by H1ghSyst3m${NC}"
echo -e "This scipt uses nginx as reverse proxy, certbot and docker to install Neko Rooms."
echo -e "${GREEN}==================================================${NC}"
echo -e "${YELLOW}This script will:${NC}"
echo -e "  1. Install Docker, Docker Compose, NGINX, and Certbot."
echo -e "  2. Deploy Neko Rooms using Docker Compose."
echo -e "  3. Configure NGINX as a reverse proxy."
echo -e "  4. Obtain and configure an SSL certificate."
echo -e "  5. Add a cron job for automatic SSL renewal."
echo -e "${GREEN}==================================================${NC}"
echo -e "${YELLOW}Please ensure the following before proceeding:${NC}"
echo -e "  - ${GREEN}OS:${NC}"
echo -e "    - Kernel version 2 or higher."
echo -e "    - Debian 9 or higher."
echo -e "    - Amazon Linux"
echo -e "  - ${GREEN}Hardware:${NC}"
echo -e "    - Memory at least 2GB."
echo -e "    - CPU at least 4 cores."
echo -e "    - Disk at least 8GB."
echo -e "  - ${GREEN}Network:${NC}"
echo -e "    - Public IP: $SERVER_IPS"
echo -e "    - Free TCP ports: 80 and 443."
echo -e "    - Free UDP port range: 59000-59100."
echo -e "    - Domain name pointing to your IP: $SERVER_IPS"
echo -e "  - Run this script as superuser."
echo -e "${GREEN}==================================================${NC}\n"

# Detect shell and root privileges
if [ -n "$BASH_VERSION" ]; then
    if [ "${BASH_VERSINFO[0]}" -lt 4 ]; then
        echo 'This installer requires Bash version 4 or higher.' >&2
        exit 1
    fi
else
    echo 'This installer needs to be run with "bash", not "sh" or another shell.' >&2
    exit 1
fi

# Check if the script is run as root
if [ "$(id -u)" -ne 0 ]; then
    echo -e "${RED}Please run as root or use sudo.${NC}"
    exit 1
fi

# Function to ask the user to continue
ask_to_continue() {
  while true; do
    echo -e "${YELLOW}Do you want to continue? (yes/no)${NC}"
    read -r response
    case "$response" in
      [Yy]* ) break ;;
      [Nn]* ) 
        echo -e "${RED}Installation aborted by user.${NC}"
        exit 1 ;; 
      * ) 
        echo -e "${RED}Invalid input. Please answer yes or no.${NC}" ;;
    esac
  done
}

# Function to detect OS and Kernel
detect_os_and_kernel() {
    # Detect OS
    if [ -f /etc/os-release ]; then
        source /etc/os-release
        if [[ "$ID" == "amzn" ]]; then
            OS_VERSION="$(echo $VERSION_ID | cut -d '.' -f1)"
            if [[ "$OS_VERSION" -lt 2 ]]; then
                echo "Amazon Linux 2 or higher is required to use this installer." >&2
                echo "This version of Amazon Linux is too old and unsupported." >&2
                exit 1
            fi
        else
            echo "This installer is designed for Amazon Linux." >&2
            echo "Current OS: $PRETTY_NAME" >&2
            exit 1
        fi
    else
        echo "Could not determine the operating system." >&2
        exit 1
    fi

    # Detect Kernel
    if [[ "$(uname -r | cut -d "." -f 1)" -lt 4 ]]; then
        echo "The system is running an old kernel, which is incompatible with this installer." >&2
        exit 1
    fi
}

# Function to check and install required packages
install_required_packages() {
  echo -e "${YELLOW}Checking and installing required packages: bind-utils and git...${NC}"

  # Update package list and install required packages
  yum update -y >/dev/null 2>&1
  yum install -y bind-utils git >/dev/null 2>&1

  # Verify installation of `dig` (from bind-utils)
  if ! command -v dig &>/dev/null; then
    echo -e "${RED}Failed to install bind-utils. Please check your network or package manager settings.${NC}"
    exit 1
  fi

  # Verify installation of `git`
  if ! command -v git &>/dev/null; then
    echo -e "${RED}Failed to install git. Please check your network or package manager settings.${NC}"
    exit 1
  fi

  echo -e "${GREEN}Required packages installed successfully.${NC}"
}


# Function to validate user input
validate_input() {
  local prompt="$1"
  local var_name="$2"
  local default="$3"

  while true; do
    echo -e "${YELLOW}${prompt}${NC} ${default:+[Default: $default]}"
    read input
    input="${input:-$default}"
    if [ -z "$input" ]; then
      echo -e "${RED}Input cannot be empty. Please try again.${NC}"
    else
      eval "$var_name=\"$input\""
      break
    fi
  done
}

# Function to validate email
validate_email() {
  local email="$1"
  if [[ "$email" =~ ^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$ ]]; then
    return 0
  else
    return 1
  fi
}

# Function to validate domain
validate_domain() {
  local domain="$1"
  if [[ "$domain" =~ ^(([a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?)\.)+[a-zA-Z]{2,}$ ]]; then
    return 0
  else
    return 1
  fi
}

# Function to validate and get domain
validate_and_get_domain() {
  while true; do
    validate_input "Enter your domain name (e.g., watch.example.com, example.com):" DOMAIN
    if validate_domain "$DOMAIN"; then
      # Get the A record of the domain
      DOMAIN_IP=$(dig +short "$DOMAIN" | tail -n 1)
      
      # Convert SERVER_IPS into a line-separated list for better matching
      SERVER_IPS_LIST=$(echo "$SERVER_IPS" | tr ',' '\n' | tr -d ' ')

      # Check if the domain IP matches any server IP
      if echo "$SERVER_IPS_LIST" | grep -q "^$DOMAIN_IP$"; then
        echo -e "${GREEN}Domain $DOMAIN is correctly pointing to the server's IP: $DOMAIN_IP.${NC}"
        break
      else
        echo -e "${RED}Domain $DOMAIN does not point to any of the server's IPs.${NC}"
        echo -e "${YELLOW}Current IP of $DOMAIN: ${DOMAIN_IP:-Not Found}${NC}"
        echo -e "${YELLOW}Your Server IPs: $SERVER_IPS${NC}"
        echo -e "${YELLOW}Please update the domain's A record to point to one of the server's IPs and try again.${NC}"
        echo -e ""

        # Ask the user if they want to continue with the domain anyway
        while true; do
          echo -e "${YELLOW}Do you want to continue with this domain anyway? (yes/no)${NC}"
          read -r response
          case "$response" in
            [Yy]* ) echo -e "${YELLOW}Continuing with domain $DOMAIN despite the IP mismatch.${NC}"; break 2 ;;
            [Nn]* ) echo -e "${RED}Please update the domain's A record and try again.${NC}"; break ;;
            * ) echo -e "${RED}Invalid input. Please answer yes or no.${NC}" ;;
          esac
        done
      fi
    else
      echo -e "${RED}Invalid domain name. Please enter a valid domain.${NC}"
    fi
  done
}

# Function to validate and get email
validate_and_get_email() {
  while true; do
    validate_input "Enter your email for SSL certificate:" EMAIL
    if validate_email "$EMAIL"; then
      break
    else
      echo -e "${RED}Invalid email address. Please enter a valid email.${NC}"
    fi
  done
}

# Function to confirm passwords
confirm_password() {
  local prompt="$1"
  local var_name="$2"

  while true; do
    echo -e "${YELLOW}${prompt}:${NC}"
    read -s password1
    echo -e "${YELLOW}Confirm password:${NC}"
    read -s password2

    if [ "$password1" != "$password2" ]; then
      echo -e "${RED}Passwords do not match. Please try again.${NC}"
    else
      eval "$var_name=\"$password1\""
      break
    fi
  done
}

# Function to validate and check Room port range
check_ports() {
  local port_range="$1"

  IFS='-' read -r start_port end_port <<<"$port_range"
  if [[ ! "$start_port" =~ ^[0-9]+$ ]] || [[ ! "$end_port" =~ ^[0-9]+$ ]] || ((start_port > end_port)); then
    echo -e "${RED}Invalid port range: $port_range. Please use the format START-END (e.g., 59000-59100).${NC}"
    return 1
  fi

  for ((port = start_port; port <= end_port; port++)); do
    if ss -lnu | grep -q ":$port "; then
      echo -e "${RED}Port $port is already in use. Please choose a different range.${NC}"
      return 1
    fi
  done

  echo -e "${GREEN}Ports $port_range are free.${NC}"
  return 0
}

# Function to validate a single port
validate_port() {
  local port="$1"
  if [[ "$port" =~ ^[0-9]+$ ]] && [ "$port" -ge 1 ] && [ "$port" -le 65535 ]; then
    return 0
  else
    return 1
  fi
}

# Check if a port is available
check_port_availability() {
  local port="$1"
  if ss -ln | grep -q ":$port "; then
    return 1
  else
    return 0
  fi
}

# Function to validate and check Room port range
validate_room_port_range() {
  while true; do
    validate_input "Enter the Room port range [Default: 59000-59100]:" ROOM_PORT_RANGE "59000-59100"
    if check_ports "$ROOM_PORT_RANGE"; then
      break
    else
      echo -e "${YELLOW}Please enter a valid and free port range.${NC}"
    fi
  done
}

# Function to validate and check Docker port
validate_docker_port() {
  while true; do
    validate_input "Enter the Docker port for Neko Rooms [Default: 8080]:" DOCKER_PORT "8080"
    if validate_port "$DOCKER_PORT" && check_port_availability "$DOCKER_PORT"; then
      echo -e "${GREEN}Port $DOCKER_PORT is available.${NC}"
      break
    else
      echo -e "${RED}Port $DOCKER_PORT is either invalid or already in use. Please choose another port.${NC}"
    fi
  done
}

# Summarize Configuration
summary() {
  echo -e "\n${GREEN}Configuration Summary:${NC}"
  echo -e "${YELLOW}Domain:${NC} $DOMAIN"
  echo -e "${YELLOW}Email:${NC} $EMAIL"
  echo -e "${YELLOW}Timezone:${NC} $TIMEZONE"
  echo -e "${YELLOW}Room Port Range:${NC} $ROOM_PORT_RANGE"
  echo -e "${YELLOW}Docker Port:${NC} $DOCKER_PORT"
  echo -e "${YELLOW}Rooms Path Prefix:${NC} $PATH_PREFIX"
  echo -e "${YELLOW}Pointed IP:${NC} $DOMAIN_IP"
  echo -e "\n${GREEN}Please review the above settings carefully.${NC}"
}

ask_to_continue
detect_os_and_kernel
echo ""
install_required_packages
echo ""
echo ""
validate_and_get_domain
echo ""
validate_and_get_email
echo ""
validate_input "Enter admin username for authentication:" ADMIN_USER
echo ""
confirm_password "Enter admin password for authentication" ADMIN_PASSWORD
echo ""
validate_input "Enter your timezone (e.g., Europe/Vienna) [Default: UTC]:" TIMEZONE "UTC"
echo ""
validate_room_port_range
echo ""
validate_docker_port
echo ""
echo -e "${YELLOW}Choose a path prefix for accessing Neko Rooms.${NC}"
echo -e "${YELLOW}Example: If you choose 'room', rooms will be accessible at:${NC} ${GREEN}https://${DOMAIN}/room/<room-id>${NC}"
validate_input "Enter your path prefix (e.g., room, home, browser) [Default: room]:" PATH_PREFIX "room"

summary
echo ""
ask_to_continue

# Function to check if a command exists
command_exists() {
  command -v "$1" >/dev/null 2>&1
}

install_packages() {
  local packages=$@
  echo -e "${YELLOW}Installing required packages: $packages...${NC}"

  # Start installation in the background
  (yum update -y && yum install -y $packages) &

  # Get the process ID and show the spinner
  show_spinner $!
}

# Precheck for Docker
check_docker() {
  if command_exists docker; then
    echo -e "${GREEN}Docker is already installed.${NC}"
  else
    echo -e "${YELLOW}Docker is not installed. Installing Docker...${NC}"
    curl -sSL https://get.docker.com/ | CHANNEL=stable bash
    systemctl enable docker
    systemctl start docker
  fi
}

# Precheck for Docker Compose
check_docker_compose() {
  if command_exists docker-compose; then
    echo -e "${GREEN}Docker Compose is already installed.${NC}"
  else
    echo -e "${YELLOW}Docker Compose is not installed. Installing Docker Compose...${NC}"
    DOCKER_COMPOSE_VERSION="1.29.2"
    curl -L "https://github.com/docker/compose/releases/download/${DOCKER_COMPOSE_VERSION}/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    chmod +x /usr/local/bin/docker-compose
  fi
}

# Precheck for Crontab
check_crontab() {
  if command -v crontab >/dev/null 2>&1; then
    echo -e "${GREEN}Crontab is already installed.${NC}"
  else
    echo -e "${YELLOW}Crontab is not installed. Installing...${NC}"
    yum install -y cronie
    systemctl enable crond
    systemctl start crond
  fi
}

# Precheck for NGINX
check_nginx() {
  if command_exists nginx; then
    echo -e "${GREEN}NGINX is already installed.${NC}"
  else
    echo -e "${YELLOW}NGINX is not installed. Installing NGINX...${NC}"
    yum install -y nginx
    systemctl enable nginx
    systemctl start nginx
  fi
}

# Precheck and Installation for Docker, Docker Compose, crontab and NGINX
echo -e "${AQUA}\n==================================================${NC}"
check_docker
echo -e "${AQUA}\n==================================================${NC}"
check_docker_compose
echo -e "${AQUA}\n==================================================${NC}"
check_crontab
echo -e "${AQUA}\n==================================================${NC}"
check_nginx
echo -e "${AQUA}\n==================================================${NC}"

# Install other dependencies
echo -e "${YELLOW}Installing required packages...${NC}"
yum install -y certbot python3-certbot-nginx httpd-tools
echo -e "${AQUA}\n==================================================${NC}"

# Clone Neko Rooms repository
echo -e "${YELLOW}Cloning Neko Rooms repository...${NC}"
git clone https://github.com/m1k1o/neko-rooms.git /opt/neko-rooms || echo -e "${GREEN}Neko Rooms repository already exists. Skipping...${NC}"
cd /opt/neko-rooms
CLONE_NEKO_ROOMS=true
echo -e "${AQUA}\n==================================================${NC}"

# Modify docker-compose.yml
echo -e "${YELLOW}Configuring Neko Rooms${NC} ${AQUA}docker-compose.yml${NC}${YELLOW}...${NC}"
cat <<EOF > docker-compose.yml
version: "3.5"

networks:
  default:
    attachable: true
    name: "neko-rooms-net"

services:
  neko-rooms:
    image: "m1k1o/neko-rooms:latest"
    restart: "unless-stopped"
    environment:
      - "TZ=${TIMEZONE}"
      - "NEKO_ROOMS_MUX=true"
      - "NEKO_ROOMS_EPR=${ROOM_PORT_RANGE}"
      - "NEKO_ROOMS_NAT1TO1=${DOMAIN_IP}"
      - "NEKO_ROOMS_INSTANCE_URL=https://${DOMAIN}/"
      - "NEKO_ROOMS_INSTANCE_NETWORK=neko-rooms-net"
      - "NEKO_ROOMS_TRAEFIK_ENABLED=false"
      - "NEKO_ROOMS_PATH_PREFIX=/${PATH_PREFIX}/"
      - "NEKO_ROOMS_STORAGE_ENABLED=true"
      - "NEKO_ROOMS_STORAGE_INTERNAL=/data"
      - "NEKO_ROOMS_STORAGE_EXTERNAL=/opt/neko-rooms/data"
    ports:
      - "127.0.0.1:${DOCKER_PORT}:8080"
    volumes:
      - "/var/run/docker.sock:/var/run/docker.sock"
      - "/opt/neko-rooms/data:/data"
EOF
CREATE_DOCKER_COMPOSE=true
echo -e "${GREEN} Neko Rooms configuration updated.${NC}"
echo -e "${AQUA}\n==================================================${NC}"

# Start Neko Rooms
echo -e "${GREEN}Starting Neko Rooms service...${NC}"
docker-compose up -d
echo -e "${AQUA}\n==================================================${NC}"

# Configure NGINX without SSL
echo -e "${YELLOW}Configuring NGINX without SSL...${NC}"
mkdir -p /etc/nginx/conf.d
cat <<EOF > /etc/nginx/conf.d/${DOMAIN}-neko-rooms.conf
server {
    listen 80;
    server_name ${DOMAIN};

    location / {
        proxy_pass http://127.0.0.1:${DOCKER_PORT};
        proxy_set_header Host \$host;
        proxy_http_version 1.1;

        # WebSocket headers
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";

        # Forwarding headers for reverse proxy
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;

        proxy_cache_bypass \$http_upgrade;
        proxy_buffering off;
        proxy_read_timeout 900s;
    }

    # Admin Panel Restricted Access
    location ~ ^/$ {
        auth_basic "Restricted Access";
        auth_basic_user_file /etc/nginx/.htpasswd;
        proxy_pass http://127.0.0.1:${DOCKER_PORT};
    }
}
EOF

# Amazon Linux uses conf.d directory, no need for symlinking
CREATE_NGINX_CONFIG=true
systemctl reload nginx
echo -e "${GREEN}NGINX configuration updated and reloaded.${NC}"
echo -e "${AQUA}\n==================================================${NC}"


# Create a password file for admin access
echo -e "${GREEN}Creating admin authentication...${NC}"
htpasswd -cb /etc/nginx/.htpasswd ${ADMIN_USER} ${ADMIN_PASSWORD}
CREATE_HTPASSWD=true
echo -e "${AQUA}\n==================================================${NC}"

# Obtain SSL certificate and configure NGINX for SSL
echo -e "${GREEN}Obtaining SSL certificate and updating NGINX for SSL...${NC}"
certbot --nginx -d ${DOMAIN} --non-interactive --agree-tos -m ${EMAIL}
OBTAIN_CERTIFICATE=true
systemctl reload nginx
echo -e "${AQUA}\n==================================================${NC}"

# Create a separate crontab file for Certbot renewal
echo -e "${GREEN}Creating a separate crontab file for Certbot auto-renewal...${NC}"
cat <<EOF > /etc/cron.d/certbot-renew
# Cron job for Certbot auto-renewal
0 3 * * 0 root certbot renew --quiet && systemctl reload nginx
EOF
CREATE_CRON_JOB=true

# Set appropriate permissions
chmod 644 /etc/cron.d/certbot-renew

# Reload cron to apply changes
echo -e "${GREEN}Reloading cron daemon to apply the new cron job...${NC}"
systemctl restart crond

# Verify the cron job
echo -e "${GREEN}Cron job file created at /etc/cron.d/certbot-renew:${NC}"
cat /etc/cron.d/certbot-renew
echo -e "${AQUA}\n==================================================${NC}"

# Generate a script to manage admin users
echo -e "${GREEN}Generating script to manage admin users...${NC}"
cat <<'EOF' > /usr/local/bin/manage_htpasswd.sh
#!/bin/bash

HTPASSWD_FILE="/etc/nginx/.htpasswd"

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check if htpasswd exists
if ! command -v htpasswd &>/dev/null; then
    echo -e "${RED}Error: htpasswd command not found. Please install httpd-tools.${NC}"
    exit 1
fi

# Function to display usage
usage() {
    echo -e "${YELLOW}Usage:${NC} $0 {add|remove|list|change-password} [username]"
    echo -e "${YELLOW}Commands:${NC}"
    echo -e "  ${GREEN}add${NC}            Add a new user or update an existing user's password"
    echo -e "  ${GREEN}remove${NC}          Remove a user"
    echo -e "  ${GREEN}list${NC}            List all users"
    echo -e "  ${GREEN}change-password${NC} Change the password of an existing user"
    exit 1
}

# Ensure the .htpasswd file exists
if [ ! -f "$HTPASSWD_FILE" ]; then
    echo -e "${YELLOW}The .htpasswd file does not exist. Creating a new one...${NC}"
    touch "$HTPASSWD_FILE"
fi

# Function to prompt for a password
prompt_password() {
    while true; do
        echo -e "${YELLOW}Enter password:${NC}"
        read -rsp "Password: " password1
        echo
        echo -e "${YELLOW}Confirm password:${NC}"
        read -rsp "Confirm Password: " password2
        echo

        if [ "$password1" != "$password2" ]; then
            echo -e "${RED}Passwords do not match. Please try again.${NC}"
        elif [ -z "$password1" ]; then
            echo -e "${RED}Password cannot be empty. Please try again.${NC}"
        else
            echo "$password1"
            return 0
        fi
    done
}

# Add or update a user
add_user() {
    local username="$1"
    echo -e "${YELLOW}Adding or updating user '$username'.${NC}"
    sudo htpasswd "$HTPASSWD_FILE" "$username"
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}User '$username' has been added or updated successfully.${NC}"
    else
        echo -e "${RED}Failed to add or update user '$username'. Please check your inputs.${NC}"
        exit 1
    fi
}

# Change user password
change_password() {
    local username="$1"
    if ! grep -q "^$username:" "$HTPASSWD_FILE"; then
        echo -e "${RED}Error: User '$username' does not exist.${NC}"
        exit 1
    fi
    echo -e "${YELLOW}Changing password for user '$username'.${NC}"
    sudo htpasswd "$HTPASSWD_FILE" "$username"
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}Password for user '$username' has been updated successfully.${NC}"
    else
        echo -e "${RED}Failed to update password for user '$username'. Please check your inputs.${NC}"
        exit 1
    fi
}

# Remove a user
remove_user() {
    local username="$1"
    if ! grep -q "^$username:" "$HTPASSWD_FILE"; then
        echo -e "${RED}Error: User '$username' does not exist.${NC}"
        exit 1
    fi
    sed -i "/^$username:/d" "$HTPASSWD_FILE"
    echo -e "${GREEN}User '$username' has been removed.${NC}"
}

# List all users
list_users() {
    if [ ! -s "$HTPASSWD_FILE" ]; then
        echo -e "${YELLOW}No users found in the .htpasswd file.${NC}"
    else
        echo -e "${GREEN}Users in the .htpasswd file:${NC}"
        cut -d: -f1 "$HTPASSWD_FILE"
    fi
}


# Main logic
case "$1" in
add)
    if [ -z "$2" ]; then
        echo -e "${RED}Error: Username is required for the 'add' command.${NC}"
        usage
    fi
    add_user "$2"
    ;;
remove)
    if [ -z "$2" ]; then
        echo -e "${RED}Error: Username is required for the 'remove' command.${NC}"
        usage
    fi
    remove_user "$2"
    ;;
list)
    list_users
    ;;
change-password)
    if [ -z "$2" ]; then
        echo -e "${RED}Error: Username is required for the 'change-password' command.${NC}"
        usage
    fi
    change_password "$2"
    ;;
*)
    usage
    ;;
esac
EOF
CREATE_MANAGE_SCRIPT=true

# Make the script executable
chmod +x /usr/local/bin/manage_htpasswd.sh

echo -e "${GREEN}Script generated: /usr/local/bin/manage_htpasswd.sh${NC}"
echo -e "${YELLOW}You can use this script to manage admin panel users:${NC}"
echo -e "${YELLOW}  - Add user: /usr/local/bin/manage_htpasswd.sh add <username>${NC}"
echo -e "${YELLOW}  - Remove user: /usr/local/bin/manage_htpasswd.sh remove <username>${NC}"
echo -e "${YELLOW}  - Change password: /usr/local/bin/manage_htpasswd.sh change-password <username>${NC}"
echo -e "${YELLOW}  - List users: /usr/local/bin/manage_htpasswd.sh list${NC}"
echo -e "${AQUA}\n==================================================${NC}"


# Final message
echo -e "${GREEN}Installation complete! Access your Neko Rooms at https://${DOMAIN}${NC}"