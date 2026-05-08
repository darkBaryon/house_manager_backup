const now = Math.floor(Date.now() / 1000);

const centralizedColl = db.getCollection("hs_hmd_centralized");
const buildingColl = db.getCollection("hs_hmd_building");
const decentralizedColl = db.getCollection("hs_hmd_decentralized");
const roomTypeColl = db.getCollection("hs_hmd_room_type_centralized");
const roomCentralizedColl = db.getCollection("hs_hmd_room_centralized");
const roomDecentralizedColl = db.getCollection("hs_hmd_room_decentralized");

print(`[seed-hmd] target db: ${db.getName()}`);

centralizedColl.createIndex({ project_code: 1 }, { unique: true, name: "project_code_1" });
centralizedColl.createIndex({ city: 1, status: 1 }, { name: "city_1_status_1" });

buildingColl.createIndex({ project_id: 1, status: 1 }, { name: "project_id_1_status_1" });
buildingColl.createIndex({ building_code: 1 }, { name: "building_code_1" });

decentralizedColl.createIndex(
  { city: 1, district: 1, community_name: 1 },
  { name: "city_1_district_1_community_name_1" }
);
decentralizedColl.createIndex({ status: 1, updated_at: -1 }, { name: "status_1_updated_at_-1" });

roomTypeColl.createIndex({ building_id: 1, status: 1 }, { name: "building_id_1_status_1" });
roomTypeColl.createIndex({ project_id: 1, room_type_name: 1 }, { name: "project_id_1_room_type_name_1" });

roomCentralizedColl.createIndex(
  { building_id: 1, room_no: 1 },
  { unique: true, name: "building_id_1_room_no_1" }
);
roomCentralizedColl.createIndex({ status: 1, updated_at: -1 }, { name: "status_1_updated_at_-1" });

roomDecentralizedColl.createIndex(
  { decentralized_id: 1, room_no: 1 },
  { unique: true, name: "decentralized_id_1_room_no_1" }
);
roomDecentralizedColl.createIndex({ status: 1, updated_at: -1 }, { name: "status_1_updated_at_-1" });

function seedImage(text, tag) {
  return {
    url: `https://dummyimage.com/960x720/e5e7eb/374151&text=${encodeURIComponent(text)}`,
    tag: tag,
  };
}

function baseFields(offset) {
  return {
    created_at: now - offset * 120,
    updated_at: now - offset * 60,
    status: 1,
    version: 1,
  };
}

function upsertOne(collection, filter, doc) {
  collection.updateOne(
    filter,
    {
      $set: {
        ...doc,
        updated_at: doc.updated_at,
        status: doc.status,
      },
      $setOnInsert: {
        _id: doc._id,
        created_at: doc.created_at,
        version: doc.version,
      },
    },
    { upsert: true }
  );
  return collection.findOne(filter);
}

const projectNs = upsertOne(
  centralizedColl,
  { project_code: "CENTRAL-NS-001" },
  {
    _id: new ObjectId(),
    project_name: "泊寓南山科技园",
    project_code: "CENTRAL-NS-001",
    city: "深圳",
    district: "南山",
    address_text: "南山区科技园科苑路 88 号",
    geo: { lng: 113.943, lat: 22.5402 },
    brand_name: "泊寓",
    ...baseFields(1),
  }
);
const projectFt = upsertOne(
  centralizedColl,
  { project_code: "CENTRAL-FT-001" },
  {
    _id: new ObjectId(),
    project_name: "冠寓福田车公庙",
    project_code: "CENTRAL-FT-001",
    city: "深圳",
    district: "福田",
    address_text: "福田区车公庙泰然八路 18 号",
    geo: { lng: 114.0373, lat: 22.5331 },
    brand_name: "冠寓",
    ...baseFields(2),
  }
);

const buildingNs = upsertOne(
  buildingColl,
  { building_code: "NS-A" },
  {
    _id: new ObjectId(),
    project_id: projectNs._id,
    building_name: "A 座",
    building_code: "NS-A",
    floor_total: 28,
    manager_name: "周星河",
    manager_phone: "13800001001",
    photos: ["https://dummyimage.com/960x720/e5e7eb/374151&text=%E5%8D%97%E5%B1%B1-A%E5%BA%A7-%E5%A4%96%E6%99%AF"],
    listing_facilities: ["elevator", "subway", "front_desk", "laundry_room"],
    ...baseFields(3),
  }
);
const buildingFt = upsertOne(
  buildingColl,
  { building_code: "FT-B" },
  {
    _id: new ObjectId(),
    project_id: projectFt._id,
    building_name: "B 座",
    building_code: "FT-B",
    floor_total: 32,
    manager_name: "沈知夏",
    manager_phone: "13800001002",
    photos: ["https://dummyimage.com/960x720/e5e7eb/374151&text=%E7%A6%8F%E7%94%B0-B%E5%BA%A7-%E5%A4%96%E6%99%AF"],
    listing_facilities: ["elevator", "gym", "security_monitoring", "reception_area"],
    ...baseFields(4),
  }
);

const roomTypeNs = upsertOne(
  roomTypeColl,
  { project_id: projectNs._id, room_type_name: "一室一厅" },
  {
    _id: new ObjectId(),
    project_id: projectNs._id,
    building_id: buildingNs._id,
    room_type_name: "一室一厅",
    room_count: 1,
    hall_count: 1,
    bathroom_count: 1,
    kitchen_count: 1,
    area_size: 42,
    orientation: "south",
    decoration_level: "fine",
    payment_cycle: "monthly",
    rent: 5200,
    deposit: 5200,
    service_fee: 300,
    agency_fee_mode: "none",
    agency_fee_value: 0,
    images: [seedImage("南山-一室一厅", "floor_plan_standard")],
    room_facilities: ["smart_lock", "bed", "wardrobe", "fridge", "air_conditioner"],
    ...baseFields(5),
  }
);
const roomTypeFt = upsertOne(
  roomTypeColl,
  { project_id: projectFt._id, room_type_name: "两室一厅" },
  {
    _id: new ObjectId(),
    project_id: projectFt._id,
    building_id: buildingFt._id,
    room_type_name: "两室一厅",
    room_count: 2,
    hall_count: 1,
    bathroom_count: 1,
    kitchen_count: 1,
    area_size: 68,
    orientation: "south_north",
    decoration_level: "fine",
    payment_cycle: "quarterly",
    rent: 7800,
    deposit: 7800,
    service_fee: 500,
    agency_fee_mode: "none",
    agency_fee_value: 0,
    images: [seedImage("福田-两室一厅", "floor_plan_standard")],
    room_facilities: ["smart_lock", "bed", "wardrobe", "fridge", "air_conditioner", "balcony"],
    ...baseFields(6),
  }
);

upsertOne(
  roomCentralizedColl,
  { building_id: buildingNs._id, room_no: "A-1203" },
  {
    _id: new ObjectId(),
    project_id: projectNs._id,
    building_id: buildingNs._id,
    room_type_id: roomTypeNs._id,
    room_no: "A-1203",
    floor_no: 12,
    rent_mode: "whole",
    layout_text: "1室1厅1卫",
    area_size: 42,
    orientation: "south",
    decoration_level: "fine",
    payment_cycle: "monthly",
    rent: 5200,
    deposit: 5200,
    service_fee: 300,
    agency_fee_mode: "none",
    agency_fee_value: 0,
    room_status: 1,
    viewing_time_rule: "anytime",
    start_rent_rule: "long_one_year",
    images: [seedImage("A-1203-卧室", "bedroom")],
    room_facilities: ["smart_lock", "bed", "wardrobe", "fridge", "air_conditioner"],
    listing_facilities: ["elevator", "subway", "front_desk"],
    ...baseFields(7),
  }
);
upsertOne(
  roomCentralizedColl,
  { building_id: buildingFt._id, room_no: "B-1508" },
  {
    _id: new ObjectId(),
    project_id: projectFt._id,
    building_id: buildingFt._id,
    room_type_id: roomTypeFt._id,
    room_no: "B-1508",
    floor_no: 15,
    rent_mode: "whole",
    layout_text: "2室1厅1卫",
    area_size: 68,
    orientation: "south_north",
    decoration_level: "fine",
    payment_cycle: "quarterly",
    rent: 7800,
    deposit: 7800,
    service_fee: 500,
    agency_fee_mode: "none",
    agency_fee_value: 0,
    room_status: 1,
    viewing_time_rule: "workday_night_weekend",
    start_rent_rule: "long_one_year",
    images: [seedImage("B-1508-客厅", "living_room")],
    room_facilities: ["smart_lock", "bed", "wardrobe", "fridge", "air_conditioner", "balcony"],
    listing_facilities: ["elevator", "gym", "reception_area"],
    ...baseFields(8),
  }
);

const decentralizedSeeds = [
  {
    community_name: "前海时代花园",
    city: "深圳",
    district: "南山",
    biz_area: "前海",
    address_text: "南山区前海路 188 号",
    geo: { lng: 113.8922, lat: 22.5241 },
    subway_station: "前湾",
    offset: 11,
  },
  {
    community_name: "车公庙天安公寓",
    city: "深圳",
    district: "福田",
    biz_area: "车公庙",
    address_text: "福田区泰然六路 66 号",
    geo: { lng: 114.0285, lat: 22.5349 },
    subway_station: "车公庙",
    offset: 12,
  },
  {
    community_name: "宝安中心壹方城",
    city: "深圳",
    district: "宝安",
    biz_area: "宝安中心",
    address_text: "宝安区新湖路 99 号",
    geo: { lng: 113.8846, lat: 22.5558 },
    subway_station: "宝安中心",
    offset: 13,
  },
];

decentralizedSeeds.forEach((seed, communityIndex) => {
  const community = upsertOne(
    decentralizedColl,
    { city: seed.city, district: seed.district, community_name: seed.community_name },
    {
      _id: new ObjectId(),
      community_name: seed.community_name,
      city: seed.city,
      district: seed.district,
      biz_area: seed.biz_area,
      address_text: seed.address_text,
      geo: seed.geo,
      subway_station: seed.subway_station,
      ...baseFields(seed.offset),
    }
  );

  for (let i = 1; i <= 4; i += 1) {
    const shared = i % 2 === 0;
    upsertOne(
      roomDecentralizedColl,
      { decentralized_id: community._id, room_no: `${communityIndex + 1}-${100 + i}` },
      {
        _id: new ObjectId(),
        decentralized_id: community._id,
        room_no: `${communityIndex + 1}-${100 + i}`,
        floor_no: 8 + i,
        rent_mode: shared ? "shared" : "whole",
        layout_text: shared ? "1室0厅1卫" : "2室1厅1卫",
        area_size: shared ? 22 + i : 58 + i * 2,
        orientation: shared ? "east" : "south_north",
        decoration_level: "fine",
        payment_cycle: shared ? "monthly" : "quarterly",
        rent: shared ? 2800 + i * 150 : 5600 + i * 200,
        deposit: shared ? 2800 + i * 150 : 5600 + i * 200,
        service_fee: shared ? 200 : 350,
        agency_fee_mode: "none",
        agency_fee_value: 0,
        room_status: 1,
        viewing_time_rule: i % 3 === 0 ? "weekend_only" : "anytime",
        start_rent_rule: shared ? "short_three_month" : "long_half_year",
        images: [seedImage(`${community.community_name}-${i}`, shared ? "bedroom" : "living_room")],
        room_facilities: ["smart_lock", "bed", "wardrobe", "air_conditioner", "washing_machine"],
        listing_facilities: ["subway", "convenience_store", "security_monitoring"],
        ...baseFields(20 + communityIndex * 4 + i),
      }
    );
  }
});

printjson({
  ok: 1,
  database: db.getName(),
  counts: {
    hs_hmd_centralized: centralizedColl.countDocuments(),
    hs_hmd_building: buildingColl.countDocuments(),
    hs_hmd_decentralized: decentralizedColl.countDocuments(),
    hs_hmd_room_type_centralized: roomTypeColl.countDocuments(),
    hs_hmd_room_centralized: roomCentralizedColl.countDocuments(),
    hs_hmd_room_decentralized: roomDecentralizedColl.countDocuments(),
  },
});
