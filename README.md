# Olympliance Backend - CVWO Assignment AY2024/25

![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-316192?style=for-the-badge&logo=postgresql&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-2CA5E0?style=for-the-badge&logo=docker&logoColor=white)
![Google Cloud](https://img.shields.io/badge/Google_Cloud-4285F4?style=for-the-badge&logo=google-cloud&logoColor=white)

## 1. About the Project

### 1.1 Description

**Olympliance (Olympiad + Alliance)** is a web forum created for high school students to discuss Science Olympiad problems as they prepare for higher-level competitions. This project was developed as part of the CVWO Assignment AY2024/25 for the School of Computing, National University of Singapore (NUS). For more details, please visit this [link](https://www.comp.nus.edu.sg/~vwo/). Feel free to also visit the [frontend repo](https://github.com/oadultradeepfield/olympliance-frontend) for the full context.

- **Project Owner**: Phanuphat Srisukhawasu
- **Matriculation Number**: A0311151B

### 1.2 Tech Stack

- **Programming Language**: Go
- **Library and Framework**: Gin and GORM
- **Database**: PostgreSQL (hosted on Neon)
- **Deployment**: Docker and Google Cloud Run

### 1.3 Database Schema

The database schema is designed to facilitate the relationships between users, threads, comments, and interactions. Below is the ER diagram illustrating the structure. The diagram was generated using Database Markup Language (DBML) on [dbdiagram.io](https://dbdiagram.io/d/Olympliance-6759091ee9daa85aca537ff5).

![Database Schema](/figure/schema.png)

### 1.4 Post-CVWO Plan

1. Add push notifications for new comments and followed threads, plus a real-time user moderation dashboard with WebSocket.
2. Make the forum a PWA for offline use and push notifications.
3. Implement lightweight automatic moderation to detect and flag abusive language for review.

## 2. Table of Contents

- [1. About the Project](#1-about-the-project)
  - [1.1 Description](#11-description)
  - [1.2 Tech Stack](#12-tech-stack)
  - [1.3 Database Schema](#13-database-schema)
  - [1.4 Post-CVWO Plan](#14-post-cvwo-plan)
- [2. Table of Contents](#2-table-of-contents)
- [3. Getting Started](#3-getting-started)
  - [3.1 Installation](#31-installation)
  - [3.2 Building and Running the App](#32-building-and-running-the-app)
- [4. Deployment](#4-deployment)
- [5. API Documentation](#5-api-documentation)
  - [5.1 Authentication Endpoints](#51-authentication-endpoints)
  - [5.2 User Endpoints](#52-user-endpoints)
  - [5.3 Thread Endpoints](#53-thread-enpoints)
  - [5.4 Comment Endpoints](#54-comment-endpoints)
  - [5.5 Interaction Endpoints](#55-interaction-endpoints)
  - [Extra: User Reputation Calculator](#extra-user-reputation-calculator)
- [6. Acknowledgment](#6-acknowledgment)
- [7. License](#7-license)

## 3. Getting Started

### 3.1 Installation

Start by cloning this repository:

```bash
git clone https://github.com/oadultradeepfield/olympliance-backend.git
cd olympliance-backend
```

Before starting the local development server, ensure that the environment variables are correctly configured. Specifically, set up the `.env` file. You can refer to [`.env.example`](/.env.example) for guidance. Below is an example of the default configuration. During development, you can set `ALLOWED_ORIGINS` to `*` to allow requests from any origin. In production, make sure to set `GO_ENVIRONMENT` to `production`. The variables with the prefix `GOOGLE` are configured to set up Google OAuth. Make sure you configure them first in the Google Cloud Console.

```
PORT=8080
DSN=your_postgres_database_connection_string
JWT_SECRET=your_jwt_secret_key
ALLOWED_ORIGINS=https://www.your-client-url.com/
GO_ENVIRONMENT=development
GOOGLE_CLIENT_ID=your_google_client_id
GOOGLE_CLIENT_SECRET=your_google_client_secret
GOOGLE_REDIRECT_URL=your_server_url
FRONTEND_REDIRECT_URL=https://www.your-client-url.com/
BACKEND_DOMAIN=localhost
```

The `DSN` variable is the database connection string, which can be obtained from the service you are using for deployment. For Neon, the connection string typically follows this format:

```
postgresql://[role]:[password]@[hostname]/[database]?sslmode=require
```

### 3.2 Building and Running the App

This Go application is containerized using Docker. To simplify the setup, it is recommended to use Docker CLI tools to build the app. Please ensure that [Docker](https://docs.docker.com/desktop/) is installed.

There are two Dockerfiles available:

- **`Dockerfile.dev`**: For local development with live reloading enabled via `air`.
- **`Dockerfile`**: For production, optimized without live reloading.

#### Development Mode (with Air for Live Reload)

To build and run the Docker container in development mode with live reloading, use the following command:

```bash
docker build -f Dockerfile.dev -t olympliance-server:latest .
docker run -v $(pwd):/app --env-file .env -d --name olympliance-server -p 8080:8080 olympliance-server:latest
```

This will mount your current directory to the container and allow `air` to detect code changes and reload the app automatically. The example configuration of `air` is specified in [`.air.toml`](.air.toml).

#### Production Mode

To build and run the Docker container in production mode, use the following command:

```bash
docker build -t olympliance-server:latest .
docker run --env-file .env -d --name olympliance-server -p 8080:8080 olympliance-server:latest
```

By default, the app will run on **PORT 8080**. To verify if the server is running correctly, visit [http://localhost:8080/health](http://localhost:8080/health). The server should return:

```json
{ "status": "ok" }
```

## 4. Deployment

To deploy using Google Cloud Run, ensure that the `gcloud` CLI is installed on your local machine and set up for authentication. Refer to the installation guide on this [page](https://cloud.google.com/sdk/docs/install) for detailed instructions.

Once authenticated, you can tag the Docker image with the Artifact Registry URL on Google Cloud by running:

```bash
docker tag [LOCAL_IMAGE_NAME] [ARTIFACT_REGISTRY_IMAGE_URL]
```

Next, push the Docker image to the Artifact Registry using:

```bash
docker push [ARTIFACT_REGISTRY_IMAGE_URL]
```

After the image is pushed, navigate to Google Cloud Run in the Google Cloud Console and deploy the server using the uploaded image. You may also need to specify environment variables and secrets except for the `PORT 8080`, which is injected by the platform.

#### Tips for Staying Within the Free Tier

- Configure the server to start based on incoming requests. Note that this will introduce a _cold start_, which may slow down the initial response time.
- Adjust resource settings, such as reducing the allocated CPU, RAM, and maximum instances, to minimize server workloads. These changes may make the app slower but will help maintain an optimal balance for this application’s needs.

**Note**: If you are using macOS devices with ARM architecture, ensure that you build the Docker image with the `linux/amd64` platform flag to guarantee compatibility with Google Cloud Run. By default, macOS devices build images using the `linux/arm64` platform, which works well for local development but may not run correctly on Google Cloud.

To build a Docker image for pushing to Artifact Registry, use the following command:

```bash
docker buildx build --platform linux/amd64 -t olympliance-server ./
```

## 5. API Documentation

### 5.1 Authentication Endpoints

This app uses username-based authentication via JWT with an alternative option for Google Signin. Note that there is a JWT refresh token with an expiration period of a week. As a result, the access token is valid for 15 minutes before the app exchanges a new access token using the refresh token. Notably, tokens are stored in the http-only cookies, so the frontend must configure the API to allow credentials so it can use cookies.

| **URL**                             | **Body**                                         | **Meaning**                                                                                                                                                       |
| ----------------------------------- | ------------------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **POST** `/api/register`            | `{ "username": "string", "password": "string" }` | Register a new user. Requires `username` and `password`.                                                                                                          |
| **POST** `/api/login`               | `{ "username": "string", "password": "string" }` | Login a user. Requires `username` and `password`. Returns both a JWT token and a refresh token upon success in the cookie.                                        |
| **POST** `/api/logout`              | None                                             | Logout a user and remove the JWT and the refresh token from the cookies.                                                                                          |
| **GET** `/api/auth/google`          | None                                             | Redirect to the Google OAuth consent screen for secure sign-in.                                                                                                   |
| **GET** `/api/auth/google/callback` | None                                             | Handle the Google OAuth2 callback, exchanges the authorization code for an access token, retrieves user information, and creates a new user if one doesn't exist. |

### 5.2 User Endpoints

These endpoints are used to manage admin and moderator controls, as well as user interactions for tasks such as changing passwords. Note that some API endpoints are protected from banner users, so they have the same accessibility as the unauthenticated users.

| **URL**                                   | **Body**                                                                                   | **Meaning**                                                                   |
| ----------------------------------------- | ------------------------------------------------------------------------------------------ | ----------------------------------------------------------------------------- |
| **GET** `/api/userinfo`                   | None                                                                                       | Get information for a user with ID or username (specify either in the query). |
| **GET** `/api/users`                      | None                                                                                       | Get the current user's information.                                           |
| **GET** `/api/users/delete`               | None                                                                                       | Delete the account of current logged in users.                                |
| **GET** `/api/leaderboard`                | None                                                                                       | Get the top 10 users based on reputation.                                     |
| **GET** `/api/users/get-id/:username`     | None                                                                                       | Get the user ID by the given username.                                        |
| **PUT** `/api/users/change-username`      | `{ "new_username": "string", "confirm_username": "string" }`                               | Change the current user's username.                                           |
| **PUT** `/api/users/change-password`      | `{ "current_password": "string", "new_password": "string", "confirm_password": "string" }` | Change the current user's password.                                           |
| **PUT** `/api/users/:id/toggle-ban`       | None                                                                                       | Toggle the ban status of a user by their user ID.                             |
| **PUT** `/api/users/:id/toggle-moderator` | None                                                                                       | Toggle moderator status for a user by their user ID.                          |

### 5.3 Thread Enpoints

The endpoints below are used to perform CRUD operations on threads. Threads are categorized using predefined categories, with each category having an associated ID for the predefined names.

| **URL**                                                                                                     | **Body**                                                                               | **Meaning**                                                                                     |
| ----------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------- |
| **GET** `/api/threads/:id`                                                                                  | None (optional)                                                                        | Retrieve a specific thread by its ID.                                                           |
| **GET** `/api/threads/category/:category_id`                                                                | None (optional)                                                                        | Retrieve all threads belonging to a specific category.                                          |
| **POST** `/api/threads`                                                                                     | `{ "title": "string", "content": "string", "category_id": "int", "tags": ["string"] }` | Create a new thread.                                                                            |
| **PUT** `/api/threads/:id`                                                                                  | `{ "title": "string", "content": "string", "tags": ["string"] }`                       | Update an existing thread by ID.                                                                |
| **DELETE** `/api/threads/:id`                                                                               | None                                                                                   | Delete an existing thread by ID.                                                                |
| **GET** `/api/followed-threads/:id?is_deleted={is_deleted}&sort_by={field}&page={number}&per_page={number}` | None                                                                                   | Retrieve threads followed by a user, with options for sorting, pagination, and deleted threads. |

### 5.4 Comment Endpoints

Like threads, comments also support CRUD operations. In fact, comments were designed based on threads. When commenting on comments, the `parent_comment_id` is used, whereas this field is empty when commenting directly on threads.

| **URL**                                                                                                   | **Body**                                                                        | **Meaning**                                                                                                      |
| --------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------- |
| **GET** /api/followed-threads/:id?is_deleted={is_deleted}&sort_by={field}&page={number}&per_page={number} | None (optional)                                                                 | Fetch all comments, optionally filtered by `thread_id`, sorted by `sort_by`, paginated by `page` and `per_page`. |
| **POST** `/api/comments`                                                                                  | `{ "thread_id": "number", "parent_comment_id": "number", "content": "string" }` | Create a new comment associated with a thread and an optional parent comment.                                    |
| **PUT** `/api/comments/:id`                                                                               | `{ "content": "string" }`                                                       | Update an existing comment's content (only if the user is the owner or has admin rights).                        |
| **DELETE** `/api/comments/:id`                                                                            | None                                                                            | Delete an existing comment (only if the user is the owner or has admin rights).                                  |

### 5.5 Interaction Endpoints

The interactions supported in this app are upvotes, downvotes, and follows (which are not available for comments). Users can interact by posting if they haven't previously interacted with the same thread or comment. If they have, the interaction will be updated or deleted.

| **URL**                                                                                     | **Body**                                                                          | **Meaning**                                                                                                                              |
| ------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| **GET** `/api/interactions?user_id={user_id}&thread_id={thread_id}&comment_id={comment_id}` | None (optional)                                                                   | Retrieves interactions based on `user_id` and optional `thread_id` or `comment_id`. Either `thread_id` or `comment_id` must be provided. |
| **POST** `/api/interactions`                                                                | `{ "thread_id": "number", "comment_id": "number", "interaction_type": "string" }` | Creates a new interaction of type `upvote`, `downvote`, or `follow` for a thread or comment.                                             |
| **PUT** `/api/interactions/:id`                                                             | `{ "interaction_type": "string" }`                                                | Updates an existing interaction identified by `id` to a new type of interaction.                                                         |

### Extra: User Reputation Calculator

In addition to the API, this app includes a user reputation calculator service that runs when the server starts. The reputation score for each thread or comment a user makes is calculated using the formula: `max(0, upvotes - downvotes) + comments + follows`. The logic for assigning badges and ranks is handled on the frontend.

## 6. Acknowledgment

I would like to express my heartfelt gratitude to my seniors who introduced me to CVWO well in advance. Their guidance allowed me to prepare and familiarize myself with the tech stack throughout the first semester, which significantly streamlined the development process.

I also want to acknowledge the incredible creators who provided invaluable tutorials that guided me through various aspects of this project. Special thanks to [NetNinja](https://www.youtube.com/channel/UCW5YeuERMmlnqo4oq8vwUpg), [Coding with Robby](https://www.youtube.com/@codingwithrobby), [Fireship](https://www.youtube.com/c/Fireship) (particularly for the deployment part), and others that weren't mention. I’m also grateful to the developer communities on discussion platforms like StackOverflow and Reddit, whose insights helped me discover PaaS with generous free tiers.

Lastly, I want to thank the School of Computing for offering this program and opportunity. Working on this project has taught me more than I ever anticipated. The challenges of debugging and integrating multiple programming components have reshaped the way I approach problem-solving and coding. Additionally, I’ve grown to appreciate Docker for its incredible ability to make everything so portable and efficient.

## 7. License

This project is licensed under the MIT License. You can view the full license text in the [`LICENSE`](/LICENSE) file.
