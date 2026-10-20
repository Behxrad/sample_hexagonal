db = db.getSiblingDB("inventory");

db.createUser({
    user: "app_user",
    pwd: "app_password",
    roles: [
        {
            role: "readWrite",
            db: "inventory"
        }
    ]
});

db.createCollection("products");