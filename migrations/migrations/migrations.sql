CREATE TABLE "Token" (
  "id" SERIAL PRIMARY KEY,
  "TokenID" string,
  "UserEmail" string,
  "created_at" datetime DEFAULT (now()),
  "updated_at" datetime,
  "deleted_at" datetime
);

CREATE TABLE "User" (
  "Name" string,
  "Email" string PRIMARY KEY,
  "created_at" datetime DEFAULT (now()),
  "updated_at" datetime,
  "deleted_at" datetime
);

CREATE TABLE "Channel" (
  "id" SERIAL PRIMARY KEY,
  "Title" string,
  "Name" string,
  "Secret" string,
  "HostPassphrase" string,
  "ViewerPassphrase" string,
  "DTMF" string,
  "created_at" datetime DEFAULT (now()),
  "updated_at" datetime,
  "deleted_at" datetime
);
