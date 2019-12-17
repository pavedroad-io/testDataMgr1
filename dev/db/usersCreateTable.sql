
CREATE TABLE IF NOT EXISTS Acme.users (
    UsersUUID UUID DEFAULT uuid_v4()::UUID PRIMARY KEY,
    users JSONB
);

CREATE INDEX IF NOT EXISTS usersIdx ON Acme.users USING GIN (users);
