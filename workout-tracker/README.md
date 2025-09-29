# Workout Tracker API

A comprehensive backend system for a workout tracker application that allows users to sign up, log in, create workout plans, schedule workouts, and track their progress.

## Features

- **User Authentication**: JWT-based authentication with user registration and login
- **Exercise Management**: Predefined exercises categorized by muscle group and type
- **Workout Planning**: Create custom workout plans with multiple exercises
- **Workout Scheduling**: Schedule workouts for specific dates and times
- **Progress Tracking**: Track workout completion and exercise performance
- **Reporting**: Generate comprehensive reports on workout history and personal records

## Database Setup

1. Create a MySQL database named `workout_tracker`
2. Run the schema creation:
   ```sql
   mysql -u root -p workout_tracker < internal/db/schema.sql
   ```
3. The application will automatically seed exercises on first run

## Environment Variables

Create a `.env` file in the project root:

```env
HMAC_SECRET=your-secret-key-here
```

## API Endpoints

### Authentication

#### Register User
- **POST** `/register`
- **Body**: 
  ```json
  {
    "username": "john_doe",
    "password": "secure_password"
  }
  ```

#### Login
- **POST** `/login`
- **Body**: 
  ```json
  {
    "username": "john_doe", 
    "password": "secure_password"
  }
  ```
- **Response**: 
  ```json
  {
    "token": "jwt_token_here"
  }
  ```

### Exercises (Protected Routes)

#### Get All Exercises
- **GET** `/api/exercises`
- **Query Parameters**:
  - `category`: Filter by category (cardio, strength, flexibility, balance, sports)
  - `muscle_group`: Filter by muscle group (chest, back, shoulders, arms, legs, core, full_body, cardio)

#### Get Exercise by ID
- **GET** `/api/exercises/{id}`

### Workout Plans (Protected Routes)

#### Create Workout Plan
- **POST** `/api/workout-plans`
- **Body**:
  ```json
  {
    "name": "Upper Body Strength",
    "description": "Focus on chest, back, and arms",
    "exercises": [
      {
        "exercise_id": 1,
        "sets": 3,
        "reps": 10,
        "weight": 50.0,
        "rest_seconds": 60,
        "order_in_workout": 1,
        "notes": "Focus on form"
      }
    ]
  }
  ```

#### Get User's Workout Plans
- **GET** `/api/workout-plans`
- **Query Parameters**:
  - `active`: Filter active plans only (true/false)

#### Get Specific Workout Plan
- **GET** `/api/workout-plans/{id}`

#### Update Workout Plan
- **PUT** `/api/workout-plans/{id}`
- **Body**:
  ```json
  {
    "name": "Updated Plan Name",
    "description": "Updated description", 
    "is_active": false
  }
  ```

#### Delete Workout Plan
- **DELETE** `/api/workout-plans/{id}`

### Workout Sessions (Protected Routes)

#### Schedule Workout
- **POST** `/api/workout-sessions`
- **Body**:
  ```json
  {
    "workout_plan_id": 1,
    "scheduled_date": "2023-12-25T10:00:00Z",
    "notes": "Morning workout"
  }
  ```

#### Get Workout Sessions
- **GET** `/api/workout-sessions`
- **Query Parameters**:
  - `status`: Filter by status (scheduled, in_progress, completed, cancelled)
  - `limit`: Limit number of results

#### Get Specific Session
- **GET** `/api/workout-sessions/{id}`

#### Start Workout
- **POST** `/api/workout-sessions/{id}/start`

#### Update Workout Session
- **PUT** `/api/workout-sessions/{id}`
- **Body**:
  ```json
  {
    "status": "in_progress",
    "notes": "Feeling strong today",
    "total_duration_minutes": 45,
    "exercises": [
      {
        "id": 1,
        "completed_sets": 3,
        "completed_reps": 10,
        "actual_weight": 55.0,
        "notes": "Increased weight",
        "completed": true
      }
    ]
  }
  ```

#### Complete Workout
- **POST** `/api/workout-sessions/{id}/complete`
- **Body**: Same as update session

#### Get Upcoming Workouts
- **GET** `/api/upcoming-workouts`
- **Query Parameters**:
  - `days`: Number of days ahead to look (default: 7)

### Reports (Protected Routes)

#### Generate Workout Report
- **GET** `/api/reports/workout`
- **Query Parameters**:
  - `start_date`: Start date in YYYY-MM-DD format
  - `end_date`: End date in YYYY-MM-DD format

#### Get Personal Records
- **GET** `/api/reports/personal-records`

#### Get Workout Streaks
- **GET** `/api/reports/streaks`
- **Response**:
  ```json
  {
    "current_streak": 5,
    "longest_streak": 12
  }
  ```

## Authentication

All protected routes require a JWT token in the Authorization header:

```
Authorization: Bearer <jwt_token>
```

## Database Schema

The application uses the following main tables:

- **users**: User accounts
- **exercises**: Predefined exercise database
- **workout_plans**: User-created workout plans
- **workout_exercises**: Exercises within workout plans
- **workout_sessions**: Scheduled workout sessions
- **session_exercises**: Exercise performance tracking

## Running the Application

1. Install dependencies:
   ```bash
   go mod download
   ```

2. Set up the database and environment variables

3. Run the application:
   ```bash
   go run cmd/main.go
   ```

The server will start on port 8080.

## Exercise Categories

The system includes predefined exercises in the following categories:

- **Strength**: Push-ups, Bench Press, Squats, Deadlifts, etc.
- **Cardio**: Running, Cycling, Jump Rope, etc.
- **Flexibility**: Forward Fold, Downward Dog, Stretches, etc.
- **Balance**: Single-leg Stand, Tree Pose, etc.

## Application Flow & API Examples

### Complete User Journey

Here's a step-by-step example of how to use the workout tracker API:

#### Step 1: User Registration & Authentication

```bash
# Register a new user
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "fitness_enthusiast",
    "password": "strongPassword123"
  }'

# Login to get JWT token
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "fitness_enthusiast",
    "password": "strongPassword123"
  }'
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

#### Step 2: Browse Available Exercises

```bash
# Get all strength exercises
curl -X GET "http://localhost:8080/api/exercises?category=strength" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"

# Get chest exercises specifically
curl -X GET "http://localhost:8080/api/exercises?muscle_group=chest" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

**Sample Response:**
```json
{
  "exercises": [
    {
      "id": 1,
      "name": "Push-ups",
      "description": "Body weight exercise targeting chest, shoulders, and triceps",
      "category": "strength",
      "muscle_group": "chest"
    },
    {
      "id": 2,
      "name": "Bench Press",
      "description": "Classic barbell or dumbbell press for chest development",
      "category": "strength",
      "muscle_group": "chest"
    }
  ]
}
```

#### Step 3: Create a Workout Plan

```bash
curl -X POST http://localhost:8080/api/workout-plans \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Push Day Workout",
    "description": "Chest, shoulders, and triceps focused session",
    "exercises": [
      {
        "exercise_id": 1,
        "sets": 3,
        "reps": 15,
        "rest_seconds": 60,
        "order_in_workout": 1,
        "notes": "Start with knee push-ups if needed"
      },
      {
        "exercise_id": 2,
        "sets": 4,
        "reps": 8,
        "weight": 60.0,
        "rest_seconds": 90,
        "order_in_workout": 2,
        "notes": "Focus on controlled movement"
      }
    ]
  }'
```

#### Step 4: Schedule a Workout Session

```bash
curl -X POST http://localhost:8080/api/workout-sessions \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -H "Content-Type: application/json" \
  -d '{
    "workout_plan_id": 1,
    "scheduled_date": "2025-09-30T07:00:00Z",
    "notes": "Morning push day session"
  }'
```

#### Step 5: Start and Track Workout

```bash
# Start the workout session
curl -X POST http://localhost:8080/api/workout-sessions/1/start \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"

# Update progress during workout
curl -X PUT http://localhost:8080/api/workout-sessions/1 \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "in_progress",
    "exercises": [
      {
        "id": 1,
        "completed_sets": 3,
        "completed_reps": 15,
        "notes": "Felt strong today",
        "completed": true
      },
      {
        "id": 2,
        "completed_sets": 4,
        "completed_reps": 8,
        "actual_weight": 65.0,
        "notes": "Increased weight by 5lbs!",
        "completed": true
      }
    ]
  }'
```

#### Step 6: Complete Workout

```bash
curl -X POST http://localhost:8080/api/workout-sessions/1/complete \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -H "Content-Type: application/json" \
  -d '{
    "total_duration_minutes": 45,
    "notes": "Great workout! Felt energized throughout.",
    "exercises": [
      {
        "id": 1,
        "completed_sets": 3,
        "completed_reps": 15,
        "completed": true
      },
      {
        "id": 2,
        "completed_sets": 4,
        "completed_reps": 8,
        "actual_weight": 65.0,
        "completed": true
      }
    ]
  }'
```

#### Step 7: View Progress and Reports

```bash
# Get personal records
curl -X GET http://localhost:8080/api/reports/personal-records \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"

# Get workout report for the past month
curl -X GET "http://localhost:8080/api/reports/workout?start_date=2025-09-01&end_date=2025-09-30" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"

# Check workout streaks
curl -X GET http://localhost:8080/api/reports/streaks \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

### Common Use Cases

#### View Upcoming Workouts
```bash
curl -X GET "http://localhost:8080/api/upcoming-workouts?days=7" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

#### Get All Your Workout Plans
```bash
curl -X GET http://localhost:8080/api/workout-plans \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

#### Filter Recent Completed Workouts
```bash
curl -X GET "http://localhost:8080/api/workout-sessions?status=completed&limit=5" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

### Quick Start Guide

1. **Setup**: Create database, set environment variables, run `go run cmd/main.go`
2. **Register**: POST to `/register` with username/password
3. **Login**: POST to `/login` to get JWT token
4. **Explore**: GET `/api/exercises` to see available exercises
5. **Plan**: POST to `/api/workout-plans` to create your routine
6. **Schedule**: POST to `/api/workout-sessions` to schedule workouts
7. **Track**: Use start/update/complete endpoints during workouts
8. **Analyze**: Use report endpoints to view progress

### Error Handling

All endpoints return appropriate HTTP status codes:
- `200` - Success
- `201` - Created (for POST requests)
- `400` - Bad Request (invalid input)
- `401` - Unauthorized (missing/invalid token)
- `404` - Not Found
- `500` - Internal Server Error

Example error response:
```json
{
  "error": "workout plan not found or access denied"
}
```

This API provides a complete backend solution for workout tracking with comprehensive features for fitness enthusiasts and trainers.