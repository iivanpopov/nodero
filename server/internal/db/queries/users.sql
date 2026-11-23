-- name: GetUserByEmail :one
SELECT id, email, created_at, updated_at
FROM users
WHERE email = ?;

-- name: UserExistsByEmail :one
SELECT EXISTS(SELECT 1 FROM users WHERE email = ?);

-- name: CreateUser :execlastid
INSERT INTO users (email)
VALUES (?);
