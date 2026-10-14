### The migration commands you will use

```bash
# create an empty pair of files with the next sequence number
migrate create -ext sql -dir db/migrations -seq create_users_and_profiles

# apply all pending migrations, or just the next N
migrate -path db/migrations -database "$DATABASE_URL" up
migrate -path db/migrations -database "$DATABASE_URL" up 1

# undo the last migration, or the last N
migrate -path db/migrations -database "$DATABASE_URL" down 1

# jump to a specific version (goes up or down as needed)
migrate -path db/migrations -database "$DATABASE_URL" goto 3

# show the current version
migrate -path db/migrations -database "$DATABASE_URL" version

# repair command, explained below
migrate -path db/migrations -database "$DATABASE_URL" force 2

# delete everything in the database (local development only)
migrate -path db/migrations -database "$DATABASE_URL" drop
```

`down` with no number is meant for undoing everything. Avoid it outside local development.