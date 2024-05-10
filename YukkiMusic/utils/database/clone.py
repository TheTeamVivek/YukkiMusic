from YukkiMusic.core.mongo import mongodb
from random import choice
from config import OWNER_ID

cloneownerdb = mongodb.cloneownerdb
clonebotdb = mongodb.clonebotdb
clonebotnamedb = mongodb.clonebotnamedb


async def save_clonebot_owner(bot_id, user_id):
    await cloneownerdb.insert_one({"bot_id": bot_id, "user_id": user_id})


async def get_clonebot_owner(bot_id):
    result = await cloneownerdb.find_one({"bot_id": bot_id})
    if result:
        return result.get("user_id")
    else:
        return choice(OWNER_ID)
