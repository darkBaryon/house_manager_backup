const now = Math.floor(Date.now() / 1000);

const landlordColl = db.getCollection("hs_lld_landlord");
const landlordAuthColl = db.getCollection("hs_lld_auth");

const seedPhone = process.env.PUBLISH_AUTH_SEED_PHONE || "18002584637";
const defaultPassword = "123456";
const defaultPasswordHash = "$2b$10$rHqS5rzmKwFELTBIiQoq4.vxlu/K38MqYnW/.9pz.vFStN0YyGVBi";
const passwordHash = process.env.PUBLISH_AUTH_SEED_PASSWORD_HASH || defaultPasswordHash;

print(`[seed-publish-auth] target db: ${db.getName()}`);
print(`[seed-publish-auth] landlord phone: ${seedPhone}`);

landlordColl.createIndex({ phone: 1 }, { unique: true, name: "phone_1" });
landlordColl.createIndex({ status: 1, updated_at: -1 }, { name: "status_1_updated_at_-1" });
landlordAuthColl.createIndex({ landlord_id: 1 }, { unique: true, name: "landlord_id_1" });
landlordAuthColl.createIndex({ auth_type: 1, status: 1 }, { name: "auth_type_1_status_1" });

function baseFields() {
  return {
    updated_at: now,
    status: 1,
  };
}

function upsertOne(collection, filter, set, setOnInsert = {}) {
  collection.updateOne(
    filter,
    {
      $set: {
        ...set,
        ...baseFields(),
      },
      $setOnInsert: {
        _id: new ObjectId(),
        created_at: now,
        version: 1,
        ...setOnInsert,
      },
    },
    { upsert: true }
  );
  return collection.findOne(filter);
}

const landlord = upsertOne(
  landlordColl,
  { phone: seedPhone },
  {
    phone: seedPhone,
  }
);

const landlordAuth = upsertOne(
  landlordAuthColl,
  { landlord_id: landlord._id },
  {
    landlord_id: landlord._id,
    auth_type: "password",
    password_hash: passwordHash,
    password_updated_at: now,
  }
);

printjson({
  ok: 1,
  database: db.getName(),
  landlord: {
    id: landlord._id,
    phone: landlord.phone,
  },
  auth: {
    id: landlordAuth._id,
    auth_type: landlordAuth.auth_type,
  },
  default_password: passwordHash === defaultPasswordHash ? defaultPassword : "<custom hash>",
});
