const seedCount = Number.parseInt(process.env.SEED_COUNT || "100", 10);
if (!Number.isFinite(seedCount) || seedCount <= 0) {
  throw new Error(`invalid SEED_COUNT: ${process.env.SEED_COUNT}`);
}

const now = Math.floor(Date.now() / 1000);
const cities = ["深圳", "广州", "杭州", "上海"];
const areas = ["南山", "福田", "宝安", "龙华", "罗湖", "龙岗"];
const rentModes = ["whole", "shared"];
const moveInPlans = ["随时入住", "一周内", "两周内", "本月内"];

const userColl = db.getCollection("hs_usr_user");
const authColl = db.getCollection("hs_usr_auth");
const profileColl = db.getCollection("hs_usr_profile_ext");
const existingUsers = userColl.countDocuments();
const startIndex = existingUsers + 1;
const endIndex = existingUsers + seedCount;

print(`[seed] target db: ${db.getName()}`);
print(`[seed] generating ${seedCount} auth users`);
print(`[seed] append range: ${startIndex}..${endIndex}`);

userColl.createIndex({ phone: 1 }, { unique: true, name: "phone_1" });
authColl.createIndex(
  { auth_provider: 1, openid: 1 },
  { unique: true, name: "auth_provider_1_openid_1" }
);
authColl.createIndex(
  { auth_provider: 1, unionid: 1 },
  { unique: true, sparse: true, name: "auth_provider_1_unionid_1" }
);
authColl.createIndex({ user_id: 1, status: 1 }, { name: "user_id_1_status_1" });
profileColl.createIndex({ user_id: 1 }, { unique: true, name: "user_id_1" });

function pick(list, index) {
  return list[index % list.length];
}

function makePhone(index) {
  return `188${String(index).padStart(8, "0")}`;
}

function makeUser(index) {
  const id = new ObjectId();
  return {
    _id: id,
    nickname: `Auth Seed ${index}`,
    avatar: "",
    phone: makePhone(index),
    city: pick(cities, index),
    source_channel: "miniapp",
    last_active_at: now - index * 60,
    created_at: now - index * 120,
    updated_at: now - index * 60,
    status: 1,
    version: 1,
  };
}

function makeAuth(user, index) {
  return {
    user_id: user._id,
    auth_provider: "wechat",
    openid: `wx-openid-seed-${String(index).padStart(4, "0")}`,
    unionid: `wx-unionid-seed-${String(index).padStart(4, "0")}`,
    last_login_at: now - index * 60,
    last_login_ip: `127.0.0.${(index % 200) + 1}`,
    created_at: now - index * 120,
    updated_at: now - index * 60,
    status: 1,
    version: 1,
  };
}

function makeProfile(user, index) {
  const budgetMin = 2000 + (index % 10) * 500;
  const budgetMax = budgetMin + 2000 + (index % 4) * 500;
  return {
    user_id: user._id,
    budget_min: budgetMin,
    budget_max: budgetMax,
    preferred_areas: [pick(areas, index), pick(areas, index + 2)],
    preferred_rent_mode: pick(rentModes, index),
    move_in_plan: pick(moveInPlans, index),
    remark: `seed profile ${index}`,
    created_at: now - index * 120,
    updated_at: now - index * 60,
    status: 1,
    version: 1,
  };
}

const users = [];
const auths = [];
const profiles = [];

for (let i = startIndex; i <= endIndex; i += 1) {
  const user = makeUser(i);
  users.push(user);
  auths.push(makeAuth(user, i));
  profiles.push(makeProfile(user, i));
}

const userResult = userColl.insertMany(users, { ordered: true });
const authResult = authColl.insertMany(auths, { ordered: true });
const profileResult = profileColl.insertMany(profiles, { ordered: true });

printjson({
  ok: 1,
  database: db.getName(),
  inserted: {
    hs_usr_user: Object.keys(userResult.insertedIds).length,
    hs_usr_auth: Object.keys(authResult.insertedIds).length,
    hs_usr_profile_ext: Object.keys(profileResult.insertedIds).length,
  },
  samples: {
    phone: users[0].phone,
    openid: auths[0].openid,
    user_id: users[0]._id,
  },
});
