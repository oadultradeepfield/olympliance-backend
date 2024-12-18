# Olympliance Backend - CVWO Assignment AY2024/25

![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-316192?style=for-the-badge&logo=postgresql&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-2CA5E0?style=for-the-badge&logo=docker&logoColor=white)
![Google Cloud](https://img.shields.io/badge/Google_Cloud-4285F4?style=for-the-badge&logo=google-cloud&logoColor=white)

## 1. About the Project

### 1.1 Description

**Olympliance (Olympiad + Alliance)** is a web forum created for high school students to discuss Science Olympiad problems as they prepare for higher-level competitions. This project was developed as part of the CVWO Assignment AY2024/25 for the School of Computing, National University of Singapore (NUS). For more details, please visit this [link](https://www.comp.nus.edu.sg/~vwo/).

### 1.2 Tech Stack

- **Programming Language**: Go
- **Library and Framework**: Gin and GORM
- **Database**: PostgreSQL (hosted on Neon)
- **Deployment**: Docker and Google Cloud Run

## 2. Table of Contents

- [1. About the Project](#1-about-the-project)
  - [1.1 Description](#11-description)
  - [1.2 Tech Stack](#12-tech-stack)
- [2. Getting Started](#2-getting-started)
  - [2.1 Installation](#21-installation)
  - [2.2 Local Development](#22-local-development)
- [3. Deployment](#3-deployment)
- [4. API Documentation](#4-api-documentation)
- [5. Acknowledgment](#5-acknowledgment)
- [6. License](#6-license)

## 2. Getting Started

### 2.1 Installation

Start by cloning this repository:

```bash
git clone https://github.com/oadultradeepfield/olympliance-backend.git
cd olympliance-backend
```

Before starting the local development server, ensure that the environment variables are correctly configured. Specifically, set up the `.env` file. You can refer to [`.env.example`](/.env.example) for guidance. Below is an example of the default configuration. During development, you can set `ALLOWED_ORIGINS` to `*` to allow requests from any origin. In production, make sure to set `GO_ENVIRONMENT` to `production`.

```
PORT=8080
DSN=your_postgres_database_connection_string
JWT_SECRET=your_jwt_secret_key
ALLOWED_ORIGINS=https://www.your-client-url.com/
GO_ENVIRONMENT=development
```

The `DSN` variable is the database connection string, which can be obtained from the service you are using for deployment. For Neon, the connection string typically follows this format:

```
postgresql://[role]:[password]@[hostname]/[database]?sslmode=require
```

### 2.2 Local Development

This Go application is containerized using Docker. To simplify the setup, it is recommended to use Docker CLI tools to start the local development server. To build and run the Docker container, use the following command:

```bash
docker compose up --build
```

Alternatively, you can build the Docker image separately and run it later using Docker Desktop:

```bash
docker compose build
```

By default, the app will run on `PORT 8080`. To verify if the server is running correctly, visit [http://localhost:8080/health](http://localhost:8080/health). The server should return:

```json
{ "status": "ok" }
```

## 3. Deployment

To deploy using Google Cloud Run, ensure that the `gcloud` CLI is installed on your local machine and set up for authentication. Refer to the installation guide on this [page](https://cloud.google.com/sdk/docs/install) for detailed instructions.

Once authenticated, you can tag the Docker image with the Artifact Registry URL on Google Cloud by running:

```bash
docker tag [LOCAL_IMAGE_NAME] [ARTIFACT_REGISTRY_IMAGE_URL]
```

Next, push the Docker image to the Artifact Registry using:

```bash
docker push [ARTIFACT_REGISTRY_IMAGE_URL]
```

After the image is pushed, navigate to Google Cloud Run in the Google Cloud Console and deploy the server using the uploaded image.

#### Tips for Staying Within the Free Tier

- Configure the server to start based on incoming requests. Note that this will introduce a _cold start_, which may slow down the initial response time.
- Adjust resource settings, such as reducing the allocated CPU, RAM, and maximum instances, to minimize server workloads. These changes may make the app slower but will help maintain an optimal balance for this application’s needs.

**Note**: If you are using macOS devices with ARM architecture, ensure that you build the Docker image with the `linux/amd64` platform flag to guarantee compatibility with Google Cloud Run. By default, macOS devices build images using the `linux/arm64` platform, which works well for local development but may not run correctly on Google Cloud.

Below is an example of how to configure the [`docker-compose.yaml`](/docker-compose.yaml) file to specify the platform:

```yaml
services:
  app:
    build:
      context: .
      dockerfile: Dockerfile
      platform: linux/amd64 # add this line
    image: olympliance-server:latest
    ports:
      - "8080:8080"
    environment:
      - PORT=${PORT}
      - DSN=${DSN}
      - JWT_SECRET=${JWT_SECRET}
      - ALLOWED_ORIGINS=${ALLOWED_ORIGINS}
      - GO_ENVIRONMENT=${GO_ENVIRONMENT}
    restart: unless-stopped
```

## 4. API Documentation

## 5. Acknowledgment

I would like to express my heartfelt gratitude to my seniors who introduced me to CVWO well in advance. Their guidance allowed me to prepare and familiarize myself with the tech stack throughout the first semester, which significantly streamlined the development process.

I also want to acknowledge the incredible creators who provided invaluable tutorials that guided me through various aspects of this project. Special thanks to [NetNinja](https://www.youtube.com/channel/UCW5YeuERMmlnqo4oq8vwUpg), [Coding with Robby](https://www.youtube.com/@codingwithrobby), [Fireship](https://www.youtube.com/c/Fireship) (particularly for the deployment part), and others that weren't mention. I’m also grateful to the developer communities on discussion platforms like StackOverflow and Reddit, whose insights helped me discover PaaS providers with generous free tiers.

Lastly, I want to thank the School of Computing for offering this program and opportunity. Working on this project has taught me more than I ever anticipated. The challenges of debugging and integrating multiple programming components have reshaped the way I approach problem-solving and coding. Additionally, I’ve grown to appreciate Docker for its incredible ability to make everything so portable and efficient.

## 6. License

This project is licensed under the MIT License. You can view the full license text in the [LICENSE](/LICENSE) file.
