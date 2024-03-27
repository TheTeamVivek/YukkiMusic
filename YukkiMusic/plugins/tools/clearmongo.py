from pyrogram import Client, filters
from pymongo import MongoClient
from YukkiMusic import app
# Replace with your MongoDB URI
mongodb_uri = ""



def clear_mongodb_data(mongodb_uri):
    try:
        # Connect to MongoDB
        client = MongoClient(mongodb_uri)
        
        # Select the database
        db = client.get_default_database()
        
        # Clear all data from the collection
        db.collection_name.delete_many({})
        
        return "All data cleared from MongoDB successfully."
        
    except Exception as e:
        return f"An error occurred: {e}"

@app.on_message(filters.command("clearmongo"))
def clear_mongo_data(client, message):
    if len(message.command) == 2:
        mongodb_uri = message.command[1]
        response = clear_mongodb_data(mongodb_uri)
        message.reply_text(response)
    else:
        message.reply_text("Invalid command usage. Correct format: /clearmongo <mongodb_uri>")
