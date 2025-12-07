# Backend Environment Variables Setup

## Local Development

1. Copy `.env.example` to `.env`:
   ```bash
   cp .env.example .env
   ```

2. Update `.env` with your actual MongoDB URI and secrets:
   ```env
   MONGO_URI=your_actual_mongodb_uri_here
   PORT=8080
   JWT_SECRET=your_actual_jwt_secret_here
   ```

## Production Deployment

### For Railway, Render, or similar platforms:

Add these environment variables in your deployment dashboard:

- `MONGO_URI` - Your MongoDB Atlas connection string
- `PORT` - `8080`
- `JWT_SECRET` - A secure random string for JWT signing

### For Docker:

Use environment variables in your docker-compose.yml or pass them via command line:

```bash
docker run -e MONGO_URI="your_uri" -e JWT_SECRET="your_secret" -p 8080:8080 your-image
```

## Security Notes

- ⚠️ Never commit `.env` to version control (it's in `.gitignore`)
- ✅ Use different MongoDB databases for development and production
- ✅ Rotate JWT_SECRET regularly
- ✅ Use MongoDB Atlas IP whitelist to restrict access
- ✅ Keep `.env.example` updated as a template (without real values)

## Running the Application

```bash
go run cmd/api/main.go
```

The application will automatically load environment variables from `.env` file.
