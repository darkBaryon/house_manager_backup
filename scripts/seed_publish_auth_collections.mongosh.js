const now = Math.floor(Date.now() / 1000);

const staffColl = db.getCollection("hs_adm_staff");
const roleColl = db.getCollection("hs_adm_role");
const permissionColl = db.getCollection("hs_adm_permission");
const staffRoleColl = db.getCollection("hs_adm_staff_role");
const rolePermissionColl = db.getCollection("hs_adm_role_permission");
const loginLogColl = db.getCollection("hs_adm_login_log");

const seedPhone = process.env.PUBLISH_AUTH_SEED_PHONE || "13900000000";
const seedName = process.env.PUBLISH_AUTH_SEED_NAME || "Publish Admin";

print(`[seed-publish-auth] target db: ${db.getName()}`);
print(`[seed-publish-auth] staff phone: ${seedPhone}`);

staffColl.createIndex({ phone: 1 }, { name: "phone_1" });
staffColl.createIndex({ status: 1, updated_at: -1 }, { name: "status_1_updated_at_-1" });
roleColl.createIndex({ role_code: 1 }, { unique: true, name: "role_code_1" });
roleColl.createIndex({ status: 1, updated_at: -1 }, { name: "status_1_updated_at_-1" });
permissionColl.createIndex({ permission_code: 1 }, { unique: true, name: "permission_code_1" });
permissionColl.createIndex({ module: 1, status: 1 }, { name: "module_1_status_1" });
staffRoleColl.createIndex({ staff_id: 1, role_id: 1 }, { unique: true, name: "staff_id_1_role_id_1" });
staffRoleColl.createIndex({ role_id: 1, status: 1 }, { name: "role_id_1_status_1" });
rolePermissionColl.createIndex({ role_id: 1, permission_id: 1 }, { unique: true, name: "role_id_1_permission_id_1" });
rolePermissionColl.createIndex({ permission_id: 1, status: 1 }, { name: "permission_id_1_status_1" });
loginLogColl.createIndex({ staff_id: 1, login_at: -1 }, { name: "staff_id_1_login_at_-1" });
loginLogColl.createIndex({ login_at: -1 }, { name: "login_at_-1" });

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

const superAdminRole = upsertOne(
  roleColl,
  { role_code: "super_admin" },
  {
    role_name: "Super Admin",
    role_code: "super_admin",
    description: "Built-in publish/admin super role",
    is_system: 1,
  }
);

const houseManagePermission = upsertOne(
  permissionColl,
  { permission_code: "house.manage" },
  {
    permission_name: "House Manage",
    permission_code: "house.manage",
    module: "house",
    action: "manage",
    description: "Manage house publish data",
  }
);

const staff = upsertOne(
  staffColl,
  { phone: seedPhone },
  {
    name: seedName,
    phone: seedPhone,
    email: "",
    contact_qr_code: "",
  }
);

upsertOne(
  staffRoleColl,
  { staff_id: staff._id, role_id: superAdminRole._id },
  {
    staff_id: staff._id,
    role_id: superAdminRole._id,
    assigned_at: now,
  }
);

upsertOne(
  rolePermissionColl,
  { role_id: superAdminRole._id, permission_id: houseManagePermission._id },
  {
    role_id: superAdminRole._id,
    permission_id: houseManagePermission._id,
    assigned_at: now,
  }
);

printjson({
  ok: 1,
  database: db.getName(),
  staff: {
    id: staff._id,
    phone: staff.phone,
    name: staff.name,
  },
  role: superAdminRole.role_code,
  permission: houseManagePermission.permission_code,
});
