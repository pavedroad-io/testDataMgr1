
CREATE TABLE IF NOT EXISTS acme.users (
    UsersUUID UUID DEFAULT uuid_v4()::UUID PRIMARY KEY,
    users JSONB
);

CREATE INDEX IF NOT EXISTS usersIdx ON acme.users USING GIN (users);
