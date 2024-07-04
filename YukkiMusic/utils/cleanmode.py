protected_messages = {}


async def protect_message(chat_id, message_id):
    if chat_id not in protected_messages:
        protected_messages[chat_id] = []
    protected_messages[chat_id].append(message_id)


async def send_message(chat_id, text):
    message = await YukkiBot().send_message(chat_id, text)
    await protect_message(chat_id, message.message_id)
