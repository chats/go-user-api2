db = db.getSiblingDB('user_service');

// Create users collection
db.createCollection('users');

// Create indexes
db.users.createIndex({ "email": 1 }, { unique: true });
db.users.createIndex({ "username": 1 }, { unique: true });
db.users.createIndex({ "status": 1 });

// Insert admin user
db.users.insertOne({
    "_id": UUID(),
    "email": "admin@example.com",
    "username": "admin",
    "password": "$2a$12$tLUB1UBHhUaJmXKDOyJEJuVeZDiEu9wcUuDmO2i6gvYqfM1qg7yLe", // admin123
    "first_name": "Admin",
    "last_name": "User",
    "role": "admin",
    "status": "active",
    "created_at": new Date(),
    "updated_at": new Date()
});

// Insert test user
db.users.insertOne({
    "_id": UUID(),
    "email": "test@example.com",
    "username": "testuser",
    "password": "$2a$12$9/KQPljPTQK4rdR1MgQ8DetkJPg8GXf3wkYbYNdRLGJYxlFTiX.S2", // test123
    "first_name": "Test",
    "last_name": "User",
    "role": "user",
    "status": "active",
    "created_at": new Date(),
    "updated_at": new Date()
});