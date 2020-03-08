db = db.getSiblingDB("coursera");
db.createUser({
  user: "admin",
  pwd: "password",
  roles: [
    {
      role: "readWrite",
      db: "coursera"
    }
  ]
});
